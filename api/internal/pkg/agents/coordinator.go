package agents

import (
	"agent/internal/pkg/llm"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ollama/ollama/api"
)

// Coordinator manages all agents and their communication
type Coordinator struct {
	agents        map[AgentType]*Agent
	messageRouter chan AgentMessage
	context       *ProjectContext
	wg            sync.WaitGroup
	active        bool
	mutex         sync.RWMutex
	llmService    *llm.LLMService
	loopManager   *LoopManager
}

// NewCoordinator creates a new coordinator
func NewCoordinator(projectName, projectPath string, llmProvider llm.LLMProvider, model string) (*Coordinator, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Printf("Warning: Failed to create Ollama client: %v", err)
		// Continue without Ollama client
	}

	// Create LLM service
	llmService, err := llm.NewLLMService(llmProvider, model)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM service: %v", err)
	}

	// Initialize project context
	projectContext := &ProjectContext{
		ProjectName:    projectName,
		ProjectPath:    projectPath,
		Requirements:   "",
		CurrentPhase:   "initialization",
		CompletedTasks: []string{},
		ActiveTasks:    []string{},
		ProjectFiles:   make(map[string]string),
	}

	coordinator := &Coordinator{
		agents:        make(map[AgentType]*Agent),
		messageRouter: make(chan AgentMessage, 1000),
		context:       projectContext,
		active:        false,
		llmService:    llmService,
	}

	// Create all agents (removed DevOpsAgent)
	agentTypes := []AgentType{SupervisorAgent, CodeEditingAgent, ReactAgent}
	for _, agentType := range agentTypes {
		agent, err := NewAgent(agentType, client, llmService, projectContext)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s agent: %v", agentType, err)
		}
		coordinator.agents[agentType] = agent
	}

	// Initialize loop manager
	coordinator.loopManager = NewLoopManager(coordinator)

	return coordinator, nil
}

// Start begins the coordinator and all agents
func (c *Coordinator) Start() error {
	c.mutex.Lock()
	c.active = true
	c.mutex.Unlock()

	log.Println("Starting Multi-Agent System...")

	// Start message router
	go c.routeMessages()

	// Start all agents
	for agentType, agent := range c.agents {
		c.wg.Add(1)
		go agent.Start(&c.wg)
		log.Printf("Started %s agent", agentType)
	}

	// Start outbox monitors for each agent
	for agentType, agent := range c.agents {
		go c.monitorAgentOutbox(agentType, agent)
	}

	log.Println("All agents started successfully")
	return nil
}

// Stop gracefully stops all agents
func (c *Coordinator) Stop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.active {
		return
	}

	c.active = false
	log.Println("Stopping Multi-Agent System...")

	// Stop all agents
	for agentType, agent := range c.agents {
		agent.Stop()
		log.Printf("Stopped %s agent", agentType)
	}

	// Close message router
	close(c.messageRouter)

	// Wait for all agents to finish
	c.wg.Wait()

	// Stop loop manager
	if c.loopManager != nil {
		c.loopManager.Stop()
	}

	log.Println("All agents stopped")
}

// ProcessUserRequestWithLoop processes a user request using the loop manager
func (c *Coordinator) ProcessUserRequestWithLoop(requestID, request string) (*AgentLoop, error) {
	c.mutex.RLock()
	active := c.active
	c.mutex.RUnlock()

	if !active {
		return nil, fmt.Errorf("coordinator is not active")
	}

	log.Printf("Starting agent loop for request %s: %s", requestID, request)
	return c.loopManager.StartLoop(requestID, request)
}

// GetLoopManager returns the loop manager
func (c *Coordinator) GetLoopManager() *LoopManager {
	return c.loopManager
}

// ProcessUserRequest processes a user request through the supervisor (legacy method)
func (c *Coordinator) ProcessUserRequest(request string) error {
	c.mutex.RLock()
	active := c.active
	c.mutex.RUnlock()

	if !active {
		return fmt.Errorf("coordinator is not active")
	}

	log.Printf("Processing user request: %s", request)

	// Update context with user requirements
	c.context.UpdateProjectContext(map[string]interface{}{
		"requirements":  request,
		"current_phase": "planning",
	})

	// Send request to supervisor agent
	msg := AgentMessage{
		ID:        generateID(),
		FromAgent: "user",
		ToAgent:   SupervisorAgent,
		TaskType:  "project_request",
		Content:   request,
		Status:    TaskPending,
		Timestamp: getCurrentTimestamp(),
	}

	return c.sendMessage(msg)
}

// monitorAgentOutbox monitors an agent's outbox and routes messages
func (c *Coordinator) monitorAgentOutbox(agentType AgentType, agent *Agent) {
	log.Printf("Starting outbox monitor for %s agent", agentType)

	for msg := range agent.Outbox {
		if !c.active {
			break
		}

		log.Printf("Agent %s sent message to %s: %s", msg.FromAgent, msg.ToAgent, msg.TaskType)
		c.messageRouter <- msg
	}

	log.Printf("Outbox monitor for %s agent stopped", agentType)
}

// sendMessage sends a message through the system
func (c *Coordinator) sendMessage(msg AgentMessage) error {
	c.mutex.RLock()
	active := c.active
	c.mutex.RUnlock()

	if !active {
		return fmt.Errorf("coordinator is not active")
	}

	select {
	case c.messageRouter <- msg:
		return nil
	default:
		return fmt.Errorf("message router is full")
	}
}

// routeMessages handles message routing between agents
func (c *Coordinator) routeMessages() {
	log.Println("Message router started")

	for msg := range c.messageRouter {
		c.routeMessage(msg)
	}

	log.Println("Message router stopped")
}

