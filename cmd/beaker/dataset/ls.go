package dataset

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	bytefmt "github.com/beaker/fileheap/bytefmt"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/beaker/client/client"
	"github.com/allenai/beaker/config"
)

type listOptions struct {
	dataset string
	prefix  string
	format  string
}

const (
	formatJSON = "json"
)

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
	cmd.Flag("format", "Output format").EnumVar(&o.format, formatJSON)
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

		switch o.format {
		case formatJSON:
			buf, err := json.Marshal(fileInfo{
				Path:    info.Path,
				Size:    info.Size,
				Updated: info.Updated,
			})
			if err != nil {
				return err
			}
			fmt.Println(string(buf))
		default:
			fmt.Printf(
				"%10s  %s  %s\n",
				bytefmt.FormatBytes(info.Size),
				info.Updated.Format(time.RFC3339),
				info.Path,
			)
		}
	}

	switch o.format {
	case formatJSON:
	default:
		fmt.Printf("Total: %d files, %s\n", totalFiles, bytefmt.FormatBytes(totalBytes))
	}

	return nil
}

type fileInfo struct {
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	Updated time.Time `json:"updated"`
}
