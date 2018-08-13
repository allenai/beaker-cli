package alpha

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/allenai/beaker-api/api"
	beaker "github.com/allenai/beaker-api/client"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type tensorboardOptions struct {
	outputPath string
	group      string
}

func newTensorboardCmd(
	parent *kingpin.CmdClause,
	parentOpts *alphaOptions,
	config *config.Config,
) {
	o := &tensorboardOptions{}
	cmd := parent.Command("tensorboard-logs", "Sync tensorboard logs for one or more experiments")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Flag("output", "Directory to which results will download").
		Required().
		Short('o').
		StringVar(&o.outputPath)
	cmd.Arg("group", "Experiment group ").Required().StringVar(&o.group)
}

func (o *tensorboardOptions) run(beaker *beaker.Client) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		quit := make(chan os.Signal, 1)
		defer close(quit)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		cancel()
	}()

	// Set up a channel that will signal exactly once for the initial update.
	once := make(chan bool, 1)
	once <- true

	// TensorBoard's default reload interval is 5 seconds, but Beaker updates
	// files every 3 minutes at time of writing. We want an interval somewhere
	// in the middle so we can remain responsive without spamming the API.
	const syncInterval = 30 * time.Second
	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()

	tracker := newResultTracker(o.outputPath)
	lastModified := make(map[string]time.Time)

	group, err := beaker.Group(ctx, o.group)
	if err != nil {
		return err
	}

	// This loop exits when the process receives an interrupt.
	for {
		select {
		case <-once:
			// Do nothing, but fall through to the loop below.

		case <-ticker.C:
			// Do nothing, but fall through to the loop below.

		case <-ctx.Done():
			return nil
		}

		paths, err := updateTracker(ctx, beaker, tracker, group)
		if err != nil {
			printError("Failed to update result cache; will retry in "+syncInterval.String(), err)
		}

		updatedTasks := 0
		updatedFiles := 0
		seen := map[string]bool{}
		for _, path := range paths {
			dataset, err := beaker.Dataset(ctx, path.DatasetID)
			if err != nil {
				if isCancelError(err) {
					break
				}
				printError("Failed to resolve dataset for "+path.Path, err)
			}

			manifest, err := dataset.Manifest(ctx)
			if err != nil {
				if isCancelError(err) {
					break
				}
				printError("Failed to get file manifest for "+path.Path, err)
			}

			lastUpdatedFiles := updatedFiles
			for _, file := range manifest.Files {
				// Filter down to only log files that may be relevant for TensorBoard.
				if !strings.Contains(file.File, ".tfevents.") {
					continue
				}

				target := filepath.Join(path.Path, file.File)
				seen[target] = true

				// Skip this file if it hasn't changed since the last time we synced.
				if lastModified[target] == file.TimeLastModified {
					continue
				}
				lastModified[target] = file.TimeLastModified

				// TODO: This needs some sort of non-linear retry.
				fileRef := dataset.FileRef(file.File)
				if err := downloadFile(ctx, fileRef, target, int64(file.Size)); err != nil {
					if !isCancelError(err) {
						break
					}
					printError("Failed to download "+target, err)
				}

				updatedFiles++
			}

			if updatedFiles != lastUpdatedFiles {
				updatedTasks++
			}

			for path := range lastModified {
				if !seen[path] {
					delete(lastModified, path)
				}
			}
		}

		now := time.Now().Format("2006-01-02 15:04:05")
		if updatedFiles == 0 {
			fmt.Printf("[%s] No changes detected.\n", now)
		} else {
			fmt.Printf("[%s] Updated %d files in %d tasks.\n", now, updatedFiles, updatedTasks)
		}
	}
}

func updateTracker(
	ctx context.Context,
	beaker *beaker.Client,
	tracker *resultTracker,
	group *beaker.GroupHandle,
) ([]datasetPath, error) {
	experimentIDs, err := group.Experiments(ctx)
	if err != nil {
		return nil, err
	}

	experiments := make([]*api.Experiment, 0, len(experimentIDs))
	for _, id := range experimentIDs {
		experiment, err := beaker.Experiment(ctx, id)
		if err != nil {
			if isCancelError(err) {
				return nil, nil
			}
			return nil, err
		}

		info, err := experiment.Get(ctx)
		if err != nil {
			if isCancelError(err) {
				return nil, nil
			}
			return nil, err
		}
		experiments = append(experiments, info)
	}

	if err := tracker.SetExperiments(experiments); err != nil {
		return nil, err
	}
	return tracker.DatasetPaths(), nil
}

// Determine whether an error stems from context cancelation.
func isCancelError(err error) bool {
	return errors.Cause(err) == context.Canceled
}

func printError(message string, err error) {
	fmt.Printf("%s: %s\n    %v\n", color.RedString("Error"), message, err)
}

// Download a file in partial or in full. This assumes TB logs are incremental,
// so we can just download the newest bytes. Overwritten bytes are ignored.
func downloadFile(
	ctx context.Context,
	fileRef *beaker.FileHandle,
	target string,
	fileSize int64,
) error {
	link, err := fileRef.PresignLink(ctx, false)
	if err != nil {
		return err
	}

	// Make sure the directory exists.
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return errors.WithStack(err)
	}

	f, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, link.URL, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	if size := info.Size(); size > 0 {
		if size == fileSize {
			// Nothing to sync.
			return nil
		}
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", size, fileSize))
	}

	httpClient := http.Client{}
	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return errors.WithStack(err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(f, resp.Body)
	return errors.WithStack(err)
}
