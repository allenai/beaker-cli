package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func newSecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret <command>",
		Short: "Manage secrets",
	}
	cmd.AddCommand(newSecretDeleteCommand())
	cmd.AddCommand(newSecretListCommand())
	cmd.AddCommand(newSecretReadCommand())
	cmd.AddCommand(newSecretWriteCommand())
	return cmd
}

func newSecretDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <workspace> <secret>",
		Short: "Permanently delete a secret",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := beaker.Workspace(ctx, args[0])
			if err != nil {
				return err
			}

			return workspace.DeleteSecret(ctx, args[1])
		},
	}
}

func newSecretListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <workspace>",
		Short: "List the metadata of all secrets in a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := beaker.Workspace(ctx, args[0])
			if err != nil {
				return err
			}

			secrets, err := workspace.ListSecrets(ctx)
			if err != nil {
				return err
			}

			if err := printTableRow("NAME", "CREATED", "UPDATED"); err != nil {
				return err
			}
			for _, secret := range secrets {
				if err := printTableRow(
					secret.Name,
					secret.Created.Format(time.RFC3339),
					secret.Updated.Format(time.RFC3339),
				); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func newSecretReadCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "read <workspace> <secret>",
		Short: "Read the value of a secret",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := beaker.Workspace(ctx, args[0])
			if err != nil {
				return err
			}

			secret, err := workspace.ReadSecret(ctx, args[1])
			if err != nil {
				return err
			}
			fmt.Printf("%s", secret)
			return nil
		},
	}
}

func newSecretWriteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "write <workspace> <secret> <value?>",
		Short: "Write a new secret or update an existing secret",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := beaker.Workspace(ctx, args[0])
			if err != nil {
				return err
			}

			var value []byte
			if len(args) == 2 {
				if value, err = ioutil.ReadAll(os.Stdin); err != nil {
					return err
				}
			} else {
				value = []byte(args[2])
			}

			_, err = workspace.PutSecret(ctx, args[1], value)
			return err
		},
	}
}
