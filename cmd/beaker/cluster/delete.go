package cluster

import (
	"context"
	"fmt"

	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type deleteOptions struct {
	cluster string
}

func newDeleteCmd(
	parent *kingpin.CmdClause,
	parentOpts *clusterOptions,
	config *config.Config,
) {
	o := &deleteOptions{}
	cmd := parent.Command("delete", "Permanently expire a cluster")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Arg("cluster", "Cluster to delete").Required().StringVar(&o.cluster)
}

func (o *deleteOptions) run(beaker *beaker.Client) error {
	fmt.Println(color.YellowString("Cluster commands are still under development and should be considered experimental."))

	if err := beaker.Cluster(o.cluster).Delete(context.TODO()); err != nil {
		return err
	}

	fmt.Printf("Deleted %s\n", color.BlueString(o.cluster))
	return nil
}
