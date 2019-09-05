package config

import (
	"context"
	"fmt"

	beaker "github.com/beaker/client/client"
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

const userTokenHelp = "Login on the Beaker website and follow the instructions to configure this Beaker CLI client."

func newTestCmd(
	parent *kingpin.CmdClause,
	parentOpts *configOptions,
	config *config.Config,
) {
	cmd := parent.Command("test", "Test the configuration")
	cmd.Action(func(c *kingpin.ParseContext) error {
		fmt.Println("Beaker Configuration Test")
		fmt.Println("")

		// Create a default config by reading in whatever config currently exists.
		config, err := beakerConfig.New()
		if err != nil {
			return err
		}

		if len(config.UserToken) == 0 {
			fmt.Println("You don't have a user token configured.")
			fmt.Println(userTokenHelp)
			return errors.New("user token not configured")
		}

		beaker, err := beaker.NewClient(config.BeakerAddress, config.UserToken)
		if err != nil {
			return err
		}

		user, err := beaker.WhoAmI(context.TODO())
		if err != nil {
			fmt.Println("There was a problem authenticating with your user token.")
			fmt.Println(userTokenHelp)
			return err
		}

		fmt.Printf("Authenticated as user: %q (%s)\n\n", user.Name, user.ID)

		if config.DefaultOrg != "" {
			fmt.Printf("Verifying default org: %q\n\n", config.DefaultOrg)
			err = beaker.VerifyOrgExists(context.TODO(), config.DefaultOrg)
			if err != nil {
				fmt.Println("There was a problem verifying your default org.")
				fmt.Println("Set the default organization in your config in the format `default_org: <org_name>`. Note that the name may be different from the name displayed in beaker UI.")
				return err
			}

			fmt.Printf("Default org verified: %q\n", config.DefaultOrg)
		} else {
			fmt.Println("No default org set.")
		}

		return nil
	})
}
