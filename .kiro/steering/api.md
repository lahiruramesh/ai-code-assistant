---
inclusion: always
---

# API Design & Implementation Guidelines

## Core API Principles
- **Project-Scoped Operations**: All API endpoints MUST operate within project context using `project_id`
- **WebSocket Streaming**: Real-time AI interactions use WebSocket endpoints (`/ws/chat/{project_id}`)
- **Consistent Error Handling**: Use FastAPI `HTTPException` with descriptive messages and proper status codes
- **Async/Await Pattern**: Use async operations for I/O-bound tasks (database, HTTP, file operations)
- **Pydantic Validation**: All request/response models use Pydantic for validation and serialization

## API Structure & Routing

### Router Organization
```python
# Import pattern for API routers
from app.api import streaming, projects, auth, github, vercel, models, tokens

# Router registration with consistent prefixes
app.include_router(streaming.router, prefix="/api/v1/chat", tags=["Chat"])
app.include_router(projects.router, prefix="/api/v1/projects", tags=["Projects"])
app.include_router(auth.router, prefix="/api/v1", tags=["Authentication"])
```

### Endpoint Naming Conventions
- **RESTful Resources**: `/api/v1/projects`, `/api/v1/projects/{project_id}`
- **WebSocket Streams**: `/api/v1/chat/stream/{project_id}`
- **Nested Resources**: `/api/v1/projects/{project_id}/conversations` (for chat messages)
- **Actions**: `/api/v1/chat/{session_id}/cancel`

## Request/Response Patterns

### Standard Response Format
```python
# Success responses use JSONResponse with consistent structure
return JSONResponse(content={
    "message": "Operation completed successfully",
    "data": {...},
    "metadata": {...}
}, status_code=200)

# Error responses use HTTPException
raise HTTPException(status_code=404, detail="Resource not found")
```

### Pydantic Models
- **Request Models**: Use `Create` suffix (`ProjectCreate`, `UserCreate`)
- **Response Models**: Use entity name (`Project`, `User`)
- **Optional Fields**: Use `Optional[Type] = None` for nullable fields
- **Validation**: Leverage Pydantic validators for complex validation logic

### WebSocket Message Format
```python
# Standard WebSocket message structure
{
    "type": "message_type",  # Required: message_received, agent_response, completion, status
    "content": "message_content",
    "session_id": "uuid",
    "project_id": "uuid",
    "metadata": {...}  # Optional additional data
}
```

## Database Integration Patterns

### Service Layer Usage
```python
from app.database.service import db_service

# Always use project-scoped operations
project = db_service.get_project_by_id(project_id)
messages = db_service.get_project_messages(project_id)
```

### Error Handling
- **Not Found**: Return 404 with descriptive message
- **Validation Errors**: Return 400 with field-specific errors
- **Server Errors**: Return 500 with sanitized error message
- **Database Errors**: Use try/catch with proper rollback

## Authentication & Authorization

### OAuth Integration
- **Google OAuth**: `/api/v1/auth/google/login` → `/api/v1/auth/google/callback`
- **GitHub OAuth**: `/api/v1/auth/github/login` → `/api/v1/auth/github/callback`
- **Token Exchange**: Always validate tokens with provider APIs
- **User Creation**: Automatic user creation on successful OAuth

### Integration Patterns
```python
# Async HTTP client pattern for external APIs
async with httpx.AsyncClient() as client:
    response = await client.get(url, headers=headers)
    if response.status_code != 200:
        raise HTTPException(status_code=400, detail="External API error")
```

## Streaming & Real-time Features

### WebSocket Lifecycle
1. **Connection**: Accept WebSocket and validate project access
2. **Session Management**: Generate session ID and store context
3. **Message Processing**: Stream responses in chunks
4. **Error Handling**: Close connection with appropriate codes
5. **Cleanup**: Store final messages and token usage

### Streaming Response Pattern
```python
async for chunk in agent.stream_response(message, project_path, container):
    if isinstance(chunk, dict) and chunk.get("type") == "content":
        await websocket.send_json({
            "type": "agent_response",
            "content": chunk.get("content", ""),
            "session_id": session_id,
            "project_id": project_id
        })
```

## Token Usage & Monitoring

### Usage Tracking
- **Per Request**: Track input/output tokens for each AI interaction
- **Project Scoped**: Associate usage with specific projects
- **Provider Agnostic**: Support multiple LLM providers (OpenRouter, Anthropic, etc.)
- **Cost Calculation**: Store token counts for billing/monitoring

### Statistics Endpoints
- **Session Usage**: `/api/v1/tokens/usage/{session_id}`
- **Project Usage**: `/api/v1/tokens/project/{project_id}`
- **Global Stats**: `/api/v1/tokens/stats`

## File & Project Management

### Project Operations
- **Creation**: Auto-generate project names, deploy Docker containers
- **File Access**: Secure file reading with path validation
- **Directory Listing**: Recursive file tree generation
- **Container Management**: Start/stop Docker containers as needed

### Security Patterns
```python
# Path validation for file access
project_path = os.path.abspath(os.path.join(PROJECTS_DIR, project.name))
full_path = os.path.abspath(full_path)
if not full_path.startswith(project_path):
    raise HTTPException(status_code=403, detail="Access denied")
```

## Configuration & Environment

### Environment Variables
- **API Keys**: `OPENROUTER_API_KEY`, `GOOGLE_CLIENT_ID`, `GITHUB_CLIENT_SECRET`
- **Paths**: `PROJECTS_DIR`, `DATABASE_DIR`, `DOCK_ROUTE_PATH`
- **URLs**: `WEB_URL` for CORS and OAuth redirects
- **Models**: `MODEL_NAME`, `LLM_PROVIDER` for default configurations

### CORS Configuration
```python
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:8080", WEB_URL],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"]
)
```

## Health & Monitoring

### Health Checks
- **Database Connection**: Test DuckDB connectivity
- **Container Status**: Verify Docker container health
- **External APIs**: Validate API key configurations

### Logging & Debugging
- **Structured Logging**: Use consistent log formats
- **Error Context**: Include request IDs and user context
- **Performance Metrics**: Track response times and token usage
