package group

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/beaker/client/api"
	beaker "github.com/beaker/client/client"

	"github.com/allenai/beaker/config"
)

type createOptions struct {
	description string
	name        string
	quiet       bool
	org         string
	workspace   string
	experiments []string
}

func newCreateCmd(
	parent *kingpin.CmdClause,
	parentOpts *groupOptions,
	config *config.Config,
) {
	o := &createOptions{}
	cmd := parent.Command("create", "Create a new experiment group")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		if o.org == "" {
			o.org = config.DefaultOrg
		}
		if o.workspace == "" {
			o.workspace = config.DefaultWorkspace
		}
		if o.workspace == "" {
			return errors.New("groups must be created within a workspace")
		}
		return o.run(beaker)
	})

	cmd.Flag("desc", "Assign a description to the group").StringVar(&o.description)
	cmd.Flag("name", "Assign a name to the group").Short('n').Required().StringVar(&o.name)
	cmd.Flag("quiet", "Only display created group's ID").Short('q').BoolVar(&o.quiet)
	cmd.Flag("org", "Org that will own the created group").Short('o').StringVar(&o.org)
	cmd.Flag("workspace", "Workspace where the group will be placed").Short('w').StringVar(&o.workspace)
	cmd.Arg("experiment", "ID of experiment to add to the group").StringsVar(&o.experiments)
}

func (o *createOptions) run(beaker *beaker.Client) error {
	spec := api.GroupSpec{
		Name:         o.name,
		Description:  o.description,
		Organization: o.org,
		Workspace:    o.workspace,
		Experiments:  trimAndUnique(o.experiments),
	}
	group, err := beaker.CreateGroup(context.TODO(), spec)
	if err != nil {
		return err
	}

	if o.quiet {
		fmt.Println(group.ID())
	} else {
		fmt.Println("Created group " + color.BlueString(group.ID()))
	}
	return nil
}
