package main

import (
	"agent/internal/pkg/agents"
	"agent/internal/pkg/llm"
	"agent/server"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get port from environment or use default
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = os.Getenv("PORT")
		if port == "" {
			port = "8084"
		}
	}

	// Set project paths from environment
	projectPath := os.Getenv("PROJECT_PATH")
	if projectPath == "" {
		projectPath = "/tmp/projects"
	}

	aiagentPath := os.Getenv("AIAGENT_PATH")
	if aiagentPath == "" {
		aiagentPath = "/tmp/aiagent"
	}

	// Get LLM provider and model from environment
	provider := os.Getenv("LLM_PROVIDER")
	if provider == "" {
		provider = "ollama"
	}

	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = "qwen2.5:1.5b"
	}

	// Map provider string to provider type
	var providerType llm.LLMProvider
	switch provider {
	case "openrouter":
		providerType = llm.OpenRouterProvider
	case "gemini":
		providerType = llm.GeminiProvider
	case "anthropic":
		providerType = llm.AnthropicProvider
	case "bedrock":
		providerType = llm.BedrockProvider
	case "ollama":
		providerType = llm.OllamaProvider
	default:
		log.Printf("Unknown provider %s, defaulting to ollama", provider)
		providerType = llm.OllamaProvider
		model = "qwen2.5:1.5b"
	}

	log.Printf("🤖 Using LLM Provider: %s with model: %s", provider, model)

	// Initialize coordinator for the server
	coordinator, err := agents.NewCoordinator("server", aiagentPath, providerType, model)
	if err != nil {
		log.Fatalf("Failed to initialize coordinator: %v", err)
	}

	// Create HTTP server
	httpServer := server.NewServer(coordinator, port, projectPath)

	log.Printf("🚀 Starting HTTP server on port %s", port)
	log.Printf("📁 Project path: %s", projectPath)
	log.Printf("🤖 AI Agent path: %s", aiagentPath)
	log.Printf("🌐 Access URL: http://localhost:%s", port)

	if err := httpServer.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
