package cmd

import (
	"context"
	"fmt"

	"github.com/lahiruramesh/dock-route/internal/docker"
	"github.com/lahiruramesh/dock-route/internal/templates"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [templates|containers]",
	Short: "List available templates or running containers",
	Long:  `List available application templates or currently running containers managed by dock-route.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	switch args[0] {
	case "templates":
		return listTemplates()
	case "containers":
		return listContainers()
	default:
		return fmt.Errorf("invalid list type: %s. Use 'templates' or 'containers'", args[0])
	}
}

func listTemplates() error {
	templateManager := templates.NewManager()
	availableTemplates := templateManager.ListTemplates()

	if len(availableTemplates) == 0 {
		fmt.Println("No templates available.")
		return nil
	}

	fmt.Println("Available Templates:")
	fmt.Println("===================")

	for _, templateType := range availableTemplates {
		template, err := templateManager.GetTemplate(templateType)
		if err != nil {
			fmt.Printf("- %s (error loading details)\n", templateType)
			continue
		}

		fmt.Printf("- **%s**: %s\n", template.Name, template.Description)
		fmt.Printf("  Port: %s, Mount: %s\n", template.Port, template.MountPath)
		fmt.Println()
	}

	return nil
}

func listContainers() error {
	ctx := context.Background()
	dockerClient, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer dockerClient.Close()

	containers, err := dockerClient.ListManagedContainers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		fmt.Println("No managed containers found.")
		return nil
	}

	fmt.Println("Managed Containers:")
	fmt.Println("==================")

	for _, container := range containers {
		fmt.Printf("- **%s**\n", container.Name)
		fmt.Printf("  Image: %s\n", container.Image)
		fmt.Printf("  Status: %s\n", container.Status)
		fmt.Printf("  Ports: %s\n", container.Ports)
		fmt.Printf("  Subdomain: preview-%s.dock-route.local\n", container.Name)
		fmt.Println()
	}

	return nil
}
