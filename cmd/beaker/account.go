package main

import (
	"github.com/beaker/client/api"
	"github.com/spf13/cobra"
)

func newAccountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account <command>",
		Short: "Manage accounts",
	}
	cmd.AddCommand(newAccountListCommand())
	cmd.AddCommand(newAccountOrganizationsCommand())
	cmd.AddCommand(newAccountWhoAmICommand())
	return cmd
}

func newAccountListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all accounts",
		Args:  cobra.NoArgs,
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
