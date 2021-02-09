package main

import "github.com/spf13/cobra"

func newExecutorCommand() *cobra.Command {
	// The executor is not supported on Mac.
	return &cobra.Command{}
}
