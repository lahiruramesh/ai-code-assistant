# AI Code Editing Agent - Deployment & Usage Guide

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- Docker (for dock-route deployments)
- Git

### 1. Environment Setup

```bash
# Clone the repository
git clone https://github.com/lahiruramesh/code-editing-agent.git
cd code-editing-agent

# Setup API environment
cd api
cp .env.example .env
# Edit .env with your API keys (see Environment Configuration section)

# Setup Frontend environment  
cd ../web
cp .env.example .env
# Verify VITE_API_BASE_URL=http://localhost:8084/api/v1
```

### 2. Build Required Binaries

```bash
cd api

# Build all binaries
go build -o bin/server cmd/server/main.go
go build -o bin/multiagent cmd/multiagent/main.go

# Build dock-route CLI
cd ../dock-route
go build -o ../api/bin/dock-route .
```

### 3. Start the System

#### Option A: Server Mode (Recommended for Development)

```bash
# Terminal 1: Start Backend API Server
cd api
./bin/server

# Terminal 2: Start Frontend Development Server
cd web
npm install
npm run dev

# Terminal 3: Start Multi-Agent Server (for chat functionality)
./start_server.sh
```

#### Option B: CLI Mode (For Testing Agent Workflows)

```bash
# Interactive CLI testing
./test_agent.sh
```

### 4. Access the Application

- **Frontend**: http://localhost:8080
- **Backend API**: http://localhost:8084
- **Multi-Agent Server**: http://localhost:8084 (same as backend)

## 🔧 Environment Configuration

### API Keys Required

Add these to `api/.env`:

```bash
# OpenRouter (Get from: https://openrouter.ai/keys)
OPENROUTER_API_KEY=your_openrouter_key_here

# Google Gemini (Get from: https://makersuite.google.com/app/apikey)
GEMINI_API_KEY=your_gemini_key_here

# Anthropic Claude (Get from: https://console.anthropic.com/)
ANTHROPIC_API_KEY=your_anthropic_key_here

# AWS Bedrock (Use your AWS credentials)
AWS_ACCESS_KEY_ID=your_aws_access_key
AWS_SECRET_ACCESS_KEY=your_aws_secret_key
AWS_REGION=us-east-1
```

### LLM Provider Selection

Set your preferred provider in `api/.env`:

```bash
# Choose from: ollama, bedrock, openrouter, gemini, anthropic
LLM_PROVIDER=openrouter
LLM_MODEL=deepseek/deepseek-chat-v3-0324:free
```

## 🏗️ Architecture Overview

### Components

1. **Backend API Server** (`bin/server`):
   - Serves the web interface
   - Provides REST API endpoints
   - Handles model configuration and token tracking

2. **Multi-Agent System** (`bin/multiagent`):
   - Processes project creation requests
   - Coordinates between specialized agents:
     - **Supervisor Agent**: Project planning and task delegation
     - **Code Editing Agent**: File operations, Docker management
     - **React Agent**: Component generation and React expertise

3. **Dock-Route CLI** (`bin/dock-route`):
   - Docker deployment automation
   - Subdomain routing for projects
   - Development environment with hot reloading

4. **Frontend Interface**:
   - Model selection and configuration
   - Chat interface for project requests
   - Token usage tracking
   - Project file management

### Workflow

1. **User Request**: Enter project requirements in chat interface
2. **Agent Coordination**: Supervisor agent analyzes and delegates tasks
3. **Project Creation**: Code editing agent creates project structure
4. **Docker Deployment**: Projects deployed using dock-route
5. **Live Development**: Hot reloading enabled for iterative development

## 🧪 Testing the Complete Workflow

### Test 1: Basic Agent Communication

```bash
cd /tmp/aiagent
echo "Create a simple React counter app" | /path/to/api/bin/multiagent -mode=cli -llm=openrouter
```

### Test 2: Frontend Integration

