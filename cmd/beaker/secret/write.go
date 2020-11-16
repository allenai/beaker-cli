package secret

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	beaker "github.com/beaker/client/client"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type writeOptions struct {
	workspace string
	name      string
	value     string
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
	cmd.Arg("value", `The value of the secret.
If the value begins with "@", it is loaded from a file.
If the value is "-", it is read from stdin.`).Required().StringVar(&o.value)

	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})
}

func (o *writeOptions) run(beaker *beaker.Client) error {
	var value []byte
	var err error
	switch {
	case strings.HasPrefix(o.value, "@"):
		value, err = ioutil.ReadFile(strings.TrimPrefix(o.value, "@"))
	case o.value == "-":
		value, err = ioutil.ReadAll(os.Stdin)
	default:
		value = []byte(o.value)
	}
	if err != nil {
		return err
	}

	fmt.Printf("%s=%s\n", o.name, value)

	// TODO Write the secret
	return nil
}
