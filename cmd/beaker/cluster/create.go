package cluster

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

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

	cmd.Flag("max-size", "Maximum number of nodes").IntVar(&o.size)
	cmd.Flag("preemptible", "Enable cheaper but more volatile nodes").BoolVar(&o.preemptible)
	cmd.Flag("cpu-count", "Number of CPUs per node").IntVar(&o.cpuCount)
	cmd.Flag("gpu-count", "Number of GPUs per node").IntVar(&o.gpuCount)
	cmd.Flag("gpu-type", "Type of GPU, e.g. p100").StringVar(&o.gpuType)
	cmd.Flag("memory", "Memory limit per node, e.g. 6.5GiB").StringVar(&o.memory)
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
		Capacity:    o.size,
		Preemptible: o.preemptible,
		Spec: api.NodeSpec{
			CPUCount: o.cpuCount,
			GPUCount: o.gpuCount,
			GPUType:  o.gpuType,
			Memory:   o.memory,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		quit := make(chan os.Signal, 1)
		defer close(quit)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		cancel()
	}()

	cluster, err := beaker.CreateCluster(ctx, account, spec)
	if err != nil {
		return err
	}

	fmt.Printf("Cluster %s created (ID %s)\n", color.BlueString(cluster.Name), color.BlueString(cluster.ID))
	fmt.Printf("Preparing cluster...")

	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ctx.Done():
			fmt.Println(" canceled")
			os.Exit(1)

		case <-ticker.C:
			cluster, err = beaker.Cluster(cluster.ID).Get(ctx)
			if err != nil {
				fmt.Println(" failed")
				return err
			}

			switch cluster.Status {
			case api.ClusterPending:
				continue

			case api.ClusterActive:
				fmt.Println("Success!")

				gpuStr := "none"
				if gpuCount := cluster.NodeShape.GPUCount; gpuCount != 0 {
					gpuStr = strconv.FormatInt(int64(gpuCount), 10)
					if gpuType := cluster.NodeShape.GPUType; gpuType != "" {
						gpuStr += " " + gpuType
					}
				}

				fmt.Print("\nEstimated cost per node: ")
				color.Green("$%v/hour", cluster.NodeCost.Round(2))
				fmt.Println("Nodes may exceed requested parameters to optimize cost:")
				fmt.Println("    CPUs:      ", cluster.NodeShape.CPUCount)
				fmt.Println("    CPU Memory:", cluster.NodeShape.Memory)
				fmt.Println("    GPUs:      ", gpuStr)
				return nil

			case api.ClusterFailed:
				fmt.Println(" failed")
				return errors.New("please contact Beaker for assistance")

			default:
				fmt.Println(" failed")
				return fmt.Errorf("unrecognized cluster state: %s", cluster.Status)
			}
		}
	}
}
