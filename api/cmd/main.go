package main

import (
	"agent/internal/pkg/agents"
	"agent/internal/pkg/docker"
	"agent/internal/pkg/llm"
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

// CLI represents the command line interface
type CLI struct {
	coordinator   *agents.Coordinator
	dockerService *docker.DockerService
	projectPath   string
	projectName   string
}

// NewCLI creates a new CLI instance
func NewCLI() (*CLI, error) {
	dockerService, err := docker.NewDockerService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Docker service: %v", err)
	}

	return &CLI{
		dockerService: dockerService,
	}, nil
}

// Start begins the CLI application
func (cli *CLI) Start() error {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	fmt.Println("🚀 Multi-Agent React Builder CLI")
	fmt.Println("================================")
	fmt.Println()

	// Get project details from user
	if err := cli.getProjectDetails(); err != nil {
		return fmt.Errorf("failed to get project details: %v", err)
	}

	// Get LLM configuration from environment variables
	llmProvider := os.Getenv("LLM_PROVIDER")
	if llmProvider == "" {
		llmProvider = "ollama" // Default for CLI
	}

	model := os.Getenv("LLM_MODEL")
	if model == "" {
		if llmProvider == "ollama" {
			model = "qwen2.5:1.5b"
		} else {
			model = "anthropic.claude-3-5-sonnet-20241022-v2:0"
		}
	}

	// Parse LLM provider
	var provider llm.LLMProvider
	switch llmProvider {
	case "ollama":
		provider = llm.OllamaProvider
	case "bedrock":
		provider = llm.BedrockProvider
	default:
		return fmt.Errorf("invalid LLM provider: %s. Use 'ollama' or 'bedrock'", llmProvider)
	}

	fmt.Printf("Using LLM provider: %s with model: %s\n", provider, model)

	// Initialize coordinator with environment-based LLM settings
	coordinator, err := agents.NewCoordinator(cli.projectName, cli.projectPath, provider, model)
	if err != nil {
		return fmt.Errorf("failed to initialize coordinator: %v", err)
	}
	cli.coordinator = coordinator

	// Start the multi-agent system
	if err := cli.coordinator.Start(); err != nil {
		return fmt.Errorf("failed to start coordinator: %v", err)
	}

	// Setup graceful shutdown
	cli.setupGracefulShutdown()

	// Start main interaction loop
	return cli.runInteractionLoop()
}

// getProjectDetails gets project information from user
func (cli *CLI) getProjectDetails() error {
	scanner := bufio.NewScanner(os.Stdin)

	// Get project name
	fmt.Print("Enter project name: ")
	if !scanner.Scan() {
		return fmt.Errorf("failed to read project name")
	}
	cli.projectName = strings.TrimSpace(scanner.Text())
	if cli.projectName == "" {
		cli.projectName = "my-react-app"
	}

	// Get project path (default to current directory + project name)
	fmt.Printf("Enter project path (default: ./%s): ", cli.projectName)
	if !scanner.Scan() {
		return fmt.Errorf("failed to read project path")
	}
	projectPath := strings.TrimSpace(scanner.Text())
	if projectPath == "" {
		projectPath = filepath.Join(".", cli.projectName)
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}
	cli.projectPath = absPath

	fmt.Printf("\n✅ Project: %s\n", cli.projectName)
	fmt.Printf("✅ Location: %s\n", cli.projectPath)
	fmt.Println()

	return nil
}

