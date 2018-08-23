package group

import (
	"context"
	"encoding/json"
	"os"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/api"
	beaker "github.com/allenai/beaker/client"
	"github.com/allenai/beaker/config"
)

type inspectOptions struct {
	contents bool
	groups   []string
}

func newInspectCmd(
	parent *kingpin.CmdClause,
	parentOpts *groupOptions,
	config *config.Config,
) {
	o := &inspectOptions{}
	cmd := parent.Command("inspect", "Display detailed information about one or more groups")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Flag("contents", "Include group contents in output").BoolVar(&o.contents)
	cmd.Arg("group", "Group name or ID").Required().StringsVar(&o.groups)
}

func (o *inspectOptions) run(beaker *beaker.Client) error {
	type detail struct {
		api.Group
		Experiments []string `json:"experiments,omitempty"`
	}

	ctx := context.TODO()

	var groups []detail
	for _, name := range o.groups {
		group, err := beaker.Group(ctx, name)
		if err != nil {
			return err
		}

		info, err := group.Get(ctx)
		if err != nil {
			return err
		}

		var experiments []string
		if o.contents {
			if experiments, err = group.Experiments(ctx); err != nil {
				return err
			}
		}

		groups = append(groups, detail{*info, experiments})
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	return encoder.Encode(groups)
}
