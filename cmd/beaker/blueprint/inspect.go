package blueprint

import (
	"context"
	"encoding/json"
	"os"

	"github.com/allenai/beaker-api/api"
	beaker "github.com/allenai/beaker-api/client"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker-cli/config"
)

type inspectOptions struct {
	blueprints []string
}

func newInspectCmd(
	parent *kingpin.CmdClause,
	parentOpts *blueprintOptions,
	config *config.Config,
) {
	o := &inspectOptions{}
	cmd := parent.Command("inspect", "Display detailed information about one or more blueprints")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Arg("blueprint", "Blueprint name or ID").Required().StringsVar(&o.blueprints)
}

func (o *inspectOptions) run(beaker *beaker.Client) error {
	ctx := context.TODO()

	var blueprints []*api.Blueprint
	for _, name := range o.blueprints {
		blueprint, err := beaker.Blueprint(ctx, name)
		if err != nil {
			return err
		}

		info, err := blueprint.Get(ctx)
		if err != nil {
			return err
		}
		blueprints = append(blueprints, info)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	return encoder.Encode(blueprints)
}
