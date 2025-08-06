package cmd

import (
    "context"
    "fmt"
    "log"
    "strings"
    
    "github.com/spf13/cobra"
    "github.com/lahiruramesh/dock-route/internal/docker"
)

var executeCmd = &cobra.Command{
    Use:   "exec [container-name] -- [command...]",
    Short: "Execute commands remotely in a running container",
    Long: `Execute any command in a running container. 
For package installations, automatically syncs package.json and package-lock.json back to host.

Examples:
  dock-route exec my-app -- npm install axios
  dock-route exec my-app -- yarn add lodash
  dock-route exec my-app -- npm install --save-dev @types/node
  dock-route exec my-app -- npm run test
  dock-route exec my-app -- ls -la`,
    Args: cobra.MinimumNArgs(1), // Changed: Only require container name
    RunE: runExecute,
    // Add this to disable flag parsing for command arguments
    DisableFlagParsing: false, // Keep this false, we'll handle it differently
}

var (
    workingDir  string
    interactive bool
    syncFiles   bool
)

func init() {
    rootCmd.AddCommand(executeCmd)
    
    executeCmd.Flags().StringVarP(&workingDir, "workdir", "w", "/app", "Working directory in container")
    executeCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Run in interactive mode")
    executeCmd.Flags().BoolVarP(&syncFiles, "sync", "s", true, "Auto-sync package files for install commands")
    
    // This is the key change - mark flags as parsed before command args
    executeCmd.Flags().SetInterspersed(false)
}

func runExecute(cmd *cobra.Command, args []string) error {
    if len(args) < 1 {
        return fmt.Errorf("container name is required")
    }
    
    containerName := args[0]
    
    // Handle command parsing - look for '--' separator
    var command []string
    if len(args) > 1 {
        if args[1] == "--" && len(args) > 2 {
            // Use everything after '--' as the command
            command = args[2:]
        } else if len(args) > 1 {
            // Use everything after container name as command
            command = args[1:]
        }
    }
    
    if len(command) == 0 {
        return fmt.Errorf("command is required. Use: exec %s -- [command]", containerName)
    }
    
    ctx := context.Background()
    
    log.Printf("Executing in container '%s': %s", containerName, strings.Join(command, " "))
    
    // Initialize Docker client
    dockerClient, err := docker.NewClient()
    if err != nil {
        return fmt.Errorf("failed to create Docker client: %w", err)
    }
    defer dockerClient.Close()
    
    // Execute command in container
    exitCode, err := dockerClient.ExecuteCommand(ctx, containerName, command, workingDir, interactive)
    if err != nil {
        return fmt.Errorf("failed to execute command: %w", err)
    }
    
    if exitCode != 0 {
        log.Printf("⚠️  Command exited with code: %d", exitCode)
    } else {
        log.Printf("✅ Command executed successfully")
    }
    
    // Auto-sync package files if this was a package installation command
    if syncFiles && isPackageInstallCommand(command) {
        log.Printf("📦 Detected package installation, syncing files...")
        if err := dockerClient.SyncPackageFiles(ctx, containerName, "./"); err != nil {
            log.Printf("⚠️  Warning: Failed to sync package files: %v", err)
        } else {
            log.Printf("✅ Package files synced to host")
        }
    }
    
    return nil
}

// isPackageInstallCommand checks if the command is a package installation
func isPackageInstallCommand(command []string) bool {
    if len(command) == 0 {
        return false
    }
    
    packageCommands := [][]string{
        {"npm", "install"},
        {"npm", "i"},
        {"yarn", "add"},
        {"pnpm", "install"},
        {"pnpm", "add"},
    }
    
    for _, pkgCmd := range packageCommands {
        if len(command) >= len(pkgCmd) {
            match := true
            for i, part := range pkgCmd {
                if command[i] != part {
                    match = false
                    break
                }
            }
            if match {
                return true
            }
        }
    }
    
    return false
}
