package main

import (
	"context"
	"errors"
	"fmt"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	beaker "github.com/allenai/beaker/client"
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

	fmt.Printf("Authenticating with user token: %q\n\n", config.UserToken)

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

	fmt.Printf("Authenticated as user: %q (%s)\n", user.Name, user.ID)
	return nil
}

func helpWithUserToken() string {
	return "Login on the Beaker website and follow the instructions to configure this Beaker CLI client."
}
