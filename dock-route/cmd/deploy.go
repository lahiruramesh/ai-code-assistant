package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/lahiruramesh/dock-route/internal/config"
	"github.com/lahiruramesh/dock-route/internal/docker"
	"github.com/lahiruramesh/dock-route/internal/proxy"
	"github.com/lahiruramesh/dock-route/internal/templates"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [app-type] [container-name] [source-path]",
	Short: "Deploy an application with automatic subdomain routing",
	Long: `Deploy an application using a specified template (nextjs, reactjs, nodejs)
with a custom container name and automatic subdomain generation.

Example:
  dock-route deploy nextjs my-next-app ./my-next-project`,
	Args: cobra.ExactArgs(3),
	RunE: runDeploy,
}

var (
    imageName    string
    hostPort     string
    startProxy   bool
    devMode      bool  // Add development mode flag
)

func init() {
    rootCmd.AddCommand(deployCmd)
    
    deployCmd.Flags().StringVarP(&imageName, "image", "i", "", "Custom image name (default: auto-generated)")
    deployCmd.Flags().StringVar(&hostPort, "host-port", "8081", "Host port to bind container port")
    deployCmd.Flags().BoolVar(&startProxy, "start-proxy", true, "Start the reverse proxy server")
    deployCmd.Flags().BoolVar(&devMode, "dev", true, "Enable development mode with live editing") // Add this
}

func runDeploy(cmd *cobra.Command, args []string) error {
    appType := args[0]
    containerName := args[1]
    sourcePath := args[2]
    
    ctx := context.Background()
    
    // Load application template
    templateManager := templates.NewManager()
    template, err := templateManager.GetTemplate(appType)
    if err != nil {
        return fmt.Errorf("failed to load template for %s: %w", appType, err)
    }
    
    // Generate image name if not provided
    if imageName == "" {
        mode := "prod"
        if devMode {
            mode = "dev"
        }
        imageName = fmt.Sprintf("%s-%s-%s:latest", appType, containerName, mode)
    }
    
    // Initialize Docker client
    dockerClient, err := docker.NewClient()
    if err != nil {
        return fmt.Errorf("failed to create Docker client: %w", err)
    }
    defer dockerClient.Close()
    
    // Build and deploy container
    deployConfig := &config.DeployConfig{
        AppType:       appType,
        ContainerName: containerName,
        ImageName:     imageName,
        SourcePath:    sourcePath,
        HostPort:      hostPort,
        Template:      template,
        DevMode:       devMode, // Add this
    }
    
    containerIP, err := dockerClient.DeployContainer(ctx, deployConfig)
    if err != nil {
        return fmt.Errorf("failed to deploy container: %w", err)
    }
    
    // Generate subdomain
    subdomain := fmt.Sprintf("preview-%s", containerName)
    domain := viper.GetString("domain")
    fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)
    
    log.Printf("Container deployed successfully!")
    log.Printf("Container: %s", containerName)
    log.Printf("Image: %s", imageName)
    log.Printf("Subdomain: %s", fullDomain)
    
    if devMode {
        log.Printf("🔥 Development mode enabled - Live editing active!")
        log.Printf("📁 Watching files in: %s", sourcePath)
    }
    
    if startProxy {
        return startProxyServer(subdomain, containerIP, template.Port)
    }
    
    return nil
}

func startProxyServer(subdomain, containerIP, containerPort string) error {
	pm := proxy.NewManager()

	targetURL := fmt.Sprintf("http://localhost:%s", hostPort)
	if err := pm.AddProxy(subdomain, targetURL); err != nil {
		return fmt.Errorf("failed to add proxy: %w", err)
	}

	port := viper.GetString("port")
	domain := viper.GetString("domain")

	log.Printf("Starting reverse proxy server on :%s", port)
	log.Printf("Access your application at: %s.%s:%s", subdomain, domain, port)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           pm,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       15 * time.Second,
	}
    server.ListenAndServe()
	return nil
}
