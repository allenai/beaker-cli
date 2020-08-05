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

	"code.cloudfoundry.org/bytefmt"
	"github.com/beaker/client/api"
	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	yaml "gopkg.in/yaml.v2"

	configCmd "github.com/allenai/beaker/cmd/beaker/config"
	"github.com/allenai/beaker/config"
)

const (
	low    = "low"
	normal = "normal"
	high   = "high"
)

// CreateOptions wraps options used to create an experiment.
type CreateOptions struct {
	Name     string
	Quiet    bool
	Force    bool
	Priority string
}

func newCreateCmd(
	parent *kingpin.CmdClause,
	parentOpts *experimentOptions,
	cfg *config.Config,
) {
	opts := &CreateOptions{}
	var workspace string
	var specPath string

	cmd := parent.Command("create", "Create a new experiment")
	cmd.Flag("file", "Load experiment spec from a file.").Short('f').StringVar(&specPath)
	cmd.Flag("name", "Assign a name to the experiment").Short('n').StringVar(&opts.Name)
	cmd.Flag("quiet", "Only display created experiment's ID").Short('q').BoolVar(&opts.Quiet)
	cmd.Flag("workspace", "Workspace where the experiment will be placed").Short('w').StringVar(&workspace)
	cmd.Flag("force", "Allow depending on uncommitted datasets").BoolVar(&opts.Force)
	cmd.Flag("priority", "Assign an execution priority to the experiment").Short('p').EnumVar(&opts.Priority, low, normal, high)

	cmd.Action(func(c *kingpin.ParseContext) error {
		specFile, err := openPath(specPath)
		if err != nil {
			return err
		}

		beaker, err := beaker.NewClient(parentOpts.addr, cfg.UserToken)
		if err != nil {
			return err
		}

		spec, err := ReadSpec(specFile)
		if err != nil {
			return err
		}

		if workspace != "" {
			// Workspace flag overrides what's written in the spec.
			spec.Workspace = workspace
		}
		if spec.Workspace == "" {
			// Neither spec nor args specified a workspace, so find the default.
			spec.Workspace, err = configCmd.EnsureDefaultWorkspace(beaker, cfg)
			if err != nil {
				return err
			}
			if !opts.Quiet {
				fmt.Printf("Using workspace %s\n", color.BlueString(spec.Workspace))
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
	spec *api.ExperimentSpec,
	opts *CreateOptions,
) (string, error) {
	if w == nil {
		w = ioutil.Discard
	}
	if opts == nil {
		opts = &CreateOptions{}
	}

	experiment, err := beaker.CreateExperiment(ctx, *spec, opts.Name, opts.Force, opts.Priority)
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
	Env map[string]string
}

// ReadSpec reads an experiment spec from YAML.
func ReadSpec(r io.Reader) (*api.ExperimentSpec, error) {
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
	if err := specTemplate.Execute(buf, templateParams{Env: envVars}); err != nil {
		return nil, err
	}

	var spec api.ExperimentSpec
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

// CanonicalizeJSONSpec fills out fields used by the API from YAML fields parsed from disk.
func CanonicalizeJSONSpec(spec *api.ExperimentSpec) error {
	// FUTURE: This should be unnecessary when the service accepts YAML directly.
	for i := range spec.Tasks {
		reqs := &spec.Tasks[i].Spec.Requirements
		if reqs.CPU < 0 {
			return errors.Errorf("couldn't parse cpu argument '%.2f' because it was negative", reqs.CPU)
		}
		reqs.MilliCPU = int(reqs.CPU * 1000)
		if reqs.MemoryHuman != "" {
			bytes, err := bytefmt.ToBytes(reqs.MemoryHuman)
			if err != nil {
				return errors.Wrapf(err, "invalid memory value %q", reqs.MemoryHuman)
			}
			reqs.Memory = int64(bytes)
		}
	}
	return nil
}
