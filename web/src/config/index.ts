// Environment configuration for the React app
export const config = {
  // API Configuration
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8084/api',
  websocketUrl: import.meta.env.VITE_WEBSOCKET_URL || 'ws://localhost:8084/api/v1/chat/stream',
  
  // LLM Configuration
  defaultLLMProvider: import.meta.env.VITE_DEFAULT_LLM_PROVIDER || 'ollama',
  defaultModel: import.meta.env.VITE_DEFAULT_MODEL || 'qwen2.5:1.5b',
  
  // Available Models
  ollamaModels: (import.meta.env.VITE_OLLAMA_MODELS || 'qwen2.5:1.5b,cogito:8b,gemma3:1b,qwen3:0.6b').split(','),
  bedrockModels: (import.meta.env.VITE_BEDROCK_MODELS || 'anthropic.claude-3-5-sonnet-20241022-v2:0,anthropic.claude-3-haiku-20240307-v1:0,anthropic.claude-3-opus-20240229-v1:0').split(','),
  
  // Development Configuration
  isDevelopment: import.meta.env.VITE_DEV_MODE === 'true' || import.meta.env.DEV,
  logLevel: import.meta.env.VITE_LOG_LEVEL || 'info',
} as const;

// Type definitions for LLM providers
export type LLMProvider = 'ollama' | 'bedrock';

// Available models by provider
export const modelsByProvider = {
  ollama: config.ollamaModels,
  bedrock: config.bedrockModels,
} as const;

// Utility function to get models for a specific provider
export const getModelsForProvider = (provider: LLMProvider): string[] => {
  return modelsByProvider[provider] || [];
};

// Utility function to validate configuration
export const validateConfig = (): boolean => {
  const required = ['apiBaseUrl', 'websocketUrl'];
  
  for (const key of required) {
    if (!config[key as keyof typeof config]) {
      console.error(`Missing required configuration: ${key}`);
      return false;
    }
  }
  
  return true;
};

// Log configuration in development
if (config.isDevelopment) {
  console.log('App Configuration:', {
    apiBaseUrl: config.apiBaseUrl,
    websocketUrl: config.websocketUrl,
    defaultProvider: config.defaultLLMProvider,
    defaultModel: config.defaultModel,
    ollamaModels: config.ollamaModels.length,
    bedrockModels: config.bedrockModels.length,
  });
}
