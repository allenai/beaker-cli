package secret

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"text/tabwriter"
	"time"

	beaker "github.com/beaker/client/client"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/cmd/beaker/options"
	"github.com/allenai/beaker/config"
)

type secretOptions struct {
	*options.AppOptions
	addr string
}

// NewSecretCmd creates the root command for this subpackage.
func NewSecretCmd(
	parent *kingpin.Application,
	parentOpts *options.AppOptions,
	config *config.Config,
) {
	o := &secretOptions{AppOptions: parentOpts}
	cmd := parent.Command("secret", "Manage secrets")

	cmd.Flag("addr", "Address of the Beaker service.").Default(config.BeakerAddress).StringVar(&o.addr)

	// Add automatic help generation for the command group.
	var helpSubcommands []string
	cmd.Command("help", "Show help.").Hidden().Default().PreAction(func(c *kingpin.ParseContext) error {
		fullCommand := append([]string{cmd.Model().Name}, helpSubcommands...)
		parent.Usage(fullCommand)
		return nil
	}).Arg("command", "Show help on command.").StringsVar(&helpSubcommands)

	newWriteCmd(cmd, o, config)
	newReadCmd(cmd, o, config)
	newListCmd(cmd, o, config)
	newDeleteCmd(cmd, o, config)
}

type writeOptions struct {
	workspace string
	name      string
	value     string
	stdin     bool
}

func newWriteCmd(
	parent *kingpin.CmdClause,
	parentOpts *secretOptions,
	config *config.Config,
) {
	o := &writeOptions{}
	cmd := parent.Command("write", "Write a new secret or update an existing secret")
	cmd.Flag("workspace", "Workspace containing the secret.").Required().StringVar(&o.workspace)
	cmd.Arg("name", "The name of the secret.").Required().StringVar(&o.name)
	cmd.Arg("value", "The value of the secret.").StringVar(&o.value)

	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}

		ctx := context.Background()
		workspace, err := beaker.Workspace(ctx, o.workspace)
		if err != nil {
			return err
		}

		var value []byte
		if o.value == "" {
			if value, err = ioutil.ReadAll(os.Stdin); err != nil {
				return err
			}
		} else {
			value = []byte(o.value)
		}

		_, err = workspace.PutSecret(ctx, o.name, value)
		return err
	})
}

type readOptions struct {
	workspace string
	name      string
}

func newReadCmd(
	parent *kingpin.CmdClause,
	parentOpts *secretOptions,
	config *config.Config,
) {
	o := &readOptions{}
	cmd := parent.Command("read", "Read the value of a secret")
	cmd.Flag("workspace", "Workspace containing the secret.").Required().StringVar(&o.workspace)
	cmd.Arg("name", "The name of the secret.").Required().StringVar(&o.name)

	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}

		ctx := context.Background()
		workspace, err := beaker.Workspace(ctx, o.workspace)
		if err != nil {
			return err
		}

		secret, err := workspace.ReadSecret(ctx, o.name)
		if err != nil {
			return err
		}
		fmt.Printf("%s", secret)
		return nil
	})
}

type listOptions struct {
	workspace string
}

func newListCmd(
	parent *kingpin.CmdClause,
	parentOpts *secretOptions,
	config *config.Config,
) {
	o := &listOptions{}
	cmd := parent.Command("list", "List the metadata of all secrets in a workspace")
	cmd.Flag("workspace", "Workspace to list secrets.").Required().StringVar(&o.workspace)
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}

		ctx := context.Background()
		workspace, err := beaker.Workspace(ctx, o.workspace)
		if err != nil {
			return err
		}

		secrets, err := workspace.ListSecrets(ctx)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		const rowFormat = "%s\t%s\t%s\n"
		fmt.Fprintf(w, rowFormat, "NAME", "CREATED", "UPDATED")
		for _, secret := range secrets {
			fmt.Fprintf(w, rowFormat,
				secret.Name,
				secret.Created.Format(time.RFC3339),
				secret.Updated.Format(time.RFC3339))
		}
		return w.Flush()
	})
}

type deleteOptions struct {
	workspace string
	name      string
}

func newDeleteCmd(
	parent *kingpin.CmdClause,
	parentOpts *secretOptions,
	config *config.Config,
) {
	o := &deleteOptions{}
	cmd := parent.Command("delete", "Permanently delete a secret")
	cmd.Flag("workspace", "Workspace containing the secret.").Required().StringVar(&o.workspace)
	cmd.Arg("name", "The name of the secret.").Required().StringVar(&o.name)

	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}

		ctx := context.Background()
		workspace, err := beaker.Workspace(ctx, o.workspace)
		if err != nil {
			return err
		}

		return workspace.DeleteSecret(ctx, o.name)
	})
}
