package docker

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
)

func (c *Client) RemoveImage(ctx context.Context, imageName string) error {
	_, err := c.cli.ImageRemove(ctx, imageName, image.RemoveOptions{
		Force:         false,
		PruneChildren: true,
	})
	if err != nil {
		return fmt.Errorf("failed to remove image %s: %w", imageName, err)
	}

	return nil
}

func (c *Client) ImageExists(ctx context.Context, imageName string) (bool, error) {
	images, err := c.cli.ImageList(ctx, image.ListOptions{
		Filters: filters.NewArgs(filters.Arg("reference", imageName)),
	})
	if err != nil {
		return false, fmt.Errorf("failed to check if image exists: %w", err)
	}

	return len(images) > 0, nil
}

func (c *Client) PullImage(ctx context.Context, imageName string) error {
	reader, err := c.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()

	// Stream the pull output to stdout for visibility
	_, err = io.Copy(os.Stdout, reader)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read image pull output: %w", err)
	}

	return nil
}

func (c *Client) ListImages(ctx context.Context) ([]image.Summary, error) {
	images, err := c.cli.ImageList(ctx, image.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", "built-by=dock-route"),
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	return images, nil
}

func (c *Client) PruneImages(ctx context.Context) error {
	_, err := c.cli.ImagesPrune(ctx, filters.NewArgs(
		filters.Arg("dangling", "true"),
	))
	if err != nil {
		return fmt.Errorf("failed to prune images: %w", err)
	}

	return nil
}
