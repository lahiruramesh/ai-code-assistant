package cmd

import (
	"context"
	"log"

	"github.com/lahiruramesh/dock-route/internal/docker"

	"github.com/spf13/cobra"
)

var (
	follow bool
	tail   string
)

var logsCmd = &cobra.Command{
	Use:   "logs [container-name]",
	Short: "Show container logs",
	Long:  `Display logs from a Docker container managed by dock-route.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerName := args[0]

		ctx := context.Background()
		dockerClient, err := docker.NewClient()
		if err != nil {
			log.Fatalf("Failed to create Docker client: %v", err)
		}
		defer dockerClient.Close()

		err = dockerClient.ShowLogs(ctx, containerName, follow, tail)
		if err != nil {
			log.Fatalf("Failed to show logs for container '%s': %v", containerName, err)
		}
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	logsCmd.Flags().StringVarP(&tail, "tail", "t", "100", "Number of lines to show from the end of the logs")
}
