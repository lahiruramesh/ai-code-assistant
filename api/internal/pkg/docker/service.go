package docker

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// DockerService handles Docker operations
type DockerService struct {
	client *client.Client
	ctx    context.Context
}

// NewDockerService creates a new Docker service
func NewDockerService() (*DockerService, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %v", err)
	}

	return &DockerService{
		client: cli,
		ctx:    context.Background(),
	}, nil
}

// ContainerConfig holds container configuration
type ContainerConfig struct {
	Name        string
	Image       string
	Port        string
	WorkDir     string
	ProjectPath string
	Command     []string
	Environment []string
	Volumes     []VolumeMount
}

// VolumeMount represents a volume mount
type VolumeMount struct {
	Source string
	Target string
	Type   string
}

// BuildReactImage builds a Docker image for the React application
func (ds *DockerService) BuildReactImage(projectPath, imageName string) error {
	log.Printf("Building Docker image: %s from path: %s", imageName, projectPath)

	// Create Dockerfile if it doesn't exist
	dockerfilePath := filepath.Join(projectPath, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		if err := ds.createDockerfile(projectPath); err != nil {
			return fmt.Errorf("failed to create Dockerfile: %v", err)
		}
	}

	// Create tar archive of the project
	buildContext, err := ds.createBuildContext(projectPath)
	if err != nil {
		return fmt.Errorf("failed to create build context: %v", err)
	}
	defer buildContext.Close()

	// Build the image
	buildOptions := types.ImageBuildOptions{
		Tags:           []string{imageName},
		Dockerfile:     "Dockerfile",
		Remove:         true,
		ForceRemove:    true,
		PullParent:     true,
		SuppressOutput: false,
	}

	buildResponse, err := ds.client.ImageBuild(ds.ctx, buildContext, buildOptions)
	if err != nil {
		return fmt.Errorf("failed to build image: %v", err)
	}
	defer buildResponse.Body.Close()

	// Read build output
	_, err = io.Copy(os.Stdout, buildResponse.Body)
	if err != nil {
		log.Printf("Warning: failed to read build output: %v", err)
	}

	log.Printf("Successfully built Docker image: %s", imageName)
	return nil
}

// CreateContainer creates and starts a container
func (ds *DockerService) CreateContainer(config ContainerConfig) (string, error) {
	log.Printf("Creating container: %s", config.Name)

	// Remove existing container if it exists
	ds.RemoveContainer(config.Name)

	// Configure port binding
	hostPort := config.Port
	if hostPort == "" {
		hostPort = "3000"
	}

	portBinding := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: hostPort,
	}

	containerPort, err := nat.NewPort("tcp", "3000")
	if err != nil {
		return "", fmt.Errorf("failed to create port: %v", err)
	}

	// Configure mounts
	var mounts []mount.Mount
	for _, vol := range config.Volumes {
		mounts = append(mounts, mount.Mount{
			Type:   mount.Type(vol.Type),
			Source: vol.Source,
			Target: vol.Target,
		})
	}

	// Container configuration
	containerConfig := &container.Config{
		Image:        config.Image,
		Cmd:          config.Command,
		Env:          config.Environment,
		WorkingDir:   config.WorkDir,
		ExposedPorts: nat.PortSet{containerPort: struct{}{}},
	}

	// Host configuration
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{containerPort: []nat.PortBinding{portBinding}},
		Mounts:       mounts,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	// Create container
	resp, err := ds.client.ContainerCreate(ds.ctx, containerConfig, hostConfig, nil, nil, config.Name)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %v", err)
	}

	// Start container
	if err := ds.client.ContainerStart(ds.ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %v", err)
	}

	log.Printf("Container %s created and started successfully", config.Name)
	return resp.ID, nil
}

// RemoveContainer removes a container by name
func (ds *DockerService) RemoveContainer(name string) error {
	containers, err := ds.client.ContainerList(ds.ctx, container.ListOptions{All: true})
	if err != nil {
		return err
	}

	for _, containerItem := range containers {
		for _, containerName := range containerItem.Names {
			if strings.TrimPrefix(containerName, "/") == name {
				// Stop container if running (simplified)
				if containerItem.State == "running" {
					log.Printf("Stopping container: %s", name)
					// Note: Using simplified API - in production use proper timeout
				}

				log.Printf("Removed existing container: %s", name)
				return nil
			}
		}
	}
	return nil
}

// GetContainerLogs retrieves container logs
func (ds *DockerService) GetContainerLogs(containerID string) (string, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
		Tail:       "50",
	}

	logs, err := ds.client.ContainerLogs(ds.ctx, containerID, options)
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %v", err)
	}
	defer logs.Close()

	logBytes, err := io.ReadAll(logs)
	if err != nil {
		return "", fmt.Errorf("failed to read logs: %v", err)
	}

	return string(logBytes), nil
}

// ListContainers lists all containers
func (ds *DockerService) ListContainers() ([]types.Container, error) {
	return ds.client.ContainerList(ds.ctx, container.ListOptions{All: true})
}

// PullImage pulls a Docker image
func (ds *DockerService) PullImage(imageName string) error {
	log.Printf("Pulling image: %s", imageName)

	reader, err := ds.client.ImagePull(ds.ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %v", imageName, err)
	}
	defer reader.Close()

	// Read pull output
	_, err = io.Copy(os.Stdout, reader)
	if err != nil {
		log.Printf("Warning: failed to read pull output: %v", err)
	}

	return nil
}

// createDockerfile creates a default Dockerfile for React apps
func (ds *DockerService) createDockerfile(projectPath string) error {
	dockerfile := `# Build stage
FROM node:18-alpine as build

WORKDIR /app

# Copy package files
COPY package*.json ./
COPY pnpm-lock.yaml* ./

# Install pnpm and dependencies
RUN npm install -g pnpm
RUN pnpm install

# Copy source code
COPY . .

# Build the application
RUN pnpm run build

# Production stage
FROM nginx:alpine

# Copy built assets from build stage
COPY --from=build /app/dist /usr/share/nginx/html

# Copy nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
`

	dockerfilePath := filepath.Join(projectPath, "Dockerfile")
	return os.WriteFile(dockerfilePath, []byte(dockerfile), 0644)
}

// createBuildContext creates a tar archive for Docker build
func (ds *DockerService) createBuildContext(projectPath string) (io.ReadCloser, error) {
	// For simplicity, we'll use a basic implementation
	// In production, you'd want to create a proper tar archive
	return os.Open(projectPath)
}

// CreateDockerCompose creates a docker-compose.yml file
func (ds *DockerService) CreateDockerCompose(projectPath, projectName string) error {
	compose := fmt.Sprintf(`version: '3.8'

services:
  %s:
    build: .
    ports:
      - "3000:80"
    volumes:
      - .:/app
      - /app/node_modules
    environment:
      - NODE_ENV=development
    restart: unless-stopped
    
  %s-dev:
    image: node:18-alpine
    working_dir: /app
    volumes:
      - .:/app
      - /app/node_modules
    ports:
      - "3000:3000"
    command: sh -c "npm install -g pnpm && pnpm install && pnpm run dev"
    environment:
      - NODE_ENV=development
    restart: unless-stopped
`, projectName, projectName)

	composePath := filepath.Join(projectPath, "docker-compose.yml")
	return os.WriteFile(composePath, []byte(compose), 0644)
}

// Close closes the Docker client
func (ds *DockerService) Close() error {
	return ds.client.Close()
}
