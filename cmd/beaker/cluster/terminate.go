package cluster

import (
	"context"
	"fmt"

	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type terminateOptions struct {
	cluster string
}

func newTerminateCmd(
	parent *kingpin.CmdClause,
	parentOpts *clusterOptions,
	config *config.Config,
) {
	o := &terminateOptions{}
	cmd := parent.Command("terminate", "Permanently expire a cluster")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Arg("cluster", "Cluster to terminate").Required().StringVar(&o.cluster)
}

func (o *terminateOptions) run(beaker *beaker.Client) error {
	fmt.Println(color.YellowString("Cluster commands are still under development and should be considered experimental."))

	if err := beaker.Cluster(o.cluster).Terminate(context.TODO()); err != nil {
		return err
	}

	fmt.Printf("Terminated %s\n", color.BlueString(o.cluster))
	return nil
}
