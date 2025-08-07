package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
	OllamaProvider     LLMProvider = "ollama"
	BedrockProvider    LLMProvider = "bedrock"
	OpenRouterProvider LLMProvider = "openrouter"
	GeminiProvider     LLMProvider = "gemini"
	AnthropicProvider  LLMProvider = "anthropic"
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
	Text         string                 `json:"text"`
	Model        string                 `json:"model"`
	Provider     string                 `json:"provider"`
	ToolCalls    []api.ToolCall         `json:"tool_calls,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	InputTokens  int                    `json:"input_tokens"`
	OutputTokens int                    `json:"output_tokens"`
	TotalTokens  int                    `json:"total_tokens"`
}

// TokenUsage represents token usage information
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// LLMService provides a unified interface for different LLM providers
type LLMService struct {
	Provider         LLMProvider
	OllamaClient     *api.Client
	BedrockClient    *bedrockruntime.Client
	HTTPClient       *http.Client
	DefaultModel     string
	OpenRouterAPIKey string
	GeminiAPIKey     string
	AnthropicAPIKey  string
}

// NewLLMService creates a new LLM service with the specified provider
func NewLLMService(provider LLMProvider, defaultModel string) (*LLMService, error) {
	service := &LLMService{
		Provider:         provider,
		DefaultModel:     defaultModel,
		HTTPClient:       &http.Client{Timeout: 60 * time.Second},
		OpenRouterAPIKey: os.Getenv("OPENROUTER_API_KEY"),
		GeminiAPIKey:     os.Getenv("GEMINI_API_KEY"),
		AnthropicAPIKey:  os.Getenv("ANTHROPIC_API_KEY"),
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

	case OpenRouterProvider:
		if service.OpenRouterAPIKey == "" {
			return nil, fmt.Errorf("OPENROUTER_API_KEY environment variable is required")
		}
		log.Printf("Initialized OpenRouter client")

	case GeminiProvider:
		if service.GeminiAPIKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
		}
		log.Printf("Initialized Gemini client")

	case AnthropicProvider:
		if service.AnthropicAPIKey == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable is required")
		}
		log.Printf("Initialized Anthropic client")

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
	case OpenRouterProvider:
		return s.generateWithOpenRouter(ctx, req)
	case GeminiProvider:
		return s.generateWithGemini(ctx, req)
	case AnthropicProvider:
		return s.generateWithAnthropic(ctx, req)
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

	// Estimate token usage (rough approximation for Ollama)
	inputTokens := len(strings.Fields(req.Prompt))
	outputTokens := len(strings.Fields(response))

	return &LLMResponse{
		Text:         response,
		Model:        req.Model,
		Provider:     string(OllamaProvider),
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  inputTokens + outputTokens,
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

	// Extract usage information if available
	if usage, ok := response["usage"].(map[string]interface{}); ok {
		if inputTokens, ok := usage["input_tokens"].(float64); ok {
			llmResponse.InputTokens = int(inputTokens)
		}
		if outputTokens, ok := usage["output_tokens"].(float64); ok {
			llmResponse.OutputTokens = int(outputTokens)
		}
		llmResponse.TotalTokens = llmResponse.InputTokens + llmResponse.OutputTokens
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

			// Estimate tokens if not provided
			if llmResponse.TotalTokens == 0 {
				llmResponse.InputTokens = len(strings.Fields(llmResponse.Text)) / 3 // Rough estimate
				llmResponse.OutputTokens = len(strings.Fields(llmResponse.Text))
				llmResponse.TotalTokens = llmResponse.InputTokens + llmResponse.OutputTokens
			}

			return llmResponse, nil
		}
	case strings.Contains(model, "llama"):
		if generation, ok := response["generation"].(string); ok {
			llmResponse.Text = generation
			llmResponse.Metadata["length"] = len(generation)

			// Estimate tokens if not provided
			if llmResponse.TotalTokens == 0 {
				llmResponse.InputTokens = len(strings.Fields(generation)) / 3
				llmResponse.OutputTokens = len(strings.Fields(generation))
				llmResponse.TotalTokens = llmResponse.InputTokens + llmResponse.OutputTokens
			}

			return llmResponse, nil
		}
	case strings.Contains(model, "titan"):
		if results, ok := response["results"].([]interface{}); ok && len(results) > 0 {
			if result, ok := results[0].(map[string]interface{}); ok {
				if text, ok := result["outputText"].(string); ok {
					llmResponse.Text = text
					llmResponse.Metadata["length"] = len(text)

					// Estimate tokens if not provided
					if llmResponse.TotalTokens == 0 {
						llmResponse.InputTokens = len(strings.Fields(text)) / 3
						llmResponse.OutputTokens = len(strings.Fields(text))
						llmResponse.TotalTokens = llmResponse.InputTokens + llmResponse.OutputTokens
					}

					return llmResponse, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("unable to parse response for model: %s", model)
}

// generateWithOpenRouter generates text using OpenRouter API
func (s *LLMService) generateWithOpenRouter(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
	log.Printf("[OPENROUTER] Generating with model: %s", req.Model)

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4000
	}

	body := map[string]interface{}{
		"model": req.Model,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": req.Prompt,
			},
		},
		"max_tokens":  maxTokens,
		"temperature": 0.7,
		"stream":      false,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.OpenRouterAPIKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/lahiruramesh/code-editing-agent")
	httpReq.Header.Set("X-Title", "Code Editing Agent")

	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(respBody))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}

	text := response.Choices[0].Message.Content
	return &LLMResponse{
		Text:         text,
		Model:        req.Model,
		Provider:     string(OpenRouterProvider),
		InputTokens:  response.Usage.PromptTokens,
		OutputTokens: response.Usage.CompletionTokens,
		TotalTokens:  response.Usage.TotalTokens,
		Metadata: map[string]interface{}{
			"length": len(text),
		},
	}, nil
}

// generateWithGemini generates text using Google Gemini API
func (s *LLMService) generateWithGemini(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
	log.Printf("[GEMINI] Generating with model: %s", req.Model)

	// Format model name for Gemini API
	model := req.Model
	if !strings.HasPrefix(model, "gemini-") {
		model = "gemini-1.5-flash" // Default model
	}

	body := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": req.Prompt,
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.7,
			"maxOutputTokens": req.MaxTokens,
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, s.GeminiAPIKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(respBody))
	}

	var response struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response content returned")
	}

	text := response.Candidates[0].Content.Parts[0].Text
	return &LLMResponse{
		Text:         text,
		Model:        req.Model,
		Provider:     string(GeminiProvider),
		InputTokens:  response.UsageMetadata.PromptTokenCount,
		OutputTokens: response.UsageMetadata.CandidatesTokenCount,
		TotalTokens:  response.UsageMetadata.TotalTokenCount,
		Metadata: map[string]interface{}{
			"length": len(text),
		},
	}, nil
}

// generateWithAnthropic generates text using Anthropic Claude API
func (s *LLMService) generateWithAnthropic(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
	log.Printf("[ANTHROPIC] Generating with model: %s", req.Model)

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4000
	}

	body := map[string]interface{}{
		"model": req.Model,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": req.Prompt,
			},
		},
		"max_tokens":  maxTokens,
		"temperature": 0.7,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", s.AnthropicAPIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(respBody))
	}

	var response struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if len(response.Content) == 0 {
		return nil, fmt.Errorf("no response content returned")
	}

	text := response.Content[0].Text
	return &LLMResponse{
		Text:         text,
		Model:        req.Model,
		Provider:     string(AnthropicProvider),
		InputTokens:  response.Usage.InputTokens,
		OutputTokens: response.Usage.OutputTokens,
		TotalTokens:  response.Usage.InputTokens + response.Usage.OutputTokens,
		Metadata: map[string]interface{}{
			"length": len(text),
		},
	}, nil
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
	case OpenRouterProvider:
		return map[string][]string{
			"openai": {
				"openai/gpt-4o",
				"openai/gpt-4o-mini",
				"openai/gpt-3.5-turbo",
			},
			"anthropic": {
				"anthropic/claude-3.5-sonnet",
				"anthropic/claude-3-sonnet",
				"anthropic/claude-3-haiku",
			},
			"google": {
				"google/gemini-2.0-flash-exp",
				"google/gemini-1.5-flash",
				"google/gemini-1.5-pro",
			},
			"meta": {
				"meta-llama/llama-3.1-405b-instruct",
				"meta-llama/llama-3.1-70b-instruct",
				"meta-llama/llama-3.1-8b-instruct",
			},
		}
	case GeminiProvider:
		return map[string][]string{
			"gemini": {
				"gemini-2.0-flash-exp",
				"gemini-1.5-flash",
				"gemini-1.5-pro",
			},
		}
	case AnthropicProvider:
		return map[string][]string{
			"claude": {
				"claude-3-5-sonnet-20241022",
				"claude-3-sonnet-20240229",
				"claude-3-haiku-20240307",
			},
		}
	default:
		return map[string][]string{}
	}
}

// GetAllAvailableModels returns all models from all providers
func (s *LLMService) GetAllAvailableModels() map[string]map[string][]string {
	allModels := make(map[string]map[string][]string)

	// Ollama models
	allModels["ollama"] = map[string][]string{
		"local": {"qwen2.5:1.5b", "cogito:14b", "cogito:8b", "llama3.2:3b"},
	}

	// OpenRouter models
	allModels["openrouter"] = map[string][]string{
		"openai": {
			"openai/gpt-4o",
			"openai/gpt-4o-mini",
			"openai/gpt-3.5-turbo",
		},
		"anthropic": {
			"anthropic/claude-3.5-sonnet",
			"anthropic/claude-3-sonnet",
			"anthropic/claude-3-haiku",
		},
		"google": {
			"google/gemini-1.5-flash",
			"google/gemini-1.5-pro",
			"google/gemini-pro-1.5",
		},
		"meta": {
			"meta-llama/llama-3.1-405b-instruct",
			"meta-llama/llama-3.1-70b-instruct",
			"meta-llama/llama-3.1-8b-instruct",
		},
	}

	// Gemini models
	allModels["gemini"] = map[string][]string{
		"gemini": {
			"gemini-1.5-flash",
			"gemini-1.5-pro",
			"gemini-1.0-pro",
		},
	}

	// Anthropic models
	allModels["anthropic"] = map[string][]string{
		"claude": {
			"claude-3-5-sonnet-20241022",
			"claude-3-sonnet-20240229",
			"claude-3-haiku-20240307",
		},
	}

	// Bedrock models
	allModels["bedrock"] = map[string][]string{
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

	return allModels
}
