package main

import (
	"context"
	"errors"
	"fmt"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	beaker "github.com/beaker/client/client"

	"github.com/allenai/beaker/config"
)

type configOptions struct{}

// NewConfigCmd allows the user to save non-default configuration values.
func NewConfigCmd(parent *kingpin.Application) {
	o := &configOptions{}
	cmd := parent.Command("configure", "Configure Beaker options")
	cmd.Command("interactive", "Interactive configuration").Default().Action(o.interactive)
	cmd.Command("test", "Test the configuration").Action(o.testConnection)
}

func (o *configOptions) interactive(_ *kingpin.ParseContext) error {
	return config.InteractiveConfiguration()
}

func (o *configOptions) testConnection(_ *kingpin.ParseContext) error {
	fmt.Println("Beaker Configuration Test")
	fmt.Println("")

	// Create a default config by reading in whatever config currently exists.
	config, err := config.New()
	if err != nil {
		return err
	}

	if len(config.UserToken) == 0 {
		fmt.Println("You don't have a user token configured.")
		fmt.Println(helpWithUserToken())
		return errors.New("user token not configured")
	}

	beaker, err := beaker.NewClient(config.BeakerAddress, config.UserToken)
	if err != nil {
		return err
	}

	user, err := beaker.WhoAmI(context.TODO())
	if err != nil {
		fmt.Println("There was a problem authenticating with your user token.")
		fmt.Println(helpWithUserToken())
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
}

func helpWithUserToken() string {
	return "Login on the Beaker website and follow the instructions to configure this Beaker CLI client."
}
