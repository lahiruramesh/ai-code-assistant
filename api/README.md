# API-v2 Backend - DuckDB Integration

## Overview

This is the enhanced API-v2 backend for the Code Editing Agent with DuckDB integration. It provides:

- 🗄️ **DuckDB Database**: Persistent storage for projects, conversations, and token usage
- 🔌 **WebSocket Streaming**: Real-time communication with frontend
- 🤖 **ReAct Agent**: LangChain-powered agent with project-aware tools
- 📊 **Token Tracking**: Monitor LLM usage and costs
- 💬 **Conversation History**: Persistent chat sessions

## Features

### Database Schema
- **Projects**: Store project metadata and configurations
- **Conversations**: Track chat sessions and messages
- **Token Usage**: Monitor LLM consumption
- **Containers**: Manage Docker container information

### API Endpoints
- `GET /` - Welcome message and feature list
- `GET /health` - Health check with database status
- `POST /api/v1/chat/create-session` - Create new chat session
- `WS /api/v1/chat/stream/{project_id}` - WebSocket streaming
- `GET /api/v1/projects` - List all projects
- `POST /api/v1/projects/` - Create new project
- `GET /api/v1/projects/{id}` - Get project details
- `GET /api/v1/projects/{id}/files` - Get project file structure
- `GET /api/v1/projects/{id}/conversations` - Get project conversations

## Quick Start

### 1. Setup Environment
```bash
# Copy environment configuration
cp .env.sample .env

# Edit .env with your API keys
nano .env
```

### 2. Start the Server
```bash
uv sync
uv run uvicorn main:app --host localhost --port 8084 --reload
```

### 3. Test the API
```bash
# Health check
curl http://localhost:8084/health

# Create a chat session
curl -X POST http://localhost:8084/api/v1/chat/create-session \
  -H "Content-Type: application/json" \
  -d '{"message": "Create a React app with TypeScript"}'

# List projects
curl http://localhost:8084/api/v1/projects/
```

## Frontend Integration

The backend is designed to work with the React frontend in the `../web` directory. The WebSocket endpoint provides real-time streaming of agent responses.

### WebSocket Usage
```javascript
const ws = new WebSocket('ws://localhost:8084/api/v1/chat/stream/1');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);
};

ws.send(JSON.stringify({
  message: "Help me build a web application",
  model: "gpt-4",
  provider: "openai"
}));
```

## Development

### Project Structure
```
api-v2/
├── app/
│   ├── agents/          # ReAct agent and tools
│   ├── api/             # FastAPI routes
│   ├── database/        # DuckDB models and services
│   └── prompts/         # Agent prompts
├── data/                # Database storage
├── projects/            # Generated projects
├── main.py              # FastAPI application
├── test_backend.py      # Test suite
└── start_server.sh      # Startup script
```

### Testing
```bash
# Run backend tests
uv run python test_backend.py

# Check database
uv run python -c "from app.database.service import db_service; print(db_service.get_all_projects())"
```

## Configuration

### Environment Variables
- `OPENROUTER_API_KEY`: Your OpenRouter API key
- `OPENROUTER_API_BASE`: OpenRouter API base URL
- `MODEL_NAME`: LLM model to use
- `DATABASE_PATH`: Path to DuckDB database file
- `PROJECTS_DIR`: Directory for generated projects

### Database Location
The DuckDB database is stored at `./data/database.db` by default.

## Frontend Connection

To connect with the frontend:

1. Ensure the frontend is configured to use `http://localhost:8084`
2. The WebSocket endpoint supports project-specific conversations
3. CORS is configured for common frontend ports (3000, 5173, 8080)

## Troubleshooting

### Common Issues
- **Import errors**: Run `uv sync` to install dependencies
- **Database errors**: Delete `data/database.db` to reset
- **Connection refused**: Check if port 8084 is available
- **CORS errors**: Verify frontend URL in CORS configuration

### Logs
The server logs all WebSocket connections and database operations for debugging.

## Next Steps

1. Start the backend: `./start_server.sh`
2. Start the frontend from `../web` directory
3. Create a new chat session and test the integration
4. Explore the project file management features
