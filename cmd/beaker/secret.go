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
			return beaker.Workspace(args[0]).DeleteSecret(ctx, args[1])
		},
	}
}

func newSecretListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <workspace>",
		Short: "List the metadata of all secrets in a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			secrets, err := beaker.Workspace(args[0]).ListSecrets(ctx)
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
			secret, err := beaker.Workspace(args[0]).ReadSecret(ctx, args[1])
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
		Use:   "write <workspace> <secret> [value]",
		Short: "Write a new secret or update an existing secret",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			var value []byte
			if len(args) == 2 {
				var err error
				if value, err = ioutil.ReadAll(os.Stdin); err != nil {
					return err
				}
			} else {
				value = []byte(args[2])
			}

			_, err := beaker.Workspace(args[0]).PutSecret(ctx, args[1], value)
			return err
		},
	}
}
