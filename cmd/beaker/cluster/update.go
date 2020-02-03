package cluster

import (
	"context"
	"fmt"

	"github.com/beaker/client/api"
	beaker "github.com/beaker/client/client"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type updateOptions struct {
	cluster string
	scale   int
}

func newUpdateCmd(
	parent *kingpin.CmdClause,
	parentOpts *clusterOptions,
	config *config.Config,
) {
	o := &updateOptions{
		scale: -1, // Impossible default signals absence of the arg.
	}
	cmd := parent.Command("update", "Modify a cluster")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Flag("max-size", "Maximum number of nodes").IntVar(&o.scale)
	cmd.Arg("cluster", "Cluster to update").Required().StringVar(&o.cluster)
}

func (o *updateOptions) run(beaker *beaker.Client) error {
	patch := api.ClusterPatch{}
	if o.scale >= 0 {
		patch.Capacity = &o.scale
	}

	if (patch == api.ClusterPatch{}) {
		fmt.Println("Nothing to update.")
		return nil
	}

	cluster, err := beaker.Cluster(o.cluster).Patch(context.TODO(), &patch)
	if err != nil {
		return err
	}

	fmt.Println(cluster.ID)
	return nil
}
