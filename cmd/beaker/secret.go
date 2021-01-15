package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"text/tabwriter"
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
	cmd := &cobra.Command{
		Use:   "delete <secret>",
		Short: "Permanently delete a secret",
		Args:  cobra.ExactArgs(1),
	}

	var workspace string
	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Secret workspace")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		workspace, err := beaker.Workspace(ctx, workspace)
		if err != nil {
			return err
		}

		return workspace.DeleteSecret(ctx, args[0])
	}
	return cmd
}

func newSecretListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the metadata of all secrets in a workspace",
		Args:  cobra.NoArgs,
	}

	var workspace string
	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Secret workspace")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		workspace, err := beaker.Workspace(ctx, workspace)
		if err != nil {
			return err
		}

		secrets, err := workspace.ListSecrets(ctx)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		const rowFormat = "%s\t%s\t%s\n"
		fmt.Fprintf(w, rowFormat, "NAME", "CREATED", "UPDATED")
		for _, secret := range secrets {
			fmt.Fprintf(w, rowFormat,
				secret.Name,
				secret.Created.Format(time.RFC3339),
				secret.Updated.Format(time.RFC3339))
		}
		return w.Flush()
	}
	return cmd
}

func newSecretReadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read <secret>",
		Short: "Read the value of a secret",
		Args:  cobra.ExactArgs(1),
	}

	var workspace string
	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Secret workspace")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		workspace, err := beaker.Workspace(ctx, workspace)
		if err != nil {
			return err
		}

		secret, err := workspace.ReadSecret(ctx, args[0])
		if err != nil {
			return err
		}
		fmt.Printf("%s", secret)
		return nil
	}
	return cmd
}

func newSecretWriteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "write <secret> <value?>",
		Short: "Write a new secret or update an existing secret",
		Args:  cobra.RangeArgs(1, 2),
	}

	var workspace string
	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Secret workspace")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		workspace, err := beaker.Workspace(ctx, workspace)
		if err != nil {
			return err
		}

		var value []byte
		if len(args) == 1 {
			if value, err = ioutil.ReadAll(os.Stdin); err != nil {
				return err
			}
		} else {
			value = []byte(args[1])
		}

		_, err = workspace.PutSecret(ctx, args[0], value)
		return err
	}
	return cmd
}
