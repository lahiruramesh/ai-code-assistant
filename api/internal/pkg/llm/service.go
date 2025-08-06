package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/ollama/ollama/api"
)

// LLMProvider represents different LLM providers
type LLMProvider string

const (
	OllamaProvider  LLMProvider = "ollama"
	BedrockProvider LLMProvider = "bedrock"
)

// LLMRequest represents a request to generate text
type LLMRequest struct {
	Model     string                 `json:"model"`
	Prompt    string                 `json:"prompt"`
	MaxTokens int                    `json:"max_tokens,omitempty"`
	Tools     []api.Tool             `json:"tools,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// LLMResponse represents the response from LLM
type LLMResponse struct {
	Text      string                 `json:"text"`
	Model     string                 `json:"model"`
	Provider  string                 `json:"provider"`
	ToolCalls []api.ToolCall         `json:"tool_calls,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// LLMService provides a unified interface for different LLM providers
type LLMService struct {
	Provider      LLMProvider
	OllamaClient  *api.Client
	BedrockClient *bedrockruntime.Client
	DefaultModel  string
}

// NewLLMService creates a new LLM service with the specified provider
func NewLLMService(provider LLMProvider, defaultModel string) (*LLMService, error) {
	service := &LLMService{
		Provider:     provider,
		DefaultModel: defaultModel,
	}

	switch provider {
	case OllamaProvider:
		client, err := api.ClientFromEnvironment()
		if err != nil {
			return nil, fmt.Errorf("failed to create Ollama client: %v", err)
		}
		service.OllamaClient = client

	case BedrockProvider:
		// Get AWS configuration from environment variables
		region := os.Getenv("AWS_REGION")
		if region == "" {
			region = "eu-central-1" // Default region
		}

		accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
		secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

		var cfg aws.Config
		var err error

		if accessKey != "" && secretKey != "" {
			// Use explicit credentials if provided
			cfg, err = config.LoadDefaultConfig(context.TODO(),
				config.WithRegion(region),
				config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
			)
		} else {
			// Use default credential chain
			cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
		}

		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %v", err)
		}
		service.BedrockClient = bedrockruntime.NewFromConfig(cfg)
		log.Printf("Initialized Bedrock client with region: %s", region)

	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}

	return service, nil
}

// Generate generates text using the configured LLM provider
func (s *LLMService) Generate(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
	if req.Model == "" {
		req.Model = s.DefaultModel
	}

	switch s.Provider {
	case OllamaProvider:
		return s.generateWithOllama(ctx, req)
	case BedrockProvider:
		return s.generateWithBedrock(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", s.Provider)
	}
}

// generateWithOllama generates text using Ollama
func (s *LLMService) generateWithOllama(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
	log.Printf("[OLLAMA] Generating with model: %s", req.Model)

	ollamaReq := &api.GenerateRequest{
		Model:  req.Model,
		Prompt: req.Prompt,
		Stream: func(b bool) *bool { return &b }(false),
	}

	var response string
	err := s.OllamaClient.Generate(ctx, ollamaReq, func(resp api.GenerateResponse) error {
		response += resp.Response
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("ollama generation failed: %v", err)
	}

	return &LLMResponse{
		Text:     response,
		Model:    req.Model,
		Provider: string(OllamaProvider),
		Metadata: map[string]interface{}{
			"length": len(response),
		},
	}, nil
}

// generateWithBedrock generates text using AWS Bedrock
func (s *LLMService) generateWithBedrock(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
	log.Printf("[BEDROCK] Generating with model: %s", req.Model)

	// Handle different model families
	var body []byte
	var err error

	switch {
	case strings.Contains(req.Model, "claude"):
		body, err = s.buildClaudeRequest(req)
	case strings.Contains(req.Model, "llama"):
		body, err = s.buildLlamaRequest(req)
	case strings.Contains(req.Model, "titan"):
		body, err = s.buildTitanRequest(req)
	default:
		// Default to Claude format
		body, err = s.buildClaudeRequest(req)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to build request body: %v", err)
	}

	input := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(req.Model),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        body,
	}

	result, err := s.BedrockClient.InvokeModel(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("bedrock invocation failed: %v", err)
	}

	// Parse response based on model type
	response, err := s.parseBedrockResponse(result.Body, req.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bedrock response: %v", err)
	}

	return response, nil
}

// buildClaudeRequest builds request body for Claude models
func (s *LLMService) buildClaudeRequest(req LLMRequest) ([]byte, error) {
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4000
	}

	body := map[string]interface{}{
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": req.Prompt,
			},
		},
		"max_tokens":        maxTokens,
		"temperature":       0.7,
		"anthropic_version": "bedrock-2023-05-31",
	}

	// Add tools if provided
	if len(req.Tools) > 0 {
		tools := make([]map[string]interface{}, 0, len(req.Tools))
		for _, tool := range req.Tools {
			toolData := map[string]interface{}{
				"name":        tool.Function.Name,
				"description": tool.Function.Description,
				"input_schema": map[string]interface{}{
					"type":       "object",
					"properties": tool.Function.Parameters.Properties,
					"required":   tool.Function.Parameters.Required,
				},
			}
			tools = append(tools, toolData)
		}
		body["tools"] = tools

		log.Printf("[BEDROCK] Added %d tools to request", len(req.Tools))
	}

	return json.Marshal(body)
}

