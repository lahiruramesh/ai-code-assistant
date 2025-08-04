package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

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

func GetAllTools() api.Tools {
	return api.Tools{
		ReadFileTool,
		WriteFileTool,
		ListDirectoryTool,
	}
}

func ExecuteToolCall(toolCall api.ToolCall) (string, error) {
	switch toolCall.Function.Name {
	case "read_file":
		return executeReadFile(toolCall.Function.Arguments)
	case "write_file":
		return executeWriteFile(toolCall.Function.Arguments)
	case "list_directory":
		return executeListDirectory(toolCall.Function.Arguments)
	default:
		return "", fmt.Errorf("unknown tool: %s", toolCall.Function.Name)
	}
}

func executeReadFile(arguments map[string]any) (string, error) {
	filePath, ok := arguments["file_path"].(string)
	if !ok {
		return "", fmt.Errorf("file_path parameter is required and must be a string")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %v", filePath, err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

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

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file %s: %v", filePath, err)
	}

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
