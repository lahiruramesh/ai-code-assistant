# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a comprehensive AI-powered code editing and development platform consisting of three main components:

1. **Web Frontend** (`/web`): React-based UI using Vite, TypeScript, and shadcn/ui components
2. **API Service** (`/api`): Go-based multi-agent system for AI-driven code generation and editing
3. **Docker CLI Tool** (`/dock-route`): Command-line tool for deploying applications with automatic subdomain routing

## Build Commands

### Frontend (Web)
```bash
# Navigate to web directory
cd web

# Install dependencies
pnpm install

# Development server
pnpm run dev

# Build for production
pnpm run build

# Lint code
pnpm run lint

# Preview production build
pnpm run preview
```

### API Service
```bash
# Navigate to api directory
cd api

# Build the Go application
go build -o main .

# Build multiagent version
go build -o multiagent ./cmd/multiagent

# Run tests (various test scripts available)
./test_multiagent.sh
./test_tools.sh
./test_loop_manager.sh
./test_interactive.sh
./test_single_loop.sh
./test_dev_workflow.sh
```

### Docker CLI Tool
```bash
# Navigate to dock-route directory
cd dock-route

# Install dependencies
pnpm install
go mod tidy

# Build CLI
go build -o dock-route .

# Example usage
dock-route deploy nextjs my-app ./src
dock-route list containers
dock-route remove my-app
```

## Architecture Overview

### Multi-Agent System (`/api`)
The core AI system uses a multi-agent architecture with specialized agents:

- **Supervisor Agent**: Orchestrates task delegation and coordinates other agents
- **Code Editing Agent**: Handles file creation, modification, and code generation
- **React Agent**: Specialized for React/TypeScript development tasks

**Key Components:**
- `internal/pkg/agents/`: Agent implementation and coordination
- `internal/pkg/tools/`: Tool execution for file system operations
- `internal/pkg/llm/`: LLM service integration (supports Ollama and AWS Bedrock)
- `internal/pkg/docker/`: Docker container management
- `internal/pkg/database/`: Project data persistence (SQLite)

**Agent Communication:**
- Agents communicate through structured messages (`AgentMessage` type)
- Tasks are delegated based on agent specialization
- Tool calling is supported for file system operations

### Frontend Architecture (`/web`)
Built with React, TypeScript, and Vite using modern development patterns:

**Key Features:**
- Multi-page routing with React Router
- Real-time chat interface for AI interaction
- Monaco code editor integration
- File explorer with drag-and-drop support
- Preview panel for live development

**Tech Stack:**
- React 18 with TypeScript
- Vite for build tooling
- shadcn/ui component library
- Tailwind CSS for styling
- React Query for data fetching
- Monaco Editor for code editing

### Docker CLI Tool (`/dock-route`)
A comprehensive CLI for deploying web applications with automatic subdomain routing:

**Supported Frameworks:**
- Next.js (`nextjs`)
- React.js (`reactjs`) 
- Node.js (`nodejs`)

**Key Features:**
- Automatic subdomain generation (`preview-{name}.dock-route.local`)
- Docker container management
- Dynamic reverse proxy routing
- Template-based deployment system
- Hot-reloading for development

## Configuration

### Environment Variables
The API service supports environment configuration:
- `LLM_PROVIDER`: Set to `ollama` (default) or `bedrock`
- `LLM_MODEL`: Specify the model to use (defaults vary by provider)
- AWS credentials for Bedrock integration

### DNS Setup
For dock-route subdomain routing:
```bash
# Add to /etc/hosts
127.0.0.1 *.dock-route.localhost
```

### Docker Integration
The system requires Docker to be running for container management and deployment features.

## Testing

### API Testing
Multiple test scripts are available in `/api`:
- `test_multiagent.sh`: Tests multi-agent coordination
- `test_tools.sh`: Tests tool calling functionality
- `test_loop_manager.sh`: Tests conversation loop management
- `test_interactive.sh`: Tests interactive CLI mode
- `test_single_loop.sh`: Tests single conversation loop
- `test_dev_workflow.sh`: Tests development workflow integration

### Frontend Testing
Run the development server with `pnpm run dev` and test at `http://localhost:5173`

## Project Structure Notes

### File Operations
The tool system uses absolute paths for file operations. When working with files, ensure proper path handling:

- For CLI operations, paths are relative to the current working directory
- For agent operations, paths are relative to the project context directory
- Use `filepath.Abs()` for consistent path resolution

### Agent Coordination
When modifying agent behavior:
1. Update system prompts in `/api/prompts/`
2. Modify agent logic in `/internal/pkg/agents/`
3. Add new tools in `/internal/pkg/tools/`
4. Update coordinator logic in `/internal/pkg/agents/coordinator.go`

### Docker Templates
To add new deployment templates:
1. Create template directory in `/dock-route/internal/templates/data/[framework]`
2. Add `Dockerfile` and `template.yaml`
3. Update the template manager to recognize the new framework type