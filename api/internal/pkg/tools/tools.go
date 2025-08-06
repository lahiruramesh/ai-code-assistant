package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

var ReadFileTool = api.Tool{
	Type: "function",
	Function: api.ToolFunction{
		Name:        "read_file",
		Description: "Read the contents of a given relative file path. Use this when you want to see what's inside a file. Do not use this with directory names.",
		Parameters: struct {
			Type       string   `json:"type"`
			Defs       any      `json:"$defs,omitempty"`
			Items      any      `json:"items,omitempty"`
			Required   []string `json:"required"`
			Properties map[string]struct {
				Type        api.PropertyType `json:"type"`
				Items       any              `json:"items,omitempty"`
				Description string           `json:"description"`
				Enum        []any            `json:"enum,omitempty"`
			} `json:"properties"`
		}{
			Type: "object",
			Properties: map[string]struct {
				Type        api.PropertyType `json:"type"`
				Items       any              `json:"items,omitempty"`
				Description string           `json:"description"`
				Enum        []any            `json:"enum,omitempty"`
			}{
				"file_path": {
					Type:        api.PropertyType{"string"},
					Description: "The relative path to the file to read",
				},
			},
			Required: []string{"file_path"},
		},
	},
}

var WriteFileTool = api.Tool{
	Type: "function",
	Function: api.ToolFunction{
		Name:        "write_file",
		Description: "Write content to a file at the given relative path. Creates the file if it doesn't exist.",
		Parameters: struct {
			Type       string   `json:"type"`
			Defs       any      `json:"$defs,omitempty"`
			Items      any      `json:"items,omitempty"`
			Required   []string `json:"required"`
			Properties map[string]struct {
				Type        api.PropertyType `json:"type"`
				Items       any              `json:"items,omitempty"`
				Description string           `json:"description"`
				Enum        []any            `json:"enum,omitempty"`
			} `json:"properties"`
		}{
			Type: "object",
			Properties: map[string]struct {
				Type        api.PropertyType `json:"type"`
				Items       any              `json:"items,omitempty"`
				Description string           `json:"description"`
				Enum        []any            `json:"enum,omitempty"`
			}{
				"file_path": {
					Type:        api.PropertyType{"string"},
					Description: "The relative path to the file to write",
				},
				"content": {
					Type:        api.PropertyType{"string"},
					Description: "The content to write to the file",
				},
			},
			Required: []string{"file_path", "content"},
		},
	},
}

var ListDirectoryTool = api.Tool{
	Type: "function",
	Function: api.ToolFunction{
		Name:        "list_directory",
		Description: "List the contents of a directory. Shows files and subdirectories.",
		Parameters: struct {
			Type       string   `json:"type"`
			Defs       any      `json:"$defs,omitempty"`
			Items      any      `json:"items,omitempty"`
			Required   []string `json:"required"`
			Properties map[string]struct {
				Type        api.PropertyType `json:"type"`
				Items       any              `json:"items,omitempty"`
				Description string           `json:"description"`
				Enum        []any            `json:"enum,omitempty"`
			} `json:"properties"`
		}{
			Type: "object",
			Properties: map[string]struct {
				Type        api.PropertyType `json:"type"`
				Items       any              `json:"items,omitempty"`
				Description string           `json:"description"`
				Enum        []any            `json:"enum,omitempty"`
			}{
				"dir_path": {
					Type:        api.PropertyType{"string"},
					Description: "The relative path to the directory to list (default: current directory)",
				},
			},
			Required: []string{},
		},
	},
}

// CreateDirectoryTool defines a tool function for creating directories
var CreateDirectoryTool = api.Tool{
	Type: "function",
	Function: api.ToolFunction{
		Name:        "create_directory",
		Description: "Create a new directory at the specified path. Creates parent directories if they don't exist.",
		Parameters: struct {
			Type       string   `json:"type"`
			Defs       any      `json:"$defs,omitempty"`
			Items      any      `json:"items,omitempty"`
			Required   []string `json:"required"`
			Properties map[string]struct {
				Type        api.PropertyType `json:"type"`
				Items       any              `json:"items,omitempty"`
				Description string           `json:"description"`
				Enum        []any            `json:"enum,omitempty"`
			} `json:"properties"`
		}{
			Type: "object",
			Properties: map[string]struct {
				Type        api.PropertyType `json:"type"`
				Items       any              `json:"items,omitempty"`
				Description string           `json:"description"`
				Enum        []any            `json:"enum,omitempty"`
			}{
				"dir_path": {
					Type:        api.PropertyType{"string"},
					Description: "The path to the directory to create",
				},
			},
			Required: []string{"dir_path"},
		},
	},
}

