package cluster

import (
	"context"
	"fmt"

	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type extendOptions struct {
	cluster string
	name    string
}

func newExtendCmd(
	parent *kingpin.CmdClause,
	parentOpts *clusterOptions,
	config *config.Config,
) {
	o := &extendOptions{}
	cmd := parent.Command("extend", "Extend a cluster's expiration")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Arg("cluster", "Cluster to extend").Required().StringVar(&o.cluster)
}

func (o *extendOptions) run(beaker *beaker.Client) error {
	fmt.Println(color.YellowString("Cluster commands are still under development and should be considered experimental."))

	expiration, err := beaker.Cluster(o.cluster).Extend(context.TODO())
	if err != nil {
		return err
	}

	fmt.Printf("Extended %s to %v\n", color.BlueString(o.cluster), expiration)
	return nil
}
