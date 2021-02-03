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
	cmd.AddCommand(newOrganizationCreateCommand())
	cmd.AddCommand(newOrganizationInspectCommand())
	cmd.AddCommand(newOrganizationListCommand())
	cmd.AddCommand(newOrganizationMembersCommand())
	return cmd
}

func newOrganizationCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "create <name>",
		Short:  "Create an organization with a name",
		Args:   cobra.ExactArgs(1),
		Hidden: true,
	}

	var displayName string
	var description string
	var owner string
	cmd.Flags().StringVar(&displayName, "display-name", "", "Organization display name")
	cmd.Flags().StringVar(&description, "description", "", "Organization description")
	cmd.Flags().StringVar(&owner, "owner", "", "Organization owner (if different than creator)")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if owner == "" {
			currentUser, err := beaker.WhoAmI(ctx)
			if err != nil {
				return err
			}
			owner = currentUser.Name
		}

		_, err := beaker.CreateOrganization(ctx, api.OrganizationSpec{
			Name:        args[0],
			Owner:       owner,
			DisplayName: displayName,
			Description: description,
		})
		return err
	}
	return cmd
}

func newOrganizationInspectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <organization...>",
		Short: "Display detailed information about one or more organizations",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var orgs []api.Organization
			for _, name := range args {
				org, err := beaker.Organization(name).Get(ctx)
				if err != nil {
					return err
				}
				orgs = append(orgs, *org)
			}
			return printOrganizations(orgs)
		},
	}
}

func newOrganizationListCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "list",
		Short:  "List all organizations",
		Args:   cobra.NoArgs,
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var orgs []api.Organization
			var cursor string
			for {
				var page []api.Organization
				var err error
				page, cursor, err = beaker.ListOrganizations(ctx, cursor)
				if err != nil {
					return err
				}
				orgs = append(orgs, page...)
				if cursor == "" {
					break
				}
			}
			return printOrganizations(orgs)
		},
	}
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