// routeMessage routes a single message to the appropriate agent
func (c *Coordinator) routeMessage(msg AgentMessage) {
	// Route to target agent
	if msg.ToAgent == "user" {
		c.handleUserResponse(msg)
		return
	}

	if agent, exists := c.agents[msg.ToAgent]; exists {
		select {
		case agent.Inbox <- msg:
			log.Printf("Routed message from %s to %s: %s", msg.FromAgent, msg.ToAgent, msg.TaskType)
		default:
			log.Printf("Warning: Agent %s inbox is full", msg.ToAgent)
		}
	} else {
		log.Printf("Warning: Unknown target agent: %s", msg.ToAgent)
	}
}

// handleUserResponse handles responses meant for the user
func (c *Coordinator) handleUserResponse(msg AgentMessage) {
	log.Printf("\n=== RESPONSE FROM %s ===", msg.FromAgent)
	log.Printf("Task: %s", msg.TaskType)
	log.Printf("Status: %s", msg.Status)
	log.Printf("Content: %s", msg.Content)
	log.Printf("========================\n")

	// Update project context based on response
	if msg.Status == TaskCompleted {
		c.context.UpdateProjectContext(map[string]interface{}{
			"completed_tasks": []string{msg.TaskType},
		})
	}
}

// GetProjectStatus returns the current project status
func (c *Coordinator) GetProjectStatus() map[string]interface{} {
	return c.context.GetProjectStatus()
}

// WaitForCompletion waits for all current tasks to complete
func (c *Coordinator) WaitForCompletion(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	lastActivityTime := time.Now()
	checkCount := 0

	for time.Now().Before(deadline) {
		// Check if all agents are idle and no messages pending
		allIdle := true
		hasActivity := false

		for _, agent := range c.agents {
			isProcessing := agent.Processing

			if len(agent.Inbox) > 0 || len(agent.Outbox) > 0 || isProcessing {
				allIdle = false
				hasActivity = true
				lastActivityTime = time.Now()
				break
			}
		}

		// Also check if message router has pending messages
		if len(c.messageRouter) > 0 {
			allIdle = false
			hasActivity = true
			lastActivityTime = time.Now()
		}

		if hasActivity {
			log.Printf("Agents still processing... (messages: %d, active agents: %d)", c.getTotalPendingMessages(), c.getActiveProcessingCount())
			checkCount = 0
		} else {
			checkCount++
		}

		// If no activity for 3 seconds and multiple checks, consider tasks complete
		if allIdle && time.Since(lastActivityTime) > 3*time.Second && checkCount >= 3 {
			log.Println("All tasks completed")
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("timeout waiting for task completion")
}

// getActiveProcessingCount returns number of agents currently processing
func (c *Coordinator) getActiveProcessingCount() int {
	count := 0
	for _, agent := range c.agents {
		if agent.Processing {
			count++
		}
	}
	return count
}

// getTotalPendingMessages returns total pending messages across all agents
func (c *Coordinator) getTotalPendingMessages() int {
	total := 0
	for _, agent := range c.agents {
		total += len(agent.Inbox) + len(agent.Outbox)
	}
	total += len(c.messageRouter)
	return total
}

// SendAgentMessage allows external sending of messages between agents
func (c *Coordinator) SendAgentMessage(from, to AgentType, taskType, content string) error {
	msg := AgentMessage{
		ID:        generateID(),
		FromAgent: from,
		ToAgent:   to,
		TaskType:  taskType,
		Content:   content,
		Status:    TaskPending,
		Timestamp: getCurrentTimestamp(),
	}

	return c.sendMessage(msg)
}

// ListActiveAgents returns a list of currently active agents
func (c *Coordinator) ListActiveAgents() []AgentType {
	var active []AgentType
	for agentType := range c.agents {
		active = append(active, agentType)
	}
	return active
}

// GetAvailableModels returns available models from the LLM service
func (c *Coordinator) GetAvailableModels() map[string][]string {
	if c.llmService != nil {
		return c.llmService.GetAvailableModels()
	}
	return map[string][]string{}
}

// GetAllAvailableModels returns all available models from all providers
func (c *Coordinator) GetAllAvailableModels() map[string]map[string][]string {
	if c.llmService != nil {
		return c.llmService.GetAllAvailableModels()
	}
	return map[string]map[string][]string{}
}

// GetLLMProvider returns the current LLM provider
func (c *Coordinator) GetLLMProvider() string {
	if c.llmService != nil {
		return string(c.llmService.Provider)
	}
	return "none"
}

// SwitchModel switches the LLM provider and model
func (c *Coordinator) SwitchModel(provider, model string, autoMode bool) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Convert string provider to LLMProvider type
	var llmProvider llm.LLMProvider
	switch provider {
	case "ollama":
		llmProvider = llm.OllamaProvider
	case "bedrock":
		llmProvider = llm.BedrockProvider
	case "openrouter":
		llmProvider = llm.OpenRouterProvider
	case "gemini":
		llmProvider = llm.GeminiProvider
	case "anthropic":
		llmProvider = llm.AnthropicProvider
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	// Create new LLM service with the specified provider and model
	newLLMService, err := llm.NewLLMService(llmProvider, model)
	if err != nil {
		return fmt.Errorf("failed to create new LLM service: %v", err)
	}

	// Update the coordinator's LLM service
	c.llmService = newLLMService

	// Update all agents with the new LLM service
	for _, agent := range c.agents {
		agent.LLMService = newLLMService
	}

	log.Printf("[MODEL_SWITCH] Successfully switched to provider=%s model=%s auto_mode=%v", provider, model, autoMode)
	return nil
}