// runInteractionLoop runs the main interaction loop
func (cli *CLI) runInteractionLoop() error {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Welcome to the Multi-Agent React Builder!")
	fmt.Println("Describe what kind of React application you want to build.")
	fmt.Println("Type 'help' for commands, 'status' for project status, or 'quit' to exit.")
	fmt.Println()

	for {
		fmt.Print("👤 You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Handle special commands
		switch strings.ToLower(input) {
		case "quit", "exit":
			fmt.Println("👋 Goodbye!")
			return nil
		case "help":
			cli.showHelp()
			continue
		case "status":
			cli.showStatus()
			continue
		case "deploy":
			cli.deployApplication()
			continue
		case "logs":
			cli.showLogs()
			continue
		}

		// Process user request through agents
		fmt.Println("\n🤖 Processing your request through the multi-agent system...")
		err := cli.coordinator.ProcessUserRequest(input)
		if err != nil {
			log.Printf("Error processing request: %v", err)
			fmt.Printf("❌ Error: %v\n", err)
		}

		// Wait for completion
		err = cli.coordinator.WaitForCompletion(30 * time.Second)
		if err != nil {
			log.Printf("Warning: %v", err)
		}

		fmt.Println("\n" + strings.Repeat("-", 50))
	}

	return nil
}

// showHelp displays available commands
func (cli *CLI) showHelp() {
	fmt.Println("\n📋 Available Commands:")
	fmt.Println("  help     - Show this help message")
	fmt.Println("  status   - Show current project status")
	fmt.Println("  deploy   - Deploy the application using Docker")
	fmt.Println("  logs     - Show application logs")
	fmt.Println("  quit     - Exit the application")
	fmt.Println("\n💡 You can also describe what you want to build:")
	fmt.Println("  'Create a todo list app with React and TypeScript'")
	fmt.Println("  'Add a header component with navigation'")
	fmt.Println("  'Set up the project structure for an e-commerce site'")
	fmt.Println()
}

// showStatus displays current project status
func (cli *CLI) showStatus() {
	if cli.coordinator == nil {
		fmt.Println("❌ No active project")
		return
	}

	status := cli.coordinator.GetProjectStatus()

	fmt.Println("\n📊 Project Status:")
	fmt.Printf("  Name: %v\n", status["project_name"])
	fmt.Printf("  Path: %v\n", status["project_path"])
	fmt.Printf("  Phase: %v\n", status["current_phase"])
	fmt.Printf("  Files: %v\n", status["file_count"])
	fmt.Printf("  Completed Tasks: %v\n", len(status["completed_tasks"].([]string)))
	fmt.Printf("  Active Agents: %v\n", cli.coordinator.ListActiveAgents())
	fmt.Println()
}

// deployApplication deploys the React app using Docker
func (cli *CLI) deployApplication() {
	if cli.coordinator == nil {
		fmt.Println("❌ No active project")
		return
	}

	fmt.Println("\n🐳 Deploying application with Docker...")

	// Check if project exists
	if _, err := os.Stat(cli.projectPath); os.IsNotExist(err) {
		fmt.Println("❌ Project directory doesn't exist. Build the project first.")
		return
	}

	// Send deployment task to Code Editing agent
	err := cli.coordinator.SendAgentMessage(
		"user",
		agents.CodeEditingAgent,
		"deploy_react_app",
		fmt.Sprintf("Deploy React application from %s using Docker", cli.projectPath),
	)

	if err != nil {
		fmt.Printf("❌ Failed to initiate deployment: %v\n", err)
		return
	}

	// Wait for deployment completion
	err = cli.coordinator.WaitForCompletion(60 * time.Second)
	if err != nil {
		fmt.Printf("⚠️  Deployment may still be in progress: %v\n", err)
	}

	fmt.Println("✅ Deployment request sent to DevOps agent")
}

// showLogs displays application logs
func (cli *CLI) showLogs() {
	if cli.dockerService == nil {
		fmt.Println("❌ Docker service not available")
		return
	}

	// List containers
	containers, err := cli.dockerService.ListContainers()
	if err != nil {
		fmt.Printf("❌ Failed to list containers: %v\n", err)
		return
	}

	fmt.Println("\n📋 Docker Containers:")
	for _, container := range containers {
		if len(container.Names) > 0 {
			fmt.Printf("  %s: %s\n", container.Names[0], container.State)

			// Show logs for running containers
			if container.State == "running" {
				logs, err := cli.dockerService.GetContainerLogs(container.ID)
				if err != nil {
					fmt.Printf("    ❌ Failed to get logs: %v\n", err)
				} else {
					// Show last few lines
					lines := strings.Split(logs, "\n")
					startIdx := len(lines) - 6
					if startIdx < 0 {
						startIdx = 0
					}
					for i := startIdx; i < len(lines) && i < len(lines); i++ {
						if lines[i] != "" {
							fmt.Printf("    %s\n", lines[i])
						}
					}
				}
			}
		}
	}
	fmt.Println()
}

// setupGracefulShutdown sets up graceful shutdown handling
func (cli *CLI) setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\n🛑 Shutting down gracefully...")

		if cli.coordinator != nil {
			cli.coordinator.Stop()
		}

		if cli.dockerService != nil {
			cli.dockerService.Close()
		}

		fmt.Println("👋 Goodbye!")
		os.Exit(0)
	}()
}

func main() {
	cli, err := NewCLI()
	if err != nil {
		log.Fatalf("Failed to initialize CLI: %v", err)
	}

	if err := cli.Start(); err != nil {
		log.Fatalf("CLI error: %v", err)
	}
}
