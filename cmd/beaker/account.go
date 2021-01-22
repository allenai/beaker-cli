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
	cmd.AddCommand(newAccountWhoAmICommand())
	//cmd.AddCommand(newAccountListCommand())
	cmd.AddCommand(newAccountOrganizationsCommand())
	return cmd
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
