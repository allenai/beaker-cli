package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/beaker/client/api"
	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type inspectOptions struct {
	clusters []string
}

func newInspectCmd(
	parent *kingpin.CmdClause,
	parentOpts *clusterOptions,
	config *config.Config,
) {
	o := &inspectOptions{}
	cmd := parent.Command("inspect", "Display detailed information about one or more clusters")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Arg("cluster", "Cluster(s) to inspect").Required().StringsVar(&o.clusters)
}

func (o *inspectOptions) run(beaker *beaker.Client) error {
	ctx := context.TODO()

	var clusters []*api.Cluster
	for _, id := range o.clusters {
		info, err := beaker.Cluster(id).Get(ctx)
		if err != nil {
			return err
		}
		clusters = append(clusters, info)
	}

	// TODO: Print this in a more human-friendly way, and include node summary.
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	return encoder.Encode(clusters)
}