// ExecuteCommandTool defines a tool function for executing shell commands
var ExecuteCommandTool = api.Tool{
	Type: "function",
	Function: api.ToolFunction{
		Name:        "execute_command",
		Description: "Execute a shell command in the specified directory. Use with caution.",
		Parameters: struct {
			Type       string   `json:"type"`
			Defs       any      `json:"$defs,omitempty"`
			Items      any      `json:"items,omitempty"`
			Required   []string `json:"required"`
			Properties map[string]struct {
				Type        api.PropertyType `json:"type"`
				Items       any              `json:"items,omitempty"`
				Description string           `json:"description"`
				Enum        []any            `json:"enum,omitempty"`
			} `json:"properties"`
		}{
			Type: "object",
			Properties: map[string]struct {
				Type        api.PropertyType `json:"type"`
				Items       any              `json:"items,omitempty"`
				Description string           `json:"description"`
				Enum        []any            `json:"enum,omitempty"`
			}{
				"command": {
					Type:        api.PropertyType{"string"},
					Description: "The command to execute",
				},
				"working_dir": {
					Type:        api.PropertyType{"string"},
					Description: "The working directory to execute the command in (optional)",
				},
			},
			Required: []string{"command"},
		},
	},
}

// GetAllTools returns all available tools
func GetAllTools() []api.Tool {
	return []api.Tool{
		ReadFileTool,
		WriteFileTool,
		ListDirectoryTool,
		CreateDirectoryTool,
		ExecuteCommandTool,
	}
}

