package experiment

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	yaml "gopkg.in/yaml.v2"

	"github.com/allenai/beaker/cmd/beaker/image"
	"github.com/allenai/beaker/config"
)

type runOptions struct {
	dryRun     bool
	expandVars bool
	specFile   *os.File
	name       string
	quiet      bool
	specArgs   specArgs
}

type specArgs struct {
	blueprint  string
	image      string
	resultPath string
	desc       string
	args       []string
	env        []string
	sources    []string
	cpu        float64
	memory     string
	gpuCount   int
	gpuType    string
}

func newRunCmd(
	parent *kingpin.CmdClause,
	parentOpts *experimentOptions,
	config *config.Config,
) {
	o := &runOptions{}
	cmd := parent.Command("run", "Run an experiment")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Flag("dry-run", "Show what will be submitted and exit.").BoolVar(&o.dryRun)
	cmd.Flag("expand-vars", "Expand occurrences of '$VAR' and '${VAR}' in the experiment spec file from environment variables. Default true.").
		Default("true").
		BoolVar(&o.expandVars)
	cmd.Flag("file", "Load experiment spec from a file.").Short('f').FileVar(&o.specFile)
	cmd.Flag("name", "Assign a name to the experiment").Short('n').StringVar(&o.name)
	cmd.Flag("quiet", "Only display the experiment's unique ID").Short('q').BoolVar(&o.quiet)

	// File spec alternatives
	cmd.Flag("blueprint", "Blueprint containing code to run").StringVar(&o.specArgs.blueprint)
	cmd.Flag("image", "Beaker image containing code to run").StringVar(&o.specArgs.image)
	cmd.Flag("desc", "Optional description for the experiment").StringVar(&o.specArgs.desc)
	cmd.Flag("result-path", "Path within the container to which results will be written").
		PlaceHolder("PATH").Required().StringVar(&o.specArgs.resultPath)
	cmd.Flag("env", "Set environment variables (e.g. NAME=value or NAME)").StringsVar(&o.specArgs.env)
	cmd.Flag("source", "Bind a remote data source (e.g. source-id:/target/path)").StringsVar(&o.specArgs.sources)
	cmd.Flag("cpu", "CPUs to reserve for this experiment (e.g., 0.5)").FloatVar(&o.specArgs.cpu)
	cmd.Flag("memory", "Memory to reserve for this experiment (e.g., 1GB)").StringVar(&o.specArgs.memory)
	cmd.Flag("gpu-count", "GPUs to use for this experiment (e.g., 2)").IntVar(&o.specArgs.gpuCount)
	cmd.Flag("gpu-type", "GPU type to use for this experiment (e.g., 'p100' or 'v100')").StringVar(&o.specArgs.gpuType)

	cmd.Arg("arg", "Argument to the Docker image").StringsVar(&o.specArgs.args)
}

func (o *runOptions) run(beaker *beaker.Client) error {
	ctx := context.TODO()

	if o.specFile != nil {
		return errors.Errorf("--file argument is no longer supported; experiment specs can be run with 'experiment create'")
	}

	spec, err := specFromArgs(o.specArgs)
	if err != nil {
		return err
	}

	// Assume spec images refer to beaker images.
	images := map[string]string{} // Map Docker image tags to beaker image IDs.
	for _, task := range spec.Tasks {
		specImage := task.Spec.Image

		// Blueprints take priority over images. Enforce that we have exactly one.
		if task.Spec.Blueprint != "" {
			continue
		}
		if specImage == "" {
			return errors.Errorf("task %q must declare either a beaker image or Docker image to run", task.Name)
		}
	}

	if o.dryRun {
		fmt.Println("Experiment spec to be submitted:")
		fmt.Println()
		return printSpec(spec)
	}

	_, err = Create(ctx, os.Stdout, beaker, spec, &CreateOptions{Name: o.name, Quiet: o.quiet})
	if err == nil {
		return err
	}

	// Create beaker images in place of Docker images, assuming spec images refer to Docker images.
	images = map[string]string{} // Map Docker image tags to beaker image IDs.
	color.Red("The --image flag should be used to specify the beaker image to use.\nPassing Docker images to the run command will be deprecated soon.\n")

	for i, task := range spec.Tasks {
		specImage := task.Spec.Image
		spec.Tasks[i].Spec.Image = ""

		// Blueprints take priority over images. Enforce that we have exactly one.
		if task.Spec.Blueprint != "" {
			continue
		}
		if specImage == "" {
			return errors.Errorf("task %q must declare either a beaker image or Docker image to run", task.Name)
		}

		imageID, ok := images[specImage]
		if !ok {
			var err error
			imageID, err = image.Create(ctx, os.Stdout, beaker, specImage, nil)
			if err != nil {
				return errors.WithMessage(err, "failed to create beaker image for Docker image "+strconv.Quote(specImage))
			}
			images[specImage] = imageID
		}
		spec.Tasks[i].Spec.Blueprint = imageID
	}

	_, err = Create(ctx, os.Stdout, beaker, spec, &CreateOptions{Name: o.name, Quiet: o.quiet})
	return err
}

func specFromArgs(args specArgs) (*ExperimentSpec, error) {
	image := args.image
	spec := TaskSpec{
		Blueprint:  args.blueprint,
		Image:      image,
		ResultPath: args.resultPath,
		Arguments:  args.args,
		Requirements: Requirements{
			CPU:      args.cpu,
			Memory:   args.memory,
			GPUCount: args.gpuCount,
			GPUType:  args.gpuType,
		},
	}

	for _, env := range args.env {
		splitEnv := strings.SplitN(env, "=", 2)
		if spec.Env == nil {
			spec.Env = make(map[string]string)
		}
		if len(splitEnv) > 1 {
			spec.Env[splitEnv[0]] = splitEnv[1]
		} else {
			// Expand the environment variable if no value is specified.
			spec.Env[splitEnv[0]] = os.Getenv(splitEnv[0])
		}
	}

	for _, source := range args.sources {
		splitSource := strings.Split(source, ":")
		if len(splitSource) != 2 {
			return nil, errors.Errorf("malformed source '%s': should be of the form 'source:target'", source)
		}

		spec.Mounts = append(spec.Mounts, DatasetMount{
			DatasetID:     splitSource[0], // May be name or ID.
			ContainerPath: splitSource[1],
		})
	}

	return &ExperimentSpec{
		Description: args.desc,
		Tasks:       []ExperimentTaskSpec{{Spec: spec}},
	}, nil
}

func experimentURL(serviceAddress string, experimentID string) string {
	return fmt.Sprintf("%s/ex/%s", serviceAddress, experimentID)
}

func printSpec(spec *ExperimentSpec) error {
	y, err := yaml.Marshal(spec)
	if err != nil {
		return err
	}
	fmt.Println(string(y))
	return nil
}
