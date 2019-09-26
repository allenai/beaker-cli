package cluster

import (
	"context"
	"fmt"
	"strings"

	"github.com/beaker/client/api"
	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type createOptions struct {
	name        string
	size        int
	preemptible bool
	cpuCount    int
	gpuCount    int
	gpuType     string
	memory      string
	galaxy      string
}

func newCreateCmd(
	parent *kingpin.CmdClause,
	parentOpts *clusterOptions,
	config *config.Config,
) {
	o := &createOptions{}
	cmd := parent.Command("create", "Create a new cluster")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Flag("galaxy", "Override default environment").StringVar(&o.galaxy)
	cmd.Flag("max-size", "Maximum number of instances").IntVar(&o.size)
	cmd.Flag("preemptible", "Enable cheaper but more volatile nodes").BoolVar(&o.preemptible)
	cmd.Flag("cpu-count", "Number of CPUs per instance").IntVar(&o.cpuCount)
	cmd.Flag("gpu-count", "Number of GPUs per instance").IntVar(&o.gpuCount)
	cmd.Flag("gpu-type", "Type of GPU, e.g. nvidia-p100").StringVar(&o.gpuType)
	cmd.Flag("memory", "Memory limit per instance, e.g. 6.5GiB").StringVar(&o.memory)
	cmd.Arg("name", "Fully qualified name to assign to the cluster").Required().StringVar(&o.name)
}

func (o *createOptions) run(beaker *beaker.Client) error {
	fmt.Println(color.YellowString("Cluster commands are still under development and should be considered experimental."))

	parts := strings.Split(o.name, "/")
	if len(parts) != 2 {
		return fmt.Errorf("cluster names must be fully scoped in the form %s", color.GreenString("galaxy/cluster"))
	}

	account, clusterName := parts[0], parts[1]
	spec := api.ClusterSpec{
		Name:        clusterName,
		Galaxy:      o.galaxy,
		Capacity:    o.size,
		Preemptible: o.preemptible,
		Spec: api.InstanceSpec{
			// TODO: If any field is omitted, try to calculate some sane defaults?
			CPUCount: o.cpuCount,
			GPUCount: o.gpuCount,
			GPUType:  o.gpuType,
			Memory:   o.memory,
		},
	}

	ctx := context.TODO()
	cluster, err := beaker.CreateCluster(ctx, account, spec)
	if err != nil {
		return err
	}

	fmt.Printf("Cluster %s created (ID %s)\n", color.BlueString(cluster.Name), color.BlueString(cluster.ID))
	return nil
}
