package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/allenai/beaker/config"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const userTokenHelp = "Login on the Beaker website and follow the instructions to configure this Beaker CLI client."

func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Manage Beaker configuration",
	}
	cmd.AddCommand(newConfigListCommand())
	cmd.AddCommand(newConfigSetCommand())
	cmd.AddCommand(newConfigTestCommand())
	cmd.AddCommand(newConfigUnsetCommand())
	return cmd
}

func newConfigListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration properties",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			t := reflect.TypeOf(*beakerConfig)
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				propertyKey := trimTag(field.Tag.Get("yaml"))
				value := reflect.ValueOf(beakerConfig).Elem().FieldByName(field.Name).String()
				if value == "" {
					value = "(unset)"
				}
				fmt.Printf("%s = %s\n", propertyKey, color.BlueString(value))
			}
			return nil
		},
	}
}

func newConfigSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set <property> <value>",
		Short: "Set a specific config setting",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			beakerCfg := &config.Config{}
			configFilePath := config.GetFilePath()
			err := config.ReadConfigFromFile(configFilePath, beakerCfg)
			if err != nil {
				if os.IsNotExist(err) {
					beakerCfg = beakerConfig
				} else {
					return err
				}
			}

			t := reflect.TypeOf(*beakerCfg)
			found := false
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				if trimTag(field.Tag.Get("yaml")) == args[0] {
					found = true
					// The following code assumes all values are strings and will not work with non-string values.
					reflect.ValueOf(beakerCfg).Elem().FieldByName(field.Name).SetString(strings.TrimSpace(args[1]))
				}
			}
			if !found {
				return errors.New(fmt.Sprintf("Unknown config property: %q", args[0]))
			}

			return config.WriteConfig(beakerCfg, configFilePath)
		},
	}
}

func newConfigTestCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Validate configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Beaker Configuration Test")
			fmt.Println("")

			// Create a default config by reading in whatever config currently exists.
			cfg, err := config.New()
			if err != nil {
				return err
			}

			if len(cfg.UserToken) == 0 {
				fmt.Println("You don't have a user token configured.")
				fmt.Println(userTokenHelp)
				return errors.New("user token not configured")
			}

			user, err := beaker.WhoAmI(ctx)
			if err != nil {
				fmt.Println("There was a problem authenticating with your user token.")
				fmt.Println(userTokenHelp)
				return err
			}

			fmt.Printf("Authenticated as user: %q (%s)\n\n", user.Name, user.ID)

			if cfg.DefaultWorkspace == "" {
				fmt.Println("No default workspace set.")
			} else {
				fmt.Printf("Verifying default workspace: %q\n\n", cfg.DefaultWorkspace)
				if _, err := beaker.Workspace(cfg.DefaultWorkspace).Get(ctx); err != nil {
					fmt.Println("There was a problem verifying your default workspace.")
					fmt.Printf("Set the default workspace using the command %s\n", color.BlueString("beaker config set default_workspace <workspace_name>"))
					return err
				}

				fmt.Printf("Default workspace verified: %q\n", cfg.DefaultWorkspace)
			}

			return nil
		},
	}
}

func newConfigUnsetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "unset <property>",
		Short: "Unset a specific config setting",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			beakerCfg := &config.Config{}
			configFilePath := config.GetFilePath()
			err := config.ReadConfigFromFile(configFilePath, beakerCfg)
			if err != nil {
				return err
			}

			t := reflect.TypeOf(*beakerCfg)
			found := false
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				if trimTag(field.Tag.Get("yaml")) == args[0] {
					found = true
					reflect.ValueOf(beakerCfg).Elem().FieldByName(field.Name).Set(reflect.Zero(field.Type))
				}
			}
			if !found {
				return errors.New(fmt.Sprintf("Unknown config property: %q", args[0]))
			}

			fmt.Printf("Unset %s\n", args[0])

			return config.WriteConfig(beakerCfg, configFilePath)
		},
	}
}

// Remove extra fields from a YAML tag e.g. "name,omitempty" -> "name".
func trimTag(tag string) string {
	return strings.Split(tag, ",")[0]
}