// ExecuteToolCall executes a tool call and returns the result with comprehensive logging
func ExecuteToolCall(toolCall api.ToolCall) (string, error) {
	executionID := generateExecutionID()
	startTime := time.Now()

	// Log tool call start
	log.Printf("[TOOL_EXEC_START] tool=%s execution_id=%s timestamp=%s args_count=%d",
		toolCall.Function.Name, executionID, startTime.Format(time.RFC3339), len(toolCall.Function.Arguments))

	var result string
	var err error
	var resultSize int

	switch toolCall.Function.Name {
	case "read_file":
		result, err = executeReadFile(map[string]any(toolCall.Function.Arguments))
	case "write_file":
		result, err = executeWriteFile(map[string]any(toolCall.Function.Arguments))
	case "list_directory":
		result, err = executeListDirectory(map[string]any(toolCall.Function.Arguments))
	case "create_directory":
		result, err = executeCreateDirectory(map[string]interface{}(toolCall.Function.Arguments))
	case "execute_command":
		result, err = executeCommand(map[string]interface{}(toolCall.Function.Arguments))
	default:
		err = fmt.Errorf("unknown tool: %s", toolCall.Function.Name)
		log.Printf("[TOOL_EXEC_ERROR] tool=%s execution_id=%s error=unknown_tool",
			toolCall.Function.Name, executionID)
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	success := err == nil

	if result != "" {
		resultSize = len(result)
	}

	// Log tool call completion
	if success {
		log.Printf("[TOOL_EXEC_SUCCESS] tool=%s execution_id=%s duration_ms=%d result_size=%d timestamp=%s",
			toolCall.Function.Name, executionID, duration.Milliseconds(), resultSize, endTime.Format(time.RFC3339))
	} else {
		log.Printf("[TOOL_EXEC_FAILURE] tool=%s execution_id=%s duration_ms=%d error_type=%s timestamp=%s",
			toolCall.Function.Name, executionID, duration.Milliseconds(), getErrorType(err), endTime.Format(time.RFC3339))
	}

	return result, err
}

func executeReadFile(arguments map[string]any) (string, error) {
	filePath, ok := arguments["file_path"].(string)
	if !ok {
		return "", fmt.Errorf("file_path parameter is required and must be a string")
	}

	// Log file access attempt (without content for security)
	log.Printf("[FILE_READ_START] file_path=%s", sanitizeFilePath(filePath))

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[FILE_READ_ERROR] file_path=%s error=open_failed", sanitizeFilePath(filePath))
		return "", fmt.Errorf("failed to open file %s: %v", filePath, err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		log.Printf("[FILE_READ_ERROR] file_path=%s error=read_failed", sanitizeFilePath(filePath))
		return "", fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	log.Printf("[FILE_READ_SUCCESS] file_path=%s content_size=%d", sanitizeFilePath(filePath), len(content))
	return string(content), nil
}

func executeWriteFile(arguments map[string]any) (string, error) {
	filePath, ok := arguments["file_path"].(string)
	if !ok {
		return "", fmt.Errorf("file_path parameter is required and must be a string")
	}

	content, ok := arguments["content"].(string)
	if !ok {
		return "", fmt.Errorf("content parameter is required and must be a string")
	}

	// Log file write attempt (without content for security)
	log.Printf("[FILE_WRITE_START] file_path=%s content_size=%d", sanitizeFilePath(filePath), len(content))

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("[FILE_WRITE_ERROR] file_path=%s error=mkdir_failed", sanitizeFilePath(filePath))
		return "", fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		log.Printf("[FILE_WRITE_ERROR] file_path=%s error=write_failed", sanitizeFilePath(filePath))
		return "", fmt.Errorf("failed to write file %s: %v", filePath, err)
	}

	log.Printf("[FILE_WRITE_SUCCESS] file_path=%s content_size=%d", sanitizeFilePath(filePath), len(content))
	return fmt.Sprintf("Successfully wrote content to %s", filePath), nil
}

func executeListDirectory(arguments map[string]any) (string, error) {
	dirPath := "."
	if path, ok := arguments["dir_path"].(string); ok && path != "" {
		dirPath = path
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %v", dirPath, err)
	}

	var result []string
	for _, entry := range entries {
		if entry.IsDir() {
			result = append(result, entry.Name()+"/")
		} else {
			result = append(result, entry.Name())
		}
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}

// executeCreateDirectory handles the create_directory tool execution
func executeCreateDirectory(arguments map[string]interface{}) (string, error) {
	dirPath, ok := arguments["dir_path"].(string)
	if !ok {
		return "", fmt.Errorf("dir_path parameter is required and must be a string")
	}

	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create directory %s: %v", dirPath, err)
	}

	return fmt.Sprintf("Successfully created directory: %s", dirPath), nil
}

// executeCommand handles the execute_command tool execution
func executeCommand(arguments map[string]interface{}) (string, error) {
	command, ok := arguments["command"].(string)
	if !ok {
		return "", fmt.Errorf("command parameter is required and must be a string")
	}

	workingDir := "."
	if dir, ok := arguments["working_dir"].(string); ok && dir != "" {
		workingDir = dir
	}

	// Execute the command
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Command failed: %v\nOutput: %s", err, string(output)), err
	}

	return string(output), nil
}

// Utility functions for logging and security

// generateExecutionID creates a unique execution ID for tracking
func generateExecutionID() string {
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}

// sanitizeFilePath removes sensitive information from file paths for logging
func sanitizeFilePath(path string) string {
	// Remove any absolute path prefixes and just show relative structure
	cleaned := filepath.Clean(path)
	if strings.HasPrefix(cleaned, "/") {
		parts := strings.Split(cleaned, "/")
		if len(parts) > 3 {
			return filepath.Join("...", parts[len(parts)-2], parts[len(parts)-1])
		}
	}
	return cleaned
}

// getErrorType categorizes errors for structured logging
func getErrorType(err error) string {
	if err == nil {
		return "none"
	}

	errStr := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errStr, "permission"):
		return "permission_denied"
	case strings.Contains(errStr, "not found") || strings.Contains(errStr, "no such"):
		return "not_found"
	case strings.Contains(errStr, "already exists"):
		return "already_exists"
	case strings.Contains(errStr, "timeout"):
		return "timeout"
	case strings.Contains(errStr, "network") || strings.Contains(errStr, "connection"):
		return "network_error"
	case strings.Contains(errStr, "space") || strings.Contains(errStr, "disk"):
		return "disk_space"
	default:
		return "unknown"
	}
}
