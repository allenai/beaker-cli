package alpha

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/beaker/client/api"
	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	"golang.org/x/xerrors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	yaml "gopkg.in/yaml.v2"

	"github.com/allenai/beaker/cmd/beaker/experiment"
	"github.com/allenai/beaker/config"
)

// TuneOptions wraps options used in parameter tuning.
type TuneOptions struct {
	Count       int
	Org         string
	Group       string
	SearchSpace string
	Template    string
}

func newTuneCmd(
	parent *kingpin.CmdClause,
	parentOpts *alphaOptions,
	config *config.Config,
) {
	opts := &TuneOptions{}

	cmd := parent.Command("tune", "Run several experiments over a parameter search space")
	cmd.Flag("count", "Total number of experiments to run (default 1)").Short('c').Default("1").IntVar(&opts.Count)
	cmd.Flag("group", "Group in which to place experiments").Short('g').StringVar(&opts.Group)
	cmd.Flag("org", "Org that will own the created experiment").Short('o').StringVar(&opts.Org)
	cmd.Flag("search", "Load a search space from a file.").Short('s').Required().StringVar(&opts.SearchSpace)
	cmd.Flag("template", "Load experiment template from a file.").Short('t').Required().StringVar(&opts.Template)

	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}

		if opts.Org == "" {
			opts.Org = config.DefaultOrg
		}

		_, err = Tune(context.TODO(), os.Stdout, beaker, opts)
		return err
	})
}

type model struct {
	Environment map[string]string
	Parameter   map[string]interface{}
}

// Tune creates multiple experiments and returns their IDs
func Tune(
	ctx context.Context,
	w io.Writer,
	beaker *beaker.Client,
	opts *TuneOptions,
) ([]string, error) {
	if w == nil {
		w = ioutil.Discard
	}
	if opts == nil {
		opts = &TuneOptions{}
	}

	if opts.Count < 1 {
		return nil, errors.New("count must be positive")
	}

	r, err := os.Open(opts.Template)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	specTemplate, err := template.New("spec").Parse(string(b))
	if err != nil {
		return nil, err
	}

	r, err = os.Open(opts.SearchSpace)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	paramSpace, err := decodeParameterSpace(r)
	if err != nil {
		return nil, err
	}

	var group interface {
		ID() string
		AddExperiments(ctx context.Context, experiments []string) error
	}

	if opts.Group != "" {
		gr, err := tryGetGroup(ctx, beaker, opts.Group)
		if err != nil {
			return nil, err
		}

		if gr == nil {
			fmt.Fprintf(w, "Creating group %s... ", opts.Group)

			// TODO: Set workspace, author token.
			if gr, err = beaker.CreateGroup(ctx, api.GroupSpec{
				Organization: opts.Org,
				Name:         opts.Group,
			}); err != nil {
				return nil, err
			}

			fmt.Fprintln(w, "done.")
		}

		group = gr
	}

	experiments, err := runParameterSearch(ctx, beaker, specTemplate, paramSpace, opts)
	if err != nil {
		// Don't return here. We still want to add created experiments to the user's group.
		fmt.Printf("Failed to create some experiments: %s\n", color.YellowString(err.Error()))
	}

	fmt.Fprintf(w, "Submitted %d experiments.\n", opts.Count)
	if group != nil && len(experiments) != 0 {
		if err := group.AddExperiments(ctx, experiments); err != nil {
			return experiments, xerrors.Errorf("Failed to add experiments to %q: %w", opts.Group, err)
		}

		url := fmt.Sprintf("%s/gr/%s", beaker.Address(), group.ID())
		fmt.Fprintf(w, "See progress at %s\n", color.BlueString(url))
	} else {
		for _, ex := range experiments {
			fmt.Printf("%s/ex/%s\n", beaker.Address(), ex)
		}
	}

	return experiments, group.AddExperiments(ctx, experiments)
}

// Run a parameter search by injecting sampled values from a parameter space
// into a templated experiment specification, returning the IDs of newly created
// experiments. Some IDs may be returned even on error.
//
// Argumnets are not validated as this function is intended to be called only by
// the 'tune' command above.
func runParameterSearch(
	ctx context.Context,
	beaker *beaker.Client,
	specTemplate *template.Template,
	paramSpace *ParameterSpace,
	opts *TuneOptions,
) ([]string, error) {
	buf := &bytes.Buffer{}

	var experiments []string
	for i := 0; i < opts.Count; i++ {
		params := &model{Parameter: paramSpace.Sample()}

		buf.Reset()
		if err := specTemplate.Execute(buf, params); err != nil {
			return experiments, err
		}

		var spec experiment.ExperimentSpec
		if err := yaml.UnmarshalStrict(buf.Bytes(), &spec); err != nil {
			return experiments, err
		}

		apiSpec, err := spec.ToAPI()
		if err != nil {
			return experiments, err
		}
		apiSpec.Organization = opts.Org

		// TODO: Set a name?
		experiment, err := beaker.CreateExperiment(ctx, apiSpec, "", false)
		if err != nil {
			return experiments, err
		}

		experiments = append(experiments, experiment.ID())
	}

	return experiments, nil
}

func tryGetGroup(
	ctx context.Context,
	beaker *beaker.Client,
	name string,
) (*beaker.GroupHandle, error) {
	gr, err := beaker.Group(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return nil, nil
		}
		return nil, err
	}
	return gr, nil
}
