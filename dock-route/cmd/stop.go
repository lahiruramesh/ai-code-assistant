package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/lahiruramesh/dock-route/internal/docker"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [container-name]",
	Short: "Stop a running container",
	Long:  `Stop a running Docker container managed by dock-route.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerName := args[0]

		ctx := context.Background()
		dockerClient, err := docker.NewClient()
		if err != nil {
			log.Fatalf("Failed to create Docker client: %v", err)
		}
		defer dockerClient.Close()

		err = dockerClient.StopContainer(ctx, containerName)
		if err != nil {
			log.Fatalf("Failed to stop container '%s': %v", containerName, err)
		}

		fmt.Printf("Container '%s' stopped successfully.\n", containerName)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
