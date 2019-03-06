package dataset

import (
	"context"
	"io"
	"os"

	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	beaker "github.com/allenai/beaker/client"
	"github.com/allenai/beaker/config"
)

type streamFileOptions struct {
	dataset string
	file    string
	offset  int64
	length  int64
}

func newStreamCmd(
	parent *kingpin.CmdClause,
	parentOpts *datasetOptions,
	config *config.Config,
) {
	o := &streamFileOptions{}
	cmd := parent.Command("stream-file", "Stream a single file from an existing dataset to stdout")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Arg("dataset", "Dataset name or ID").Required().StringVar(&o.dataset)
	cmd.Arg("file", "File in dataset to fetch. Optional for single-file datasets.").StringVar(&o.file)
	cmd.Flag("offset", "Offset in bytes.").Int64Var(&o.offset)
	cmd.Flag("length", "Number of bytes to read.").Int64Var(&o.length)
}

func (o *streamFileOptions) run(beaker *beaker.Client) error {
	ctx := context.TODO()
	dataset, err := beaker.Dataset(ctx, o.dataset)
	if err != nil {
		return err
	}

	// Single file dataset.
	filename := o.file
	if filename == "" {
		if dataset.Filename() == "" {
			return errors.Errorf("filename required for multi-file dataset %s", dataset.ID())
		}
		filename = dataset.Filename()
	}
	fileRef := dataset.FileRef(filename)

	var r io.ReadCloser
	if o.offset != 0 || o.length != 0 {
		if o.length == 0 {
			// Length not specified; read the rest of the file.
			r, err = fileRef.DownloadRange(ctx, o.offset, -1)
		} else {
			r, err = fileRef.DownloadRange(ctx, o.offset, o.length)
		}
	} else {
		r, err = fileRef.Download(ctx)
	}
	if err != nil {
		return err
	}
	defer r.Close()

	_, err = io.Copy(os.Stdout, r)
	return errors.WithStack(err)
}
