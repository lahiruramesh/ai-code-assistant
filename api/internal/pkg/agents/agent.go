package agents

import (
	"agent/internal/pkg/llm"
	"agent/internal/pkg/tools"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ollama/ollama/api"
)

// AgentType represents different types of agents
type AgentType string

const (
	SupervisorAgent  AgentType = "supervisor"
	CodeEditingAgent AgentType = "code_editing"
	ReactAgent       AgentType = "react"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskInProgress TaskStatus = "in_progress"
	TaskCompleted  TaskStatus = "completed"
	TaskFailed     TaskStatus = "failed"
)

// AgentMessage represents communication between agents
type AgentMessage struct {
	ID        string                 `json:"id"`
	FromAgent AgentType              `json:"from_agent"`
	ToAgent   AgentType              `json:"to_agent"`
	TaskType  string                 `json:"task_type"`
	Content   string                 `json:"content"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Status    TaskStatus             `json:"status"`
	Timestamp int64                  `json:"timestamp"`
	ReplyTo   string                 `json:"reply_to,omitempty"`
}

// Task represents a unit of work
type Task struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	AssignedTo   AgentType              `json:"assigned_to"`
	Status       TaskStatus             `json:"status"`
	Input        map[string]interface{} `json:"input"`
	Output       map[string]interface{} `json:"output"`
	Dependencies []string               `json:"dependencies"`
}

// ProjectContext holds the current project state
type ProjectContext struct {
	ProjectName    string            `json:"project_name"`
	ProjectPath    string            `json:"project_path"`
	Requirements   string            `json:"requirements"`
	CurrentPhase   string            `json:"current_phase"`
	CompletedTasks []string          `json:"completed_tasks"`
	ActiveTasks    []string          `json:"active_tasks"`
	ProjectFiles   map[string]string `json:"project_files"`
}

// Agent represents a specialized agent
// Agent represents an agent in the system
type Agent struct {
	Type           AgentType
	Client         *api.Client
	LLMService     *llm.LLMService
	SystemPrompt   string
	Context        map[string]interface{}
	Inbox          chan AgentMessage
	Outbox         chan AgentMessage
	Processing     bool
	UseToolCalling bool // New field to determine if agent should use tool calling
}

// NewAgent creates a new agent with the specified type
func NewAgent(agentType AgentType, client *api.Client, llmService *llm.LLMService, context *ProjectContext) (*Agent, error) {
	systemPrompt, err := loadSystemPrompt(agentType)
	if err != nil {
		return nil, fmt.Errorf("failed to load system prompt for %s: %v", agentType, err)
	}

	// Determine if this agent should use tool calling
	// All agents should have access to tools for their specialized tasks
	useToolCalling := true // Enable tool calling for all agents

	// Convert context to map for compatibility
	contextMap := make(map[string]interface{})
	if context != nil {
		contextMap["project_name"] = context.ProjectName
		contextMap["project_path"] = context.ProjectPath
		contextMap["current_phase"] = context.CurrentPhase
		contextMap["completed_tasks"] = context.CompletedTasks
		contextMap["active_tasks"] = context.ActiveTasks
		contextMap["project_files"] = context.ProjectFiles
	}

	return &Agent{
		Type:           agentType,
		SystemPrompt:   systemPrompt,
		Client:         client,     // Keep for backward compatibility
		LLMService:     llmService, // New LLM service
		Inbox:          make(chan AgentMessage, 100),
		Outbox:         make(chan AgentMessage, 100),
		Context:        contextMap,
		Processing:     false,
		UseToolCalling: useToolCalling,
	}, nil
}

// Start begins the agent's message processing loop
func (a *Agent) Start(wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("[%s] Agent started", a.Type)

	for msg := range a.Inbox {
		a.processMessage(msg)
	}

	log.Printf("[%s] Agent stopped", a.Type)
}

// Stop gracefully stops the agent
func (a *Agent) Stop() {
	close(a.Inbox)
}

// SendMessage sends a message to another agent
func (a *Agent) SendMessage(msg AgentMessage) {
	msg.FromAgent = a.Type
	a.Outbox <- msg
}

// processMessage handles incoming messages
func (a *Agent) processMessage(msg AgentMessage) {
	a.Processing = true

	defer func() {
		a.Processing = false
	}()

	log.Printf("[%s] Processing message from %s: %s", a.Type, msg.FromAgent, msg.TaskType)

	// Generate response using LLM
	response, err := a.generateResponse(msg)
	if err != nil {
		log.Printf("[%s] Error generating response: %v", a.Type, err)
		// Send error response
		errorMsg := AgentMessage{
			ID:        generateID(),
			FromAgent: a.Type,
			ToAgent:   msg.FromAgent,
			TaskType:  "error",
			Content:   fmt.Sprintf("Error processing task: %v", err),
			Status:    TaskFailed,
			ReplyTo:   msg.ID,
			Timestamp: getCurrentTimestamp(),
		}
		a.SendMessage(errorMsg)
		return
	}

	// For supervisor agent, parse delegation instructions
	if a.Type == SupervisorAgent {
		log.Printf("[%s] Parsing delegation from response length: %d", a.Type, len(response))
		a.parseDelegation(response, msg)
	}

	// Send response back
	responseMsg := AgentMessage{
		ID:        generateID(),
		FromAgent: a.Type,
		ToAgent:   msg.FromAgent,
		TaskType:  msg.TaskType + "_response",
		Content:   response,
		Status:    TaskCompleted,
		ReplyTo:   msg.ID,
		Timestamp: getCurrentTimestamp(),
	}
	log.Printf("[%s] Sending response to %s: %s", a.Type, responseMsg.ToAgent, responseMsg.TaskType)
	a.SendMessage(responseMsg)
}

// generateResponse uses the LLM service to generate a response
func (a *Agent) generateResponse(msg AgentMessage) (string, error) {
	log.Printf("[%s] Starting response generation for message: %s", a.Type, msg.Content)

	// Prepare the prompt with system prompt and context
	prompt := a.buildPrompt(msg)
	log.Printf("[%s] Built prompt with length: %d", a.Type, len(prompt))

	// Use LLM service if available, otherwise fallback to Ollama client
	if a.LLMService != nil {
		req := llm.LLMRequest{
			Prompt:    prompt,
			MaxTokens: 4000,
			Tools:     tools.GetAllTools(), // Add tools support
			Metadata: map[string]interface{}{
				"agent_type": string(a.Type),
				"task_type":  msg.TaskType,
			},
		}

		log.Printf("[%s] Sending request to LLM service (%s) with %d tools", a.Type, a.LLMService.Provider, len(req.Tools))

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		response, err := a.LLMService.Generate(ctx, req)
		if err != nil {
			log.Printf("[%s] LLM service generation failed: %v", a.Type, err)
			return "", fmt.Errorf("LLM generation failed: %v", err)
		}

		log.Printf("[%s] Generated response with length: %d", a.Type, len(response.Text))
		log.Printf("[%s] Response preview: %.200s...", a.Type, response.Text)

		// Handle tool calls if present
		if len(response.ToolCalls) > 0 {
			log.Printf("[%s] Processing %d tool calls", a.Type, len(response.ToolCalls))
			toolResults := make([]string, 0, len(response.ToolCalls))

			for i, toolCall := range response.ToolCalls {
				log.Printf("[%s] ===== TOOL CALL %d START =====", a.Type, i+1)
				log.Printf("[%s] Tool Name: %s", a.Type, toolCall.Function.Name)
				
				// Log input parameters - convert arguments to JSON for logging
				if len(toolCall.Function.Arguments) > 0 {
					argsJSON, err := json.Marshal(toolCall.Function.Arguments)
					if err != nil {
						log.Printf("[%s] Tool Input Parameters: (failed to marshal: %v)", a.Type, err)
					} else {
						log.Printf("[%s] Tool Input Parameters: %s", a.Type, string(argsJSON))
					}
				} else {
					log.Printf("[%s] Tool Input Parameters: (none)", a.Type)
				}
				
				log.Printf("[%s] Executing tool call %d: %s", a.Type, i+1, toolCall.Function.Name)
				result, err := tools.ExecuteToolCall(toolCall)
				
				if err != nil {
					log.Printf("[%s] Tool call %d FAILED: %v", a.Type, i+1, err)
					log.Printf("[%s] Tool Error Details: %v", a.Type, err)
					toolResults = append(toolResults, fmt.Sprintf("Tool %s failed: %v", toolCall.Function.Name, err))
				} else {
					log.Printf("[%s] Tool call %d SUCCEEDED", a.Type, i+1)
					log.Printf("[%s] Tool Response: %s", a.Type, result)
					toolResults = append(toolResults, fmt.Sprintf("Tool %s result: %s", toolCall.Function.Name, result))
				}
				log.Printf("[%s] ===== TOOL CALL %d END =====", a.Type, i+1)
			}

			// Combine the text response with tool results
			fullResponse := response.Text
			if len(toolResults) > 0 {
				fullResponse += "\n\nTool Execution Results:\n" + strings.Join(toolResults, "\n")
			}

			return fullResponse, nil
		}

		return response.Text, nil
	}

	// Fallback to original Ollama client
	req := &api.GenerateRequest{
		Model:  "cogito:8b",
		Prompt: prompt,
		Stream: func(b bool) *bool { return &b }(false),
	}

	log.Printf("[%s] Sending request to Ollama with model: %s", a.Type, req.Model)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var response string
	err := a.Client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		response += resp.Response
		return nil
	})

	if err != nil {
		log.Printf("[%s] Ollama generation failed: %v", a.Type, err)
		return "", fmt.Errorf("ollama generation failed: %v", err)
	}

	log.Printf("[%s] Generated response with length: %d", a.Type, len(response))
	log.Printf("[%s] Response preview: %.200s...", a.Type, response)
	return response, nil
}

// buildPrompt constructs the prompt for Ollama
func (a *Agent) buildPrompt(msg AgentMessage) string {
	contextJSON, _ := json.MarshalIndent(a.Context, "", "  ")

	return fmt.Sprintf(`%s

CURRENT PROJECT CONTEXT:
%s

INCOMING TASK:
Type: %s
From: %s
Content: %s

Additional Data: %v

Please provide a detailed response that:
1. Acknowledges the task
2. Provides specific actions taken or to be taken
3. Returns structured data if needed
4. Indicates next steps or completion status
5. Maintains consistency with the project context

Response:`,
		a.SystemPrompt,
		string(contextJSON),
		msg.TaskType,
		msg.FromAgent,
		msg.Content,
		msg.Data,
	)
}

// loadSystemPrompt loads the system prompt from file
func loadSystemPrompt(agentType AgentType) (string, error) {
	filename := fmt.Sprintf("prompts/%s.txt", string(agentType))
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// parseDelegation parses delegation instructions from supervisor response
func (a *Agent) parseDelegation(response string, originalMsg AgentMessage) {
	log.Printf("[%s] Starting delegation parsing for response", a.Type)
	lines := strings.Split(response, "\n")

	var delegateToAgent string
	var task string
	var instructions string

	for i, line := range lines {
		line = strings.TrimSpace(line)
		log.Printf("[%s] Processing line %d: '%s'", a.Type, i, line)

		// Handle both plain format and markdown formatting (with asterisks)
		if strings.HasPrefix(line, "**DELEGATE_TO:**") || strings.HasPrefix(line, "DELEGATE_TO:") {
			if strings.HasPrefix(line, "**DELEGATE_TO:**") {
				delegateToAgent = strings.TrimSpace(strings.TrimPrefix(line, "**DELEGATE_TO:**"))
			} else {
				delegateToAgent = strings.TrimSpace(strings.TrimPrefix(line, "DELEGATE_TO:"))
			}
			// Remove backticks, quotes, and other formatting characters
			delegateToAgent = strings.Trim(delegateToAgent, "`\"'* ")
			log.Printf("[%s] Found delegation target: '%s'", a.Type, delegateToAgent)
		} else if strings.HasPrefix(line, "**TASK:**") || strings.HasPrefix(line, "TASK:") {
			if strings.HasPrefix(line, "**TASK:**") {
				task = strings.TrimSpace(strings.TrimPrefix(line, "**TASK:**"))
			} else {
				task = strings.TrimSpace(strings.TrimPrefix(line, "TASK:"))
			}
			// Remove backticks, quotes, and other formatting characters
			task = strings.Trim(task, "`\"'* ")
			log.Printf("[%s] Found task: '%s'", a.Type, task)
		} else if strings.HasPrefix(line, "**INSTRUCTIONS:**") || strings.HasPrefix(line, "INSTRUCTIONS:") {
			if strings.HasPrefix(line, "**INSTRUCTIONS:**") {
				instructions = strings.TrimSpace(strings.TrimPrefix(line, "**INSTRUCTIONS:**"))
			} else {
				instructions = strings.TrimSpace(strings.TrimPrefix(line, "INSTRUCTIONS:"))
			}
			log.Printf("[%s] Found instructions: '%s'", a.Type, instructions[:minInt(100, len(instructions))]+"...")
		}
		// Handle JSON format (e.g., "DELEGATE_TO": "react",)
		if strings.Contains(line, `"DELEGATE_TO":`) {
			// Extract value from JSON format: "DELEGATE_TO": "value",
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				value := strings.TrimSpace(parts[1])
				// Remove quotes, commas, and other JSON formatting
				delegateToAgent = strings.Trim(value, `"',`)
				log.Printf("[%s] Found JSON delegation target: '%s'", a.Type, delegateToAgent)
			}
		} else if strings.Contains(line, `"TASK":`) {
			// Extract value from JSON format: "TASK": "value",
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				value := strings.TrimSpace(parts[1])
				// Remove quotes, commas, and other JSON formatting
				task = strings.Trim(value, `"',`)
				log.Printf("[%s] Found JSON task: '%s'", a.Type, task)
			}
		} else if strings.Contains(line, `"INSTRUCTIONS":`) {
			// Extract value from JSON format: "INSTRUCTIONS": "value",
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				value := strings.TrimSpace(strings.Join(parts[1:], ":")) // In case instructions contain colons
				// Remove quotes and leading comma, but keep the content
				value = strings.TrimLeft(value, ` "`)
				if strings.HasSuffix(value, `",`) {
					value = value[:len(value)-2] // Remove trailing ",
				} else if strings.HasSuffix(value, `"`) {
					value = value[:len(value)-1] // Remove trailing "
				}
				instructions = value
				log.Printf("[%s] Found JSON instructions: '%s'", a.Type, instructions[:minInt(100, len(instructions))]+"...")
			}
		}
	}

	log.Printf("[%s] Delegation parsing results - Agent: '%s', Task: '%s', Instructions length: %d",
		a.Type, delegateToAgent, task, len(instructions))

	// If we have delegation instructions, create and send message
	if delegateToAgent != "" && task != "" {
		var targetAgent AgentType
		switch delegateToAgent {
		case "code_editing":
			targetAgent = CodeEditingAgent
		case "react":
			targetAgent = ReactAgent
		default:
			log.Printf("[%s] Unknown agent type for delegation: '%s' (cleaned from response)", a.Type, delegateToAgent)
			return
		}

		delegationMsg := AgentMessage{
			ID:        generateID(),
			FromAgent: a.Type,
			ToAgent:   targetAgent,
			TaskType:  task,
			Content:   instructions,
			Status:    TaskPending,
			ReplyTo:   originalMsg.ID,
			Timestamp: getCurrentTimestamp(),
		}

		log.Printf("[%s] Delegating task '%s' to %s", a.Type, task, targetAgent)
		a.SendMessage(delegationMsg)
	}
}

