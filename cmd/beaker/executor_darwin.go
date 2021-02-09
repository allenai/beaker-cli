package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newExecutorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "executor <command>",
		Short: "Manage the executor",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("executor is not supported on Mac")
		},
	}
}
