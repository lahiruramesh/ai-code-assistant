package docker

import (
	"archive/tar"
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/lahiruramesh/dock-route/internal/config"
)

type Client struct {
	cli *client.Client
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &Client{cli: cli}, nil
}

func (c *Client) Close() error {
	return c.cli.Close()
}

func (c *Client) DeployContainer(ctx context.Context, config *config.DeployConfig) (string, error) {
	// Build Docker image
	if err := c.buildImage(ctx, config); err != nil {
		return "", fmt.Errorf("failed to build image: %w", err)
	}

	// Start container
	containerIP, err := c.startContainer(ctx, config)
	if err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	return containerIP, nil
}

func (c *Client) buildImage(ctx context.Context, config *config.DeployConfig) error {
	log.Printf("Building Docker image '%s'...", config.ImageName)

	// Create build context with Dockerfile
	buildCtxReader, cleanup, err := c.createBuildContext(config)
	if err != nil {
		return err
	}
	defer cleanup()

	buildOptions := types.ImageBuildOptions{
		Tags:       []string{config.ImageName},
		Dockerfile: "Dockerfile",
		Remove:     true,
		BuildArgs:  c.convertBuildArgs(config.Template.BuildArgs), // Convert to *string map
	}

	buildResponse, err := c.cli.ImageBuild(ctx, buildCtxReader, buildOptions)
	if err != nil {
		return err
	}
	defer buildResponse.Body.Close()

	// Stream build output and check for errors
	scanner := bufio.NewScanner(buildResponse.Body)
	buildSuccess := true
	var buildError string

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)

		// Check for error indicators in the build output
		if strings.Contains(line, `"errorDetail"`) ||
			strings.Contains(line, `"error"`) ||
			strings.Contains(line, "returned a non-zero code") {
			buildSuccess = false
			buildError = line
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read build output: %w", err)
	}

	if !buildSuccess {
		return fmt.Errorf("docker build failed: %s", buildError)
	}

	log.Printf("Docker image '%s' built successfully.", config.ImageName)
	return nil
}

// Convert map[string]string to map[string]*string for Docker API
func (c *Client) convertBuildArgs(buildArgs map[string]string) map[string]*string {
	converted := make(map[string]*string)
	for key, value := range buildArgs {
		val := value // Create a copy to get the address
		converted[key] = &val
	}
	return converted
}

func (c *Client) createBuildContext(config *config.DeployConfig) (io.Reader, func(), error) {
	pr, pw := io.Pipe()
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer pw.Close()

		tw := tar.NewWriter(pw)
		defer tw.Close()

		// Add Dockerfile
		dockerfileHeader := &tar.Header{
			Name:   "Dockerfile",
			Mode:   0644,
			Size:   int64(len(config.Template.Dockerfile)),
			Format: tar.FormatPAX, // Use PAX format for long paths
		}

		if err := tw.WriteHeader(dockerfileHeader); err != nil {
			pw.CloseWithError(err)
			return
		}

		if _, err := tw.Write([]byte(config.Template.Dockerfile)); err != nil {
			pw.CloseWithError(err)
			return
		}

		// Add source files with exclusions
		err := filepath.Walk(config.SourcePath, func(file string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(config.SourcePath, file)
			if err != nil {
				return err
			}

			if relPath == "." {
				return nil
			}

			// Skip excluded files and directories
			if c.shouldExclude(relPath) {
				if fi.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Handle long paths using PAX format
			header, err := tar.FileInfoHeader(fi, relPath)
			if err != nil {
				return err
			}

			// Clean and validate the path
			cleanPath := filepath.ToSlash(relPath)
			if len(cleanPath) > 100 {
				header.Format = tar.FormatPAX // Use PAX format for long paths
			}
			header.Name = cleanPath

			if err := tw.WriteHeader(header); err != nil {
				return fmt.Errorf("failed to write tar header for %s: %w", cleanPath, err)
			}

			if !fi.IsDir() && fi.Size() > 0 {
				srcFile, err := os.Open(file)
				if err != nil {
					return fmt.Errorf("failed to open file %s: %w", file, err)
				}
				defer srcFile.Close()

				_, err = io.Copy(tw, srcFile)
				if err != nil {
					return fmt.Errorf("failed to copy file %s to tar: %w", file, err)
				}
			}

			return nil
		})

		if err != nil {
			pw.CloseWithError(err)
		}
	}()

	cleanup := func() {
		wg.Wait()
	}

	return pr, cleanup, nil
}

