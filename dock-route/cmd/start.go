package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/lahiruramesh/dock-route/internal/docker"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [container-name]",
	Short: "Start a stopped container",
	Long:  `Start a previously stopped Docker container managed by dock-route.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerName := args[0]

		ctx := context.Background()
		dockerClient, err := docker.NewClient()
		if err != nil {
			log.Fatalf("Failed to create Docker client: %v", err)
		}
		defer dockerClient.Close()

		err = dockerClient.StartContainer(ctx, containerName)
		if err != nil {
			log.Fatalf("Failed to start container '%s': %v", containerName, err)
		}

		fmt.Printf("Container '%s' started successfully.\n", containerName)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
