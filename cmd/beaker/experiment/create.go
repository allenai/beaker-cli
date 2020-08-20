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
	"github.com/beaker/client/client"
	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"

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
	cmd.Flag("priority", "Assign an execution priority to the experiment").Short('p').EnumVar(&opts.Priority, low, normal, high)

	cmd.Action(func(c *kingpin.ParseContext) error {
		ctx := context.TODO()
		specFile, err := openPath(specPath)
		if err != nil {
			return err
		}

		beaker, err := beaker.NewClient(parentOpts.addr, cfg.UserToken)
		if err != nil {
			return err
		}

		// Find a default workspace if there's no explicit argument.
		if workspace == "" {
			if workspace, err = configCmd.EnsureDefaultWorkspace(beaker, cfg); err != nil {
				return err
			}
			if !opts.Quiet {
				fmt.Printf("Using workspace %s\n", color.BlueString(workspace))
			}
		}

		rawSpec, err := readSpec(specFile)
		if err != nil {
			return err
		}

		var header struct {
			Version string `yaml:"version,omitempty"`
		}
		if err := yaml.Unmarshal(rawSpec, &header); err != nil {
			return err
		}

		// TODO: We should be able to blindly pass the raw spec to the service,
		// but need to update the API to accept YAML first.
		var experimentID string
		switch header.Version {
		case "v2", "v2-alpha":
			var spec api.ExperimentSpecV2
			if err := yaml.UnmarshalStrict(rawSpec, &spec); err != nil {
				return err
			}

			ws, err := beaker.Workspace(ctx, workspace)
			if err != nil {
				return err
			}
			experiment, err := ws.CreateExperiment(ctx, &spec, &client.ExperimentOpts{
				Name: opts.Name,
			})
			if err != nil {
				return err
			}
			experimentID = experiment.ID

		case "", "v1":
			var spec api.ExperimentSpec
			if err := yaml.UnmarshalStrict(rawSpec, &spec); err != nil {
				return err
			}
			if err := canonicalizeSpecV1(&spec); err != nil {
				return err
			}

			spec.Workspace = workspace
			experiment, err := beaker.CreateExperiment(ctx, spec, opts.Name, opts.Priority)
			if err != nil {
				return err
			}
			experimentID = experiment.ID()
		}

		if opts.Quiet {
			fmt.Println(experimentID)
		} else {
			fmt.Printf("Experiment %s submitted. See progress at %s/ex/%s\n",
				color.BlueString(experimentID), beaker.Address(), experimentID)
		}
		return nil
	})
}

type templateParams struct {
	Env map[string]string
}

// readSpec reads an experiment spec from YAML.
func readSpec(r io.Reader) ([]byte, error) {
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

	return buf.Bytes(), nil
}

func openPath(p string) (io.Reader, error) {
	// Special case: "-" means read from STDIN.
	if p == "-" {
		return os.Stdin, nil
	}
	return os.Open(p)
}

// canonicalizeSpecV1 fills out JSON fields used by the API from YAML fields parsed from disk.
func canonicalizeSpecV1(spec *api.ExperimentSpec) error {
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
