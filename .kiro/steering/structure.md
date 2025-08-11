---
inclusion: always
---

# Project Structure & Conventions

## Architecture Overview
This is a full-stack Code Editing Agent platform with FastAPI backend and React frontend, using WebSocket streaming for real-time AI interactions.

## Directory Structure
```
/
├── api/           # FastAPI backend with LangChain agents
├── web/           # React frontend with TypeScript
├── dock-route/    # Docker template management
├── bin/           # Binary executables
└── scripts/       # Setup utilities
```

## Backend Structure (`/api`)
- **Entry Point**: `main.py` - FastAPI app with CORS and WebSocket support
- **Agents**: `app/agents/` - LangChain ReAct agent with custom tools
- **API Routes**: `app/api/` - RESTful endpoints and WebSocket handlers
- **Database**: `app/database/` - DuckDB service layer with models
- **Prompts**: `app/prompts/` - LLM prompt templates
- **Config**: `app/config.py` - Environment and model configurations

## Frontend Structure (`/web`)
- **Components**: `src/components/` - React components with shadcn/ui
- **Pages**: `src/pages/` - Route-level components
- **Hooks**: `src/hooks/` - Custom React hooks for state management
- **Services**: `src/services/` - API clients and utilities
- **Types**: `src/types/` - TypeScript type definitions

## Code Conventions

### Backend (Python)
- **Naming**: Snake case for files/functions (`react_agent.py`, `get_projects()`)
- **Imports**: Relative imports within app (`from .database import service`)
- **Async**: Use async/await for all I/O operations
- **Error Handling**: Raise HTTPException with proper status codes
- **Database**: Always use project_id scoping for data operations

### Frontend (TypeScript)
- **Naming**: PascalCase for components (`ChatPanel.tsx`), camelCase for utilities
- **Imports**: Use `@/` alias for src imports (`import { api } from '@/services/api'`)
- **State**: TanStack Query for server state, useState for local state
- **Components**: Functional components with TypeScript interfaces
- **Styling**: Tailwind classes with shadcn/ui components

## Critical Patterns

### WebSocket Streaming
- All AI interactions MUST use WebSocket endpoints (`/ws/chat/{project_id}`)
- Stream responses in chunks with proper message formatting
- Handle connection lifecycle (connect, message, disconnect, error)

### Project Context
- Every operation requires `project_id` parameter
- Database queries MUST be scoped to specific projects
- Chat sessions are project-specific and persistent

### Agent Integration
- Use ReAct agent pattern with custom tools in `app/agents/tools.py`
- Tools should be atomic and focused on single operations
- Always validate tool inputs and handle errors gracefully

### Error Handling
- Backend: Use FastAPI HTTPException with descriptive messages
- Frontend: Use try/catch with user-friendly error displays
- WebSocket: Handle connection errors and reconnection logic

## File Organization Rules
- Keep components focused and single-purpose
- Group related functionality in dedicated directories
- Use index files for clean imports where appropriate
- Separate business logic from UI components