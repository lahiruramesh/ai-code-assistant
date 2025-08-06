package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/lahiruramesh/dock-route/internal/docker"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [container-name]",
	Short: "Remove a deployed container",
	Long:  `Remove a deployed container and clean up associated resources.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

var (
	forceRemove bool
	removeImage bool
)

func init() {
	rootCmd.AddCommand(removeCmd)

	removeCmd.Flags().BoolVarP(&forceRemove, "force", "f", false, "Force remove running container")
	removeCmd.Flags().BoolVar(&removeImage, "remove-image", false, "Also remove the associated Docker image")
}

func runRemove(cmd *cobra.Command, args []string) error {
	containerName := args[0]
	ctx := context.Background()

	dockerClient, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer dockerClient.Close()

	log.Printf("Removing container: %s", containerName)

	imageName, err := dockerClient.RemoveContainer(ctx, containerName, forceRemove)
	if err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	log.Printf("Container '%s' removed successfully", containerName)

	if removeImage && imageName != "" {
		log.Printf("Removing associated image: %s", imageName)
		if err := dockerClient.RemoveImage(ctx, imageName); err != nil {
			log.Printf("Warning: failed to remove image %s: %v", imageName, err)
		} else {
			log.Printf("Image '%s' removed successfully", imageName)
		}
	}

	fmt.Printf("Deployment '%s' has been removed.\n", containerName)
	fmt.Printf("Subdomain 'preview-%s.domain.localhost' is no longer accessible.\n", containerName)

	return nil
}
