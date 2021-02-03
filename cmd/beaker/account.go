package main

import (
	"fmt"

	"github.com/allenai/beaker/config"

	"github.com/beaker/client/api"
	"github.com/spf13/cobra"
)

func newAccountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account <command>",
		Short: "Manage accounts",
	}
	cmd.AddCommand(newAccountGenerateTokenCommand())
	cmd.AddCommand(newAccountListCommand())
	cmd.AddCommand(newAccountOrganizationsCommand())
	cmd.AddCommand(newAccountTokenCommand())
	cmd.AddCommand(newAccountWhoAmICommand())
	return cmd
}

func newAccountGenerateTokenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-token",
		Short: "Generate a new token for authentication",
		Args:  cobra.NoArgs,
	}

	var noUpdateConfig bool
	cmd.Flags().BoolVar(&noUpdateConfig, "no-update-config", false, "Don't update config with the new token.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		confirmed, err := confirm(`Generating a new token will invalidate your old token.
Are you sure want to generate a new token?`)
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}

		token, err := beaker.GenerateToken(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("New token: %q\n", token)

		if noUpdateConfig {
			return nil
		}

		beakerConfig.UserToken = token
		if err := config.WriteConfig(beakerConfig, config.GetFilePath()); err != nil {
			return err
		}
		fmt.Println("New token written to config")
		return nil
	}
	return cmd
}

func newAccountListCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "list",
		Short:  "List all accounts",
		Args:   cobra.NoArgs,
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var users []api.UserDetail
			var cursor string
			for {
				var page []api.UserDetail
				var err error
				page, cursor, err = beaker.ListUsers(ctx, cursor)
				if err != nil {
					return err
				}
				users = append(users, page...)
				if cursor == "" {
					break
				}
			}
			return printUsers(users)
		},
	}
}

func newAccountOrganizationsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "organizations",
		Short: "List organizations that you are a member of",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			orgs, err := beaker.ListMyOrgs(ctx)
			if err != nil {
				return err
			}
			return printOrganizations(orgs)
		},
	}
}

func newAccountTokenCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "token",
		Short: "Print user token",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Println(beakerConfig.UserToken)
			return err
		},
	}
}

func newAccountWhoAmICommand() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Display information about your account",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			user, err := beaker.WhoAmI(ctx)
			if err != nil {
				return err
			}
			return printUsers([]api.UserDetail{*user})
		},
	}
}