// buildLlamaRequest builds request body for Llama models
func (s *LLMService) buildLlamaRequest(req LLMRequest) ([]byte, error) {
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4000
	}

	body := map[string]interface{}{
		"prompt":      req.Prompt,
		"max_gen_len": maxTokens,
		"temperature": 0.7,
		"top_p":       0.9,
	}

	return json.Marshal(body)
}

// buildTitanRequest builds request body for Titan models
func (s *LLMService) buildTitanRequest(req LLMRequest) ([]byte, error) {
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4000
	}

	body := map[string]interface{}{
		"inputText": req.Prompt,
		"textGenerationConfig": map[string]interface{}{
			"maxTokenCount": maxTokens,
			"temperature":   0.7,
			"topP":          0.9,
		},
	}

	return json.Marshal(body)
}

// parseBedrockResponse parses the response from Bedrock based on model type
func (s *LLMService) parseBedrockResponse(body []byte, model string) (*LLMResponse, error) {
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	llmResponse := &LLMResponse{
		Model:    model,
		Provider: string(BedrockProvider),
		Metadata: map[string]interface{}{
			"length":    0,
			"timestamp": time.Now().Unix(),
		},
	}

	switch {
	case strings.Contains(model, "claude"):
		if content, ok := response["content"].([]interface{}); ok && len(content) > 0 {
			var textParts []string
			var toolCalls []api.ToolCall

			for _, item := range content {
				if contentBlock, ok := item.(map[string]interface{}); ok {
					if contentType, exists := contentBlock["type"]; exists {
						switch contentType {
						case "text":
							if text, ok := contentBlock["text"].(string); ok {
								textParts = append(textParts, text)
							}
						case "tool_use":
							if name, ok := contentBlock["name"].(string); ok {
								if _, ok := contentBlock["id"].(string); ok {
									var args map[string]interface{}
									if input, ok := contentBlock["input"].(map[string]interface{}); ok {
										args = input
									}

									toolCall := api.ToolCall{
										Function: api.ToolCallFunction{
											Name:      name,
											Arguments: args,
										},
									}
									toolCalls = append(toolCalls, toolCall)

									log.Printf("[BEDROCK] Found tool call: %s with args: %v", name, args)
								}
							}
						}
					}
				}
			}

			llmResponse.Text = strings.Join(textParts, "\n")
			llmResponse.ToolCalls = toolCalls
			llmResponse.Metadata["length"] = len(llmResponse.Text)
			return llmResponse, nil
		}
	case strings.Contains(model, "llama"):
		if generation, ok := response["generation"].(string); ok {
			llmResponse.Text = generation
			llmResponse.Metadata["length"] = len(generation)
			return llmResponse, nil
		}
	case strings.Contains(model, "titan"):
		if results, ok := response["results"].([]interface{}); ok && len(results) > 0 {
			if result, ok := results[0].(map[string]interface{}); ok {
				if text, ok := result["outputText"].(string); ok {
					llmResponse.Text = text
					llmResponse.Metadata["length"] = len(text)
					return llmResponse, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("unable to parse response for model: %s", model)
}

// GetAvailableModels returns available models for the provider
func (s *LLMService) GetAvailableModels() map[string][]string {
	switch s.Provider {
	case OllamaProvider:
		return map[string][]string{
			"ollama": {"qwen2.5:1.5b", "cogito:14b", "cogito:8b", "llama3.2:3b"},
		}
	case BedrockProvider:
		return map[string][]string{
			"claude": {
				"anthropic.claude-3-5-sonnet-20241022-v2:0",
				"anthropic.claude-3-sonnet-20240229-v1:0",
				"anthropic.claude-3-haiku-20240307-v1:0",
			},
			"llama": {
				"meta.llama3-2-11b-instruct-v1:0",
				"meta.llama3-2-3b-instruct-v1:0",
				"meta.llama3-2-1b-instruct-v1:0",
			},
			"titan": {
				"amazon.titan-text-express-v1",
				"amazon.titan-text-lite-v1",
			},
		}
	default:
		return map[string][]string{}
	}
}