// Helper functions
func generateID() string {
	return fmt.Sprintf("msg_%d", getCurrentTimestamp())
}

func getCurrentTimestamp() int64 {
	return 1704067200 // Simplified timestamp
}

// minInt returns the smaller of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// UpdateProjectContext safely updates the project context
func (pc *ProjectContext) UpdateProjectContext(updates map[string]interface{}) {
	for key, value := range updates {
		switch key {
		case "current_phase":
			if phase, ok := value.(string); ok {
				pc.CurrentPhase = phase
			}
		case "completed_tasks":
			if tasks, ok := value.([]string); ok {
				pc.CompletedTasks = append(pc.CompletedTasks, tasks...)
			}
		case "project_files":
			if files, ok := value.(map[string]string); ok {
				for fileName, content := range files {
					pc.ProjectFiles[fileName] = content
				}
			}
		}
	}
}

// GetProjectStatus returns current project status
func (pc *ProjectContext) GetProjectStatus() map[string]interface{} {
	return map[string]interface{}{
		"project_name":    pc.ProjectName,
		"project_path":    pc.ProjectPath,
		"current_phase":   pc.CurrentPhase,
		"completed_tasks": pc.CompletedTasks,
		"active_tasks":    pc.ActiveTasks,
		"file_count":      len(pc.ProjectFiles),
	}
}
