package experiment

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	beaker "github.com/beaker/client/client"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

// ResumeOptions wraps options used to create an experiment.
type ResumeOptions struct {
	Name string
}

func newResumeCmd(
	parent *kingpin.CmdClause,
	parentOpts *experimentOptions,
	config *config.Config,
) {
	var experimentToResume string

	opts := &ResumeOptions{}
	cmd := parent.Command("resume", "Resume a preempted experiment and return the experiment ID for the new experiment.")
	cmd.Flag("experiment-name", "Experiment to resume (name or experiment ID).").Short('e').Required().StringVar(&experimentToResume)
	cmd.Flag("name", "Assign a name to the experiment").Short('n').StringVar(&opts.Name)

	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}

		_, err = Resume(context.TODO(), os.Stdout, beaker, experimentToResume, opts)
		return err
	})
}

// Resume creates a new experiment based on the state of the previously preempted experiment in experimentToResume
// and returns its ID.
func Resume(
	ctx context.Context,
	w io.Writer,
	beaker *beaker.Client,
	experimentToResume string,
	opts *ResumeOptions,
) (string, error) {
	if w == nil {
		w = ioutil.Discard
	}
	if opts == nil {
		opts = &ResumeOptions{}
	}

	experiment, err := beaker.ResumeExperiment(ctx, experimentToResume, opts.Name)
	if err != nil {
		return "", err
	}

	fmt.Fprintln(w, experiment.ID())

	return experiment.ID(), nil
}