// shouldExclude determines if a file/directory should be excluded from the build context
func (c *Client) shouldExclude(relPath string) bool {
	excludePatterns := []string{
		"node_modules",
		".git",
		".gitignore",
		".dockerignore",
		".next",
		"dist",
		"build",
		".vscode",
		".idea",
		"*.log",
		".env",
		".env.local",
		".env.development.local",
		".env.test.local",
		".env.production.local",
		"coverage",
		".nyc_output",
		".cache",
		"tmp",
		"temp",
	}

	for _, pattern := range excludePatterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(relPath)); matched {
			return true
		}
		// Check if any parent directory matches the pattern
		parts := strings.Split(relPath, string(filepath.Separator))
		for _, part := range parts {
			if matched, _ := filepath.Match(pattern, part); matched {
				return true
			}
		}
	}

	return false
}

func (c *Client) startContainer(ctx context.Context, config *config.DeployConfig) (string, error) {
	// Remove existing container if it exists
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", config.ContainerName)),
	})
	if err != nil {
		return "", err
	}

	if len(containers) > 0 {
		log.Printf("Removing existing container '%s'...", config.ContainerName)
		if err := c.cli.ContainerRemove(ctx, containers[0].ID, container.RemoveOptions{Force: true}); err != nil {
			return "", err
		}
	}

	exposedPorts := nat.PortSet{nat.Port(config.Template.Port + "/tcp"): struct{}{}}

	// Prepare container command based on mode
	var cmd []string
	if config.DevMode && len(config.Template.DevCommand) > 0 {
		cmd = config.Template.DevCommand
	} else if len(config.Template.ProdCommand) > 0 {
		cmd = config.Template.ProdCommand
	}

	containerConfig := &container.Config{
		Image:        config.ImageName,
		ExposedPorts: exposedPorts,
		Env:          c.buildEnvVars(config.Template.Environment),
		Labels: map[string]string{
			"managed-by": "dock-route",
			"mode":       c.getMode(config.DevMode),
		},
		WorkingDir: config.Template.MountPath,
	}

	// Set command if specified
	if len(cmd) > 0 {
		containerConfig.Cmd = cmd
	}

	// Configure host config with proper mount options for development
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(config.Template.Port + "/tcp"): []nat.PortBinding{
				{HostIP: "0.0.0.0", HostPort: config.HostPort},
			},
		},
	}

	// Add bind mount for live editing
	if config.DevMode {
		hostConfig.Mounts = []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: config.SourcePath,
				Target: config.Template.MountPath,
				BindOptions: &mount.BindOptions{
					Propagation: mount.PropagationRPrivate,
				},
			},
			// Mount node_modules as a volume to avoid conflicts
			{
				Type:   mount.TypeVolume,
				Source: fmt.Sprintf("%s-node_modules", config.ContainerName),
				Target: filepath.Join(config.Template.MountPath, "node_modules"),
			},
		}
	}

	resp, err := c.cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, config.ContainerName)
	if err != nil {
		return "", err
	}

	if err := c.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	containerJSON, err := c.cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return "", err
	}

	var containerIP string
	if bridgeNet, ok := containerJSON.NetworkSettings.Networks["bridge"]; ok {
		containerIP = bridgeNet.IPAddress
	}

	if config.DevMode {
		log.Printf("🚀 Container '%s' started in DEVELOPMENT mode", config.ContainerName)
		log.Printf("📁 Live editing enabled for: %s", config.SourcePath)
	} else {
		log.Printf("Container '%s' started in production mode", config.ContainerName)
	}

	return containerIP, nil
}

func (c *Client) getMode(devMode bool) string {
	if devMode {
		return "development"
	}
	return "production"
}

func (c *Client) buildEnvVars(envMap map[string]string) []string {
	var envVars []string
	for key, value := range envMap {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}
	return envVars
}

func (c *Client) ExecuteCommand(ctx context.Context, containerName string, command []string, workingDir string, interactive bool) (int, error) {
	// Find the container
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", containerName)),
	})
	if err != nil {
		return -1, fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		return -1, fmt.Errorf("container '%s' not found or not running", containerName)
	}

	containerID := containers[0].ID

	// Create exec configuration
	execConfig := container.ExecOptions{
		Cmd:          command,
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   workingDir,
	}

	if interactive {
		execConfig.AttachStdin = true
		execConfig.Tty = true
	}

	// Create exec instance
	execResp, err := c.cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return -1, fmt.Errorf("failed to create exec: %w", err)
	}

	// Attach to exec
	attachResp, err := c.cli.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{
		Tty: interactive,
	})
	if err != nil {
		return -1, fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer attachResp.Close()

	// Start the exec
	if err := c.cli.ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{
		Tty: interactive,
	}); err != nil {
		return -1, fmt.Errorf("failed to start exec: %w", err)
	}

	// Handle output streaming
	if interactive {
		// For interactive mode, copy stdin/stdout directly
		go func() {
			io.Copy(attachResp.Conn, os.Stdin)
		}()
		io.Copy(os.Stdout, attachResp.Reader)
	} else {
		// For non-interactive, stream output with prefixes
		c.streamOutput(attachResp.Reader)
	}

	// Wait for completion and get exit code
	inspectResp, err := c.cli.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return -1, fmt.Errorf("failed to inspect exec: %w", err)
	}

	// Wait for exec to complete
	for inspectResp.Running {
		time.Sleep(100 * time.Millisecond)
		inspectResp, err = c.cli.ContainerExecInspect(ctx, execResp.ID)
		if err != nil {
			return -1, fmt.Errorf("failed to inspect exec: %w", err)
		}
	}

	return inspectResp.ExitCode, nil
}

