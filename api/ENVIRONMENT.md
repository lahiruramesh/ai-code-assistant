# Environment Configuration

This project supports comprehensive environment variable configuration for both the Go backend and React frontend.

## Quick Setup

1. **Copy the example files:**
   ```bash
   cp .env.example .env
   cp web/.env.example web/.env
   ```

2. **Update your credentials in `.env`:**
   ```env
   # For AWS Bedrock
   AWS_ACCESS_KEY_ID=your-access-key-here
   AWS_SECRET_ACCESS_KEY=your-secret-key-here
   AWS_REGION=eu-central-1
   
   # Set your preferred LLM provider
   LLM_PROVIDER=bedrock
   LLM_MODEL=anthropic.claude-3-5-sonnet-20240620-v1:0
   ```

## Go Backend Configuration

The Go backend loads environment variables from `.env` file at startup.

### Core Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `LLM_PROVIDER` | `bedrock` | LLM provider: `ollama` or `bedrock` |
| `LLM_MODEL` | `anthropic.claude-3-5-sonnet-20240620-v1:0` | Model to use |
| `SERVER_PORT` | `8084` | HTTP server port |
| `PROJECT_PATH` | `./projects` | Default project directory |

### AWS Bedrock Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `AWS_REGION` | `eu-central-1` | AWS region |
| `AWS_ACCESS_KEY_ID` | - | AWS access key |
| `AWS_SECRET_ACCESS_KEY` | - | AWS secret key |

### Ollama Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `OLLAMA_HOST` | `http://localhost:11434` | Ollama server URL |

### Available Models

| Variable | Description |
|----------|-------------|
| `OLLAMA_MODELS` | Comma-separated list of Ollama models |
| `BEDROCK_MODELS` | Comma-separated list of Bedrock models |

## React Frontend Configuration

The React frontend loads environment variables from `web/.env` file.

### Core Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_API_BASE_URL` | `http://localhost:8084/api` | Backend API URL |
| `VITE_WEBSOCKET_URL` | `ws://localhost:8084/api/v1/chat/stream` | WebSocket URL |
| `VITE_DEFAULT_LLM_PROVIDER` | `bedrock` | Default LLM provider |
| `VITE_DEFAULT_MODEL` | `anthropic.claude-3-5-sonnet-20240620-v1:0` | Default model |

### Development Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_DEV_MODE` | `true` | Enable development features |
| `VITE_LOG_LEVEL` | `debug` | Log level: `debug`, `info`, `warn`, `error` |

## Usage Examples

### Using Bedrock with eu-central-1

```env
# .env
LLM_PROVIDER=bedrock
LLM_MODEL=anthropic.claude-3-5-sonnet-20240620-v1:0
AWS_REGION=eu-central-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
```

### Using Ollama locally

```env
# .env
LLM_PROVIDER=ollama
LLM_MODEL=qwen2.5:1.5b
OLLAMA_HOST=http://localhost:11434
```

### Starting the servers

```bash
# Start Go backend (reads from .env)
go run cmd/multiagent/main.go

# Start React frontend (reads from web/.env)
cd web && npx vite
```

### Command line overrides

You can still override environment variables with command line flags:

```bash
# Override LLM provider and model
go run cmd/multiagent/main.go -llm=ollama -model=qwen2.5:1.5b

# Override port
go run cmd/multiagent/main.go -port=8085
```

## Priority Order

Configuration is loaded in this order (highest priority first):

1. **Command line flags** (highest priority)
2. **Environment variables from .env file**
3. **Default values** (lowest priority)

This means you can set defaults in .env files and override them with command line flags when needed.

## Security Notes

- Never commit `.env` files with real credentials to version control
- Use `.env.example` files to document required variables
- Consider using AWS IAM roles in production instead of access keys
- Rotate credentials regularly

## Troubleshooting

### Bedrock Access Issues

If you get "AccessDeniedException":
1. Verify your AWS credentials are correct
2. Check that your region supports the model you're trying to use
3. Ensure your AWS account has access to the specific Bedrock model
4. Try a different model that's available in your region

### Environment Variables Not Loading

1. Make sure the `.env` file is in the project root
2. Check file permissions (should be readable)
3. Verify there are no syntax errors in the .env file
4. Restart the server after changing environment variables
