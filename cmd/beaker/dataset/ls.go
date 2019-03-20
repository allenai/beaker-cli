package dataset

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	beaker "github.com/allenai/beaker/client"
	"github.com/allenai/beaker/config"
	bytefmt "github.com/beaker/fileheap/bytefmt"
	fileheap "github.com/beaker/fileheap/client"
)

type listOptions struct {
	dataset string
	prefix  string
	json    bool
}

func newListCmd(
	parent *kingpin.CmdClause,
	parentOpts *datasetOptions,
	config *config.Config,
) {
	o := &listOptions{}
	cmd := parent.Command("ls", "List files in a dataset")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Arg("dataset", "Dataset name or ID.").Required().StringVar(&o.dataset)
	cmd.Arg("prefix", "Path prefix.").StringVar(&o.prefix)
	cmd.Flag("json", "Output a JSON object for each file.").BoolVar(&o.json)
}

func (o *listOptions) run(beaker *beaker.Client) error {
	ctx := context.Background()
	dataset, err := beaker.Dataset(ctx, o.dataset)
	if err != nil {
		return err
	}

	if dataset.Storage == nil {
		return errors.New("dataset ls is only supported for FileHeap datasets")
	}

	var totalFiles, totalBytes int64
	files := dataset.Storage.Files(ctx, o.prefix)
	for {
		info, err := files.Next()
		if err == fileheap.ErrDone {
			break
		}
		if err != nil {
			return err
		}
		totalFiles++
		totalBytes += info.Size

		if o.json {
			buf, err := json.Marshal(fileInfo{
				Path:    info.Path,
				Size:    info.Size,
				Updated: info.Updated.UnixNano(),
			})
			if err != nil {
				return err
			}
			fmt.Println(string(buf))
		} else {
			fmt.Printf(
				"%10s  %s  %s\n",
				bytefmt.FormatBytes(info.Size),
				info.Updated.Format(time.RFC3339),
				info.Path,
			)
		}
	}

	if !o.json {
		fmt.Printf("Total: %d files, %s\n", totalFiles, bytefmt.FormatBytes(totalBytes))
	}

	return nil
}

type fileInfo struct {
	Path    string `json:"path"`
	Size    int64  `json:"bytes"`
	Updated int64  `json:"updated"`
}
