package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"agent/internal/pkg/agents"
	"agent/internal/pkg/llm"
	"agent/server"
)

// getEnvOrDefault returns the environment variable value or a default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	bytes := make([]byte, 8)
	_, err := rand.Read(bytes)
	if err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return "req_" + hex.EncodeToString(bytes)
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Command line flags with environment variable defaults
	var (
		port        = flag.String("port", getEnvOrDefault("SERVER_PORT", "8080"), "HTTP server port")
		projectName = flag.String("project", getEnvOrDefault("DEFAULT_PROJECT_NAME", "my-react-app"), "Default project name")
		projectPath = flag.String("path", getEnvOrDefault("PROJECT_PATH", "./projects"), "Default project path")
		mode        = flag.String("mode", "server", "Run mode: 'cli' or 'server'")
		llmProvider = flag.String("llm", getEnvOrDefault("LLM_PROVIDER", "bedrock"), "LLM provider: 'ollama' or 'bedrock'")
		model       = flag.String("model", getEnvOrDefault("LLM_MODEL", "anthropic.claude-3-5-sonnet-20241022-v2:0"), "Model to use")
	)
	flag.Parse()

	// Parse LLM provider
	var provider llm.LLMProvider
	switch *llmProvider {
	case "ollama":
		provider = llm.OllamaProvider
		if *model == "anthropic.claude-3-5-sonnet-20241022-v2:0" || *model == "deepseek/deepseek-chat-v3-0324:free" {
			*model = "qwen2.5:1.5b" // Default Ollama model
		}
	case "bedrock":
		provider = llm.BedrockProvider
	case "openrouter":
		provider = llm.OpenRouterProvider
	case "gemini":
		provider = llm.GeminiProvider
	case "anthropic":
		provider = llm.AnthropicProvider
	default:
		log.Fatalf("Invalid LLM provider: %s. Use 'ollama', 'bedrock', 'openrouter', 'gemini', or 'anthropic'", *llmProvider)
	}

	log.Printf("Using LLM provider: %s with model: %s", provider, *model)

	// Initialize coordinator
	coordinator, err := agents.NewCoordinator(*projectName, *projectPath, provider, *model)
	if err != nil {
		log.Fatalf("Failed to create coordinator: %v", err)
	}

	// Start agents
	if err := coordinator.Start(); err != nil {
		log.Fatalf("Failed to start coordinator: %v", err)
	}

	if *mode == "server" {
		// Start HTTP server
		log.Printf("Starting Multi-Agent React Builder Server on port %s", *port)

		httpServer := server.NewServer(coordinator, *port, *projectPath)

		// Setup graceful shutdown
		go func() {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			<-sigChan

			log.Println("Shutting down server...")
			coordinator.Stop()
			os.Exit(0)
		}()

		// Start server
		log.Fatal(httpServer.Start())

	} else {
		// CLI mode with loop manager
		log.Println("Starting CLI mode with Agent Loop Manager...")
		log.Println("Type your request and press Enter. Use Ctrl+C to quit.")
		log.Println("Each request will run in its own agent loop with 20-minute timeout.")

		// Use buffered reader for better input handling
		reader := bufio.NewReader(os.Stdin)

		// Start result monitor
		go monitorLoopResults(coordinator.GetLoopManager())

		// Setup graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Handle shutdown in background
		go func() {
			<-sigChan
			log.Println("\nReceived shutdown signal...")

			activeLoops := coordinator.GetLoopManager().GetActiveLoops()
			if len(activeLoops) > 0 {
				log.Printf("Waiting for %d active loops to complete (max 30 seconds)...", len(activeLoops))

				// Wait up to 30 seconds for loops to complete
				timeout := time.After(30 * time.Second)
				ticker := time.NewTicker(2 * time.Second)
				defer ticker.Stop()

				for {
					select {
					case <-timeout:
						log.Println("Timeout reached, forcing shutdown...")
						coordinator.Stop()
						os.Exit(0)
					case <-ticker.C:
						remaining := coordinator.GetLoopManager().GetActiveLoops()
						if len(remaining) == 0 {
							log.Println("All loops completed, shutting down...")
							coordinator.Stop()
							os.Exit(0)
						}
						log.Printf("Still waiting for %d loops...", len(remaining))
					}
				}
			} else {
				coordinator.Stop()
				os.Exit(0)
			}
		}()

		for {
			log.Print("Enter your request: ")
			input, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Printf("Error reading input: %v", err)
				break
			}

			input = strings.TrimSpace(input)
			if input == "quit" || input == "exit" {
				break
			}

			// Skip empty inputs
			if input == "" {
				continue
			}

			// Generate unique request ID
			requestID := generateRequestID()

			log.Printf("Starting agent loop for request %s: %s", requestID, input)

			// Start agent loop
			loop, err := coordinator.ProcessUserRequestWithLoop(requestID, input)
			if err != nil {
				log.Printf("Error starting agent loop: %v", err)
				continue
			}

			log.Printf("Agent loop %s started for request %s", loop.ID, requestID)
			log.Printf("Status: %s | Timeout: 20 minutes", loop.GetStatus())
			log.Println("Loop is running in background. You can enter another request or wait for completion.")
		}

		// Show active loops before shutdown
		activeLoops := coordinator.GetLoopManager().GetActiveLoops()
		if len(activeLoops) > 0 {
			log.Printf("Shutting down with %d active loops...", len(activeLoops))
			for _, loop := range activeLoops {
				log.Printf("- Loop %s (Request: %s) - Status: %s - Duration: %v",
					loop.ID, loop.RequestID, loop.GetStatus(), loop.GetDuration())
			}
		}

		coordinator.Stop()
	}
}

// monitorLoopResults monitors and displays results from completed agent loops
func monitorLoopResults(loopManager *agents.LoopManager) {
	for result := range loopManager.GetResultChannel() {
		log.Printf("\n%s", strings.Repeat("=", 60))
		log.Printf("AGENT LOOP COMPLETED")
		log.Printf("Request ID: %s", result.RequestID)
		log.Printf("Status: %s", result.Status)
		log.Printf("Duration: %v", result.Duration)
		log.Printf("Completed At: %s", result.CompletedAt.Format("15:04:05"))

		if result.Error != nil {
			log.Printf("Error: %v", result.Error)
		}

		if result.Status == agents.RequestCompleted {
			log.Printf("✅ Request completed successfully!")
		} else if result.Status == agents.RequestTimeout {
			log.Printf("⏰ Request timed out after 20 minutes")
		} else if result.Status == agents.RequestFailed {
			log.Printf("❌ Request failed")
		}

		log.Printf("%s\n", strings.Repeat("=", 60))
	}
}
