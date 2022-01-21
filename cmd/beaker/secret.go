package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
)

func newSecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret <command>",
		Short: "Manage secrets",
	}

	var workspace string
	cmd.PersistentFlags().StringVarP(&workspace, "workspace", "w", "", "Workspace of the secret")

	cmd.AddCommand(newSecretDeleteCommand(&workspace))
	cmd.AddCommand(newSecretListCommand(&workspace))
	cmd.AddCommand(newSecretReadCommand(&workspace))
	cmd.AddCommand(newSecretWriteCommand(&workspace))
	return cmd
}

func newSecretDeleteCommand(workspace *string) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <secret>",
		Short: "Permanently delete a secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := ensureWorkspace(*workspace)
			if err != nil {
				return err
			}
			return beaker.Workspace(workspace).DeleteSecret(ctx, args[0])
		},
	}
}

func newSecretListCommand(workspace *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the metadata of all secrets in a workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := ensureWorkspace(*workspace)
			if err != nil {
				return err
			}
			secrets, err := beaker.Workspace(workspace).ListSecrets(ctx)
			if err != nil {
				return err
			}
			return printSecrets(secrets)
		},
	}
}

func newSecretReadCommand(workspace *string) *cobra.Command {
	return &cobra.Command{
		Use:   "read <secret>",
		Short: "Read the value of a secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := ensureWorkspace(*workspace)
			if err != nil {
				return err
			}
			secret, err := beaker.Workspace(workspace).ReadSecret(ctx, args[0])
			if err != nil {
				return err
			}
			fmt.Printf("%s", secret)
			return nil
		},
	}
}

func newSecretWriteCommand(workspace *string) *cobra.Command {
	return &cobra.Command{
		Use:   "write <secret> [value]",
		Short: "Write a new secret or update an existing secret",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := ensureWorkspace(*workspace)
			if err != nil {
				return err
			}

			var value []byte
			if len(args) == 1 {
				var err error
				if value, err = ioutil.ReadAll(os.Stdin); err != nil {
					return err
				}
			} else {
				value = []byte(args[1])
			}

			_, err = beaker.Workspace(workspace).PutSecret(ctx, args[0], value)
			return err
		},
	}
}
