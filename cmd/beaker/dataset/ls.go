package dataset

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	bytefmt "github.com/beaker/fileheap/bytefmt"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/client"
	"github.com/allenai/beaker/config"
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
	cmd := parent.Command("ls", "List files in a dataset.")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := client.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Arg("dataset", "Dataset name or ID").Required().StringVar(&o.dataset)
	cmd.Arg("prefix", "Path prefix").StringVar(&o.prefix)
	cmd.Flag("json", "Output a JSON object for each file.").BoolVar(&o.json)
}

func (o *listOptions) run(beaker *client.Client) error {
	ctx := context.Background()
	dataset, err := beaker.Dataset(ctx, o.dataset)
	if err != nil {
		return err
	}

	var totalFiles, totalBytes int64
	files, err := dataset.Files(ctx, o.prefix)
	if err != nil {
		return err
	}
	for {
		_, info, err := files.Next()
		if err == client.ErrDone {
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
