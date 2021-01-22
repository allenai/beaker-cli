package main

import (
	"github.com/beaker/client/api"
	"github.com/spf13/cobra"
)

func newOrganizationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "organization <command>",
		Short: "Manage organizations",
	}
	//cmd.AddCommand(newOrganizationCreateCommand())
	//cmd.AddCommand(newOrganizationListCommand())
	cmd.AddCommand(newOrganizationMembersCommand())
	return cmd
}

func newOrganizationCreateCommand() *cobra.Command {
	// TODO client support
	return nil
}

func newOrganizationListCommand() *cobra.Command {
	// TODO client support
	return nil
}

func newOrganizationMembersCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member <command>",
		Short: "Manage organization membership",
	}
	cmd.AddCommand(newOrganizationMemberAddCommand())
	cmd.AddCommand(newOrganizationMemberInspectCommand())
	cmd.AddCommand(newOrganizationMemberListCommand())
	cmd.AddCommand(newOrganizationMemberRemoveCommand())
	return cmd
}

func newOrganizationMemberAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <organization> <account>",
		Short: "Add an account to an organization",
		Args:  cobra.ExactArgs(2),
	}

	var role string
	cmd.Flags().StringVar(&role, "role", "member", `Role in the organization, defaults to "member"`)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return beaker.Organization(args[0]).SetMember(ctx, args[1], role)
	}
	return cmd
}

func newOrganizationMemberInspectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <organization> <member...>",
		Short: "Display detailed information about one or more organization members",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var members []api.OrgMembership
			for _, name := range args[1:] {
				member, err := beaker.Organization(args[0]).GetMember(ctx, name)
				if err != nil {
					return err
				}
				members = append(members, *member)
			}
			return printMembers(members)
		},
	}
}

func newOrganizationMemberListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <organization>",
		Short: "List members of an organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var users []api.UserDetail
			var cursor string
			for {
				var page []api.UserDetail
				var err error
				page, cursor, err = beaker.Organization(args[0]).ListMembers(ctx, cursor)
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

func newOrganizationMemberRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <organization> <member>",
		Short: "Remove a member from an organization",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return beaker.Organization(args[0]).RemoveMember(ctx, args[1])
		},
	}
}
