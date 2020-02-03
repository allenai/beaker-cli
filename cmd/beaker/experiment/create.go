package experiment

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	yaml "gopkg.in/yaml.v2"

	configCmd "github.com/allenai/beaker/cmd/beaker/config"
	"github.com/allenai/beaker/config"
)

// CreateOptions wraps options used to create an experiment.
type CreateOptions struct {
	Name      string
	Quiet     bool
	Force     bool
	Workspace string
}

func newCreateCmd(
	parent *kingpin.CmdClause,
	parentOpts *experimentOptions,
	cfg *config.Config,
) {
	opts := &CreateOptions{}
	specPath := new(string)

	cmd := parent.Command("create", "Create a new experiment")
	cmd.Flag("file", "Load experiment spec from a file.").Short('f').StringVar(specPath)
	cmd.Flag("name", "Assign a name to the experiment").Short('n').StringVar(&opts.Name)
	cmd.Flag("quiet", "Only display created experiment's ID").Short('q').BoolVar(&opts.Quiet)
	cmd.Flag("workspace", "Workspace where the experiment will be placed").Short('w').StringVar(&opts.Workspace)
	cmd.Flag("force", "Allow depending on uncommitted datasets").BoolVar(&opts.Force)

	cmd.Action(func(c *kingpin.ParseContext) error {
		specFile, err := openPath(*specPath)
		if err != nil {
			return err
		}

		spec, err := ReadSpec(specFile)
		if err != nil {
			return err
		}

		beaker, err := beaker.NewClient(parentOpts.addr, cfg.UserToken)
		if err != nil {
			return err
		}

		if opts.Workspace == "" {
			opts.Workspace, err = configCmd.EnsureDefaultWorkspace(beaker, cfg)
			if err != nil {
				return err
			}
			if !opts.Quiet {
				fmt.Printf("Using workspace %s\n", color.BlueString(opts.Workspace))
			}
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
	apiSpec.Workspace = opts.Workspace

	experiment, err := beaker.CreateExperiment(ctx, apiSpec, opts.Name, opts.Force)
	if err != nil {
		return "", err
	}

	if opts.Quiet {
		fmt.Fprintln(w, experiment.ID())
	} else {
		url := experimentURL(beaker.Address(), experiment.ID())
		fmt.Fprintf(w, "Experiment %s submitted. See progress at %s\n", color.BlueString(experiment.ID()), url)
	}

	return experiment.ID(), nil
}

type templateParams struct {
	Environment map[string]string
}

// ReadSpec reads an experiment spec from YAML.
func ReadSpec(r io.Reader) (*ExperimentSpec, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	specTemplate, err := template.New("spec").Parse(string(b))
	if err != nil {
		return nil, err
	}

	envVars := map[string]string{}
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		envVars[parts[0]] = parts[1]
	}

	buf := &bytes.Buffer{}
	if err := specTemplate.Execute(buf, templateParams{Environment: envVars}); err != nil {
		return nil, err
	}

	var spec ExperimentSpec
	if err := yaml.UnmarshalStrict(buf.Bytes(), &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}

func openPath(p string) (io.Reader, error) {
	// Special case: "-" means read from STDIN.
	if p == "-" {
		return os.Stdin, nil
	}

	return os.Open(p)
}
