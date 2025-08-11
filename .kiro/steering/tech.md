---
inclusion: always
---

# Technology Stack & Development Guidelines

## Backend (API) - Python/FastAPI
- **Framework**: FastAPI with async/await patterns
- **Package Manager**: `uv` (REQUIRED - use `uv run` for all Python commands)
- **Database**: DuckDB with project-scoped queries
- **WebSocket**: Real-time streaming for all AI interactions
- **Agent Framework**: LangChain ReAct pattern with custom tools
- **Port**: 8084 (default)

### Critical Backend Rules
- ALL database operations MUST be project-scoped using `project_id`
- Use async/await for I/O operations (file, database, HTTP)
- WebSocket streaming is REQUIRED for AI chat responses
- Import patterns: `from app.database.service import db_service`
- Error handling: Raise `HTTPException` with descriptive messages

### Key Backend Dependencies
- `fastapi` + `uvicorn` - Web framework and ASGI server
- `langchain` + `langchain_openai` - LLM agent framework
- `duckdb` - Embedded database (project-scoped queries)
- `pydantic` - Data validation and serialization
- `aiofiles` - Async file operations

## Frontend (Web) - React/TypeScript
- **Framework**: React 18 with TypeScript (strict mode)
- **Build Tool**: Vite with hot reload
- **Package Manager**: pnpm (preferred) or npm
- **UI**: shadcn/ui components with Radix primitives
- **Styling**: Tailwind CSS utility classes
- **State**: TanStack Query for server state, useState for local
- **Port**: 8080 (default)

### Critical Frontend Rules
- Use `@/` import alias for src imports: `import { api } from '@/services/api'`
- ALL AI interactions MUST use WebSocket endpoints (`/ws/chat/{project_id}`)
- Components should be functional with TypeScript interfaces
- Handle WebSocket connection lifecycle (connect, message, disconnect, error)
- Use TanStack Query for server state management

### Key Frontend Dependencies
- `react` + `react-dom` + `typescript` - Core framework
- `@radix-ui/*` + `tailwindcss` - UI components and styling
- `@tanstack/react-query` - Server state management
- `@monaco-editor/react` - Code editor integration
- `lucide-react` - Icon library

## Development Workflow

### Backend Commands (run from `/api`)
```bash
# Install dependencies
uv sync

# Start development server
uv run uvicorn main:app --host localhost --port 8084 --reload

# Database test
uv run python -c "from app.database.service import db_service; print(db_service.get_all_projects())"
```

### Frontend Commands (run from `/web`)
```bash
# Install dependencies
pnpm install

# Start development server
pnpm run dev

# Build for production
pnpm run build
```

## Configuration Requirements
- **Backend**: Copy `api/.env.example` to `api/.env` and configure
- **Frontend**: Copy `web/.env.example` to `web/.env` and configure
- **CORS**: Pre-configured for development ports (3000, 5173, 8080)
- **WebSocket**: Ensure proper connection handling for streaming responses

## Architecture Patterns
- **Project Context**: Every operation requires `project_id` parameter
- **Streaming First**: Use WebSocket streaming for all AI interactions
- **Error Boundaries**: Implement proper error handling at component and API levels
- **Async Operations**: Use async/await patterns consistently across the stack