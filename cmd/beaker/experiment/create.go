package experiment

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	yaml "gopkg.in/yaml.v2"

	beaker "github.com/allenai/beaker/client"
	"github.com/allenai/beaker/config"
)

// CreateOptions wraps options used to create an experiment.
type CreateOptions struct {
	Name  string
	Quiet bool
}

func newCreateCmd(
	parent *kingpin.CmdClause,
	parentOpts *experimentOptions,
	config *config.Config,
) {
	opts := &CreateOptions{}
	expandVars := new(bool)
	specPath := new(string)

	cmd := parent.Command("create", "Create a new experiment")
	cmd.Flag("expand-vars", "Expand occurrences of '$VAR' and '${VAR}' in the experiment spec file from environment variables. Default true.").
		Default("true").
		BoolVar(expandVars)
	cmd.Flag("file", "Load experiment spec from a file.").Short('f').StringVar(specPath)
	cmd.Flag("name", "Assign a name to the experiment").Short('n').StringVar(&opts.Name)
	cmd.Flag("quiet", "Only display created experiment's ID").Short('q').BoolVar(&opts.Quiet)

	cmd.Action(func(c *kingpin.ParseContext) error {
		var specFile io.Reader
		if *specPath == "-" {
			specFile = os.Stdin
		} else {
			var err error
			specFile, err = os.Open(*specPath)
			if err != nil {
				return err
			}
		}

		spec, err := ReadSpec(specFile, *expandVars)
		if err != nil {
			return err
		}

		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}

		_, err = Create(context.TODO(), os.Stdout, beaker, spec, opts)
		return err
	})
}

// Create creates a new experiment and returns its ID.
func Create(
	ctx context.Context,
	w io.Writer,
	beaker *beaker.Client,
	spec *ExperimentSpec,
	opts *CreateOptions,
) (string, error) {
	if w == nil {
		w = ioutil.Discard
	}
	if opts == nil {
		opts = &CreateOptions{}
	}

	// Dataset IDs may be names or IDs. Fix them up now by resolving them in the service.
	// TODO: It would be far cleaner and more efficient to do this implicitly in the create request.
	for i, exp := range spec.Tasks {
		for j, mount := range exp.Spec.Mounts {
			dataset, err := beaker.Dataset(ctx, mount.DatasetID)
			if err != nil {
				return "", err
			}

			ds, err := dataset.Get(ctx)
			if err != nil {
				return "", err
			}
			spec.Tasks[i].Spec.Mounts[j].DatasetID = ds.ID
		}
	}

	apiSpec, err := spec.ToAPI()
	if err != nil {
		return "", err
	}

	experiment, err := beaker.CreateExperiment(ctx, apiSpec, opts.Name)
	if err != nil {
		return "", err
	}

	if opts.Quiet {
		fmt.Fprintln(w, experiment.ID())
	} else {
		url := experimentURL(beaker.Address(), experiment.ID())
		fmt.Fprintf(w, "Experiment %s submitted. See progress at %s\n", color.BlueString(experiment.ID()), url)
		if apiSpec.EnableComet {
			// get the Experiment from Beaker to show the Comet URL(s)
			createdExp, err := experiment.Get(ctx)
			if err != nil {
				// TODO: Return error instead? But the return value is still good...
				errorMsg := fmt.Sprintf("error getting additional experiment details: %s", err.Error())
				fmt.Fprintf(w, "%s\n", color.RedString(errorMsg))
				return experiment.ID(), nil
			}
			// TODO: This supposes nothing went wrong creating everything Comet-side.
			fmt.Fprintf(w, "Comet.ML experiments were created for each task in this experiment.\n")
			// Arbitrary cutoff so Beaker doesn't spam the user's console on very large experiments.
			if len(createdExp.Nodes) < 20 {
				for _, node := range createdExp.Nodes {
					fmt.Fprintf(w, "%s: %s\n", color.BlueString(node.TaskID), node.CometURL)
				}
			} else {
				fmt.Fprintf(w, "View the Experiment page in your browser for Comet.ML links.\n")
			}
		}
	}

	return experiment.ID(), nil
}

// ReadSpec reads an experiment spec from YAML.
func ReadSpec(r io.Reader, expandVars bool) (*ExperimentSpec, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if expandVars {
		b = []byte(os.ExpandEnv(string(b)))
	}

	var spec ExperimentSpec
	if err := yaml.UnmarshalStrict(b, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}