func (c *Client) streamOutput(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		// Remove Docker's stream header if present
		if len(line) > 8 {
			fmt.Println(line)
		} else if len(line) > 0 {
			fmt.Println(line)
		}
	}
}

func (c *Client) SyncPackageFiles(ctx context.Context, containerName string, hostPath string) error {
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", containerName)),
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		return fmt.Errorf("container '%s' not found", containerName)
	}

	containerID := containers[0].ID

	// Files to sync from container to host
	packageFiles := []struct {
		containerPath string
		hostFileName  string
		required      bool
	}{
		{"/app/package.json", "package.json", true},
		{"/app/package-lock.json", "package-lock.json", false},
		{"/app/yarn.lock", "yarn.lock", false},
		{"/app/pnpm-lock.yaml", "pnpm-lock.yaml", false},
	}

	syncedFiles := 0

	for _, file := range packageFiles {
		hostFilePath := filepath.Join(hostPath, file.hostFileName)

		err := c.copyFileFromContainer(ctx, containerID, file.containerPath, hostFilePath)
		if err != nil {
			if file.required {
				log.Printf("⚠️  Failed to sync required file %s: %v", file.hostFileName, err)
			} else {
				log.Printf("ℹ️  Optional file %s not found (this is normal)", file.hostFileName)
			}
		} else {
			log.Printf("📄 Synced %s", file.hostFileName)
			syncedFiles++
		}
	}

	if syncedFiles == 0 {
		return fmt.Errorf("no package files were synced")
	}

	return nil
}

func (c *Client) copyFileFromContainer(ctx context.Context, containerID, srcPath, dstPath string) error {
	reader, _, err := c.cli.CopyFromContainer(ctx, containerID, srcPath)
	if err != nil {
		return fmt.Errorf("failed to copy from container: %w", err)
	}
	defer reader.Close()

	// Extract file from tar archive
	tr := tar.NewReader(reader)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		if header.Typeflag == tar.TypeReg {
			// Create host file
			outFile, err := os.Create(dstPath)
			if err != nil {
				return fmt.Errorf("failed to create host file: %w", err)
			}
			defer outFile.Close()

			// Copy content
			_, err = io.Copy(outFile, tr)
			if err != nil {
				return fmt.Errorf("failed to write file content: %w", err)
			}

			log.Printf("✅ Copied %s from container to %s", srcPath, dstPath)
			return nil
		}
	}

	return fmt.Errorf("file not found in tar archive")
}

// Helper method to get container info (you might need this for other commands)
func (c *Client) GetContainerInfo(ctx context.Context, containerName string) (*container.Summary, error) {
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", containerName)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		return nil, fmt.Errorf("container '%s' not found", containerName)
	}

	return &containers[0], nil
}

// StartContainer starts a stopped container
func (c *Client) StartContainer(ctx context.Context, containerName string) error {
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", containerName)),
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		return fmt.Errorf("container '%s' not found", containerName)
	}

	containerInfo := containers[0]

	if containerInfo.State == "running" {
		return fmt.Errorf("container '%s' is already running", containerName)
	}

	err = c.cli.ContainerStart(ctx, containerInfo.ID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	return nil
}

// StopContainer stops a running container
func (c *Client) StopContainer(ctx context.Context, containerName string) error {
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", containerName)),
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		return fmt.Errorf("container '%s' not found", containerName)
	}

	containerInfo := containers[0]

	if containerInfo.State != "running" {
		return fmt.Errorf("container '%s' is not running", containerName)
	}

	timeout := 10 // seconds
	err = c.cli.ContainerStop(ctx, containerInfo.ID, container.StopOptions{
		Timeout: &timeout,
	})
	if err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	return nil
}

// ShowLogs displays container logs
func (c *Client) ShowLogs(ctx context.Context, containerName string, follow bool, tail string) error {
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", containerName)),
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		return fmt.Errorf("container '%s' not found", containerName)
	}

	containerInfo := containers[0]

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       tail,
		Timestamps: true,
	}

	logs, err := c.cli.ContainerLogs(ctx, containerInfo.ID, options)
	if err != nil {
		return fmt.Errorf("failed to get container logs: %w", err)
	}
	defer logs.Close()

	// Stream logs to stdout
	_, err = io.Copy(os.Stdout, logs)
	if err != nil {
		return fmt.Errorf("failed to stream logs: %w", err)
	}

	return nil
}
