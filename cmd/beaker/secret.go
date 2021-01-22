package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/beaker/client/api"
	"github.com/spf13/cobra"
)

func newSecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret <command>",
		Short: "Manage secrets",
	}
	cmd.AddCommand(newSecretDeleteCommand())
	cmd.AddCommand(newSecretInspectCommand())
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

func newSecretInspectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <workspace> <secret...>",
		Short: "Display detailed information about one or more secrets",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := beaker.Workspace(ctx, args[0])
			if err != nil {
				return err
			}

			var secrets []api.Secret
			for _, name := range args[1:] {
				secret, err := workspace.GetSecret(ctx, name)
				if err != nil {
					return err
				}
				secrets = append(secrets, *secret)
			}
			return printSecrets(secrets)
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
			return printSecrets(secrets)
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
