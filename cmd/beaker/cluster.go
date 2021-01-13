package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/beaker/client/api"
	"github.com/beaker/client/client"
	"github.com/spf13/cobra"
)

func newClusterCommand(client *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage clusters",
	}
	cmd.AddCommand(newClusterInspectCommand(client))
	return cmd
}

func newClusterInspectCommand(client *client.Client) *cobra.Command {
	return &cobra.Command{
		Use:   "inspect",
		Short: "Display detailed information about one or more clusters",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var clusters []*api.Cluster
			for _, id := range args {
				info, err := client.Cluster(id).Get(context.Background())
				if err != nil {
					return err
				}
				clusters = append(clusters, info)
			}

			// TODO: Print this in a more human-friendly way, and include node summary.
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "    ")
			return encoder.Encode(clusters)
		},
	}
}