1. Access http://localhost:8080
2. Select your preferred LLM provider in Model Configuration
3. Enter project request: "Create a React todo app with add, edit, delete functionality"
4. Monitor token usage and project progress

### Test 3: Docker Deployment

```bash
# Deploy a project using dock-route
./api/bin/dock-route deploy reactjs my-todo-app /tmp/aiagent/my-todo-app --host-port 8085

# Access deployed app
open http://preview-my-todo-app.dock-route.local:8080
```

## 🔍 Troubleshooting

### Common Issues

1. **Port Conflicts**:
   ```bash
   # Kill processes on ports 8080, 8084
   lsof -ti:8080 | xargs kill -9
   lsof -ti:8084 | xargs kill -9
   ```

2. **Missing Dependencies**:
   ```bash
   # Install Go dependencies
   cd api && go mod tidy
   
   # Install Node dependencies
   cd web && npm install
   ```

3. **API Key Issues**:
   - Verify API keys in `api/.env`
   - Check provider-specific documentation for key format
   - Test connection: `curl -H "Authorization: Bearer YOUR_KEY" API_ENDPOINT`

4. **Agent Timeout Issues**:
   - Increase timeout in agent loop manager
   - Check LLM provider rate limits
   - Monitor token usage and quotas

### Debug Mode

Enable verbose logging:

```bash
# API Server debug
DEBUG=true ./bin/server

# Multi-agent debug
./bin/multiagent -mode=cli -llm=openrouter -verbose

# Frontend debug
VITE_LOG_LEVEL=debug npm run dev
```

### Log Files

Monitor logs for issues:

```bash
# API logs
tail -f /tmp/aiagent/api.log

# Agent workflow logs
tail -f /tmp/aiagent/agents.log

# Docker deployment logs
docker logs preview-container-name
```

## 📁 Project Structure

```
code-editing-agent/
├── api/                    # Backend Go application
│   ├── cmd/               # Command-line applications
│   ├── internal/          # Internal packages
│   ├── prompts/           # Agent system prompts
│   ├── bin/               # Built binaries
│   └── .env               # Environment configuration
├── web/                   # Frontend React application
│   ├── src/               # Source code
│   ├── public/            # Static files
│   └── .env               # Frontend environment
├── dock-route/            # Docker deployment CLI
│   ├── cmd/               # CLI commands
│   ├── internal/          # Docker management
│   └── templates/         # Project templates
├── templates/             # Project templates
│   ├── react-shadcn-template/
│   └── nextjs-shadcn-template/
└── scripts/               # Deployment scripts
```

## 🚀 Production Deployment

### Docker Deployment

```bash
# Build production images
docker build -t ai-agent-api ./api
docker build -t ai-agent-web ./web

# Run with docker-compose
docker-compose up -d
```

### Environment Variables

Set production environment variables:

```bash
# Production API URL
VITE_API_BASE_URL=https://your-domain.com/api/v1

# Production LLM configuration
LLM_PROVIDER=openrouter
LLM_MODEL=anthropic/claude-3.5-sonnet

# Security settings
API_CORS_ORIGINS=https://your-domain.com
JWT_SECRET=your-jwt-secret
```

### Monitoring

Monitor system health:

```bash
# API health check
curl http://localhost:8084/health

# Agent system status
curl http://localhost:8084/api/v1/agents/status

# Model availability
curl http://localhost:8084/api/v1/models/all
```

## 🛠️ Development

### Adding New LLM Providers

1. Add provider to `llm/service.go`
2. Update environment configuration
3. Add to frontend model selector
4. Test integration

### Adding New Agent Types

1. Create agent prompt in `prompts/`
2. Add agent type to `agents/types.go`
3. Update coordinator initialization
4. Test agent communication

### Contributing

1. Fork the repository
2. Create feature branch
3. Add tests for new functionality
4. Submit pull request

## 📞 Support

For issues and questions:

1. Check the troubleshooting section
2. Review logs for error details
3. Test with different LLM providers
4. Verify environment configuration

---

*Last updated: August 8, 2025*
