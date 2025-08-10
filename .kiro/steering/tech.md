# Technology Stack

## Backend (API)
- **Framework**: FastAPI with Python
- **Package Manager**: uv (Python package manager)
- **Database**: DuckDB for persistent storage
- **WebSocket**: Real-time streaming communication
- **Agent Framework**: LangChain with ReAct agent pattern
- **Environment**: Python with dotenv configuration

### Key Dependencies
- `fastapi` - Web framework
- `uvicorn` - ASGI server
- `langchain` + `langchain_openai` - LLM agent framework
- `duckdb` - Embedded database
- `websockets` - WebSocket support
- `pydantic` - Data validation
- `aiofiles` - Async file operations

## Frontend (Web)
- **Framework**: React 18 with TypeScript
- **Build Tool**: Vite
- **Package Manager**: npm/pnpm (lockfiles present for both)
- **UI Library**: shadcn/ui with Radix UI primitives
- **Styling**: Tailwind CSS
- **State Management**: TanStack Query (React Query)
- **Routing**: React Router DOM
- **Code Editor**: Monaco Editor

### Key Dependencies
- `react` + `react-dom` - Core React
- `typescript` - Type safety
- `@radix-ui/*` - UI primitives
- `tailwindcss` - Styling
- `@tanstack/react-query` - Server state management
- `@monaco-editor/react` - Code editor component
- `lucide-react` - Icons

## Development Commands

### Backend (from `/api` directory)
```bash
# Install dependencies
uv sync

# Start development server
uv run uvicorn main:app --host localhost --port 8084 --reload

# Run tests
uv run python test_backend.py

# Quick test script
uv run python -c "from app.database.service import db_service; print(db_service.get_all_projects())"
```

### Frontend (from `/web` directory)
```bash
# Install dependencies
npm install
# or
pnpm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Lint code
npm run lint
```

## Configuration
- **Backend**: Environment variables in `.env` (copy from `.env.sample`)
- **Frontend**: Environment variables in `.env` (copy from `.env.example`)
- **CORS**: Configured for ports 3000, 5173, 8080
- **Default Ports**: Backend on 8084, Frontend on 8080