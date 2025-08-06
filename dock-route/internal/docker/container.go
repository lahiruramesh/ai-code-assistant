package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
)

type ContainerInfo struct {
	ID     string
	Name   string
	Image  string
	Status string
	Ports  string
}

func (c *Client) ListManagedContainers(ctx context.Context) ([]ContainerInfo, error) {
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", "managed-by=dock-route"),
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var result []ContainerInfo

	for _, container := range containers {
		name := strings.TrimPrefix(container.Names[0], "/")

		// Format port information
		var ports []string
		for _, port := range container.Ports {
			if port.PublicPort != 0 {
				ports = append(ports, fmt.Sprintf("%d:%d", port.PublicPort, port.PrivatePort))
			}
		}
		portStr := strings.Join(ports, ", ")

		result = append(result, ContainerInfo{
			ID:     container.ID[:12], // Short ID
			Name:   name,
			Image:  container.Image,
			Status: container.Status,
			Ports:  portStr,
		})
	}

	return result, nil
}

func (c *Client) RemoveContainer(ctx context.Context, containerName string, force bool) (string, error) {
	// Find the container
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", containerName)),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		return "", fmt.Errorf("container '%s' not found", containerName)
	}

	containerInfo := containers[0]
	imageName := containerInfo.Image

	// Remove the container
	err = c.cli.ContainerRemove(ctx, containerInfo.ID, container.RemoveOptions{
		Force: force,
	})
	if err != nil {
		return "", fmt.Errorf("failed to remove container: %w", err)
	}

	return imageName, nil
}

func (c *Client) GetContainerStatus(ctx context.Context, containerName string) (string, error) {
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", containerName)),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		return "not found", nil
	}

	return containers[0].Status, nil
}
