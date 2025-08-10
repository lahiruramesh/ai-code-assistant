# Project Structure

## Root Directory Layout
```
/
├── api/           # FastAPI backend
├── web/           # React frontend  
├── bin/           # Binary executables
├── dock-route/    # Docker routing utilities
└── scripts/       # Setup and utility scripts
```

## Backend Structure (`/api`)
```
api/
├── app/
│   ├── agents/          # LangChain ReAct agent implementation
│   │   ├── react_agent.py
│   │   └── tools.py
│   ├── api/             # FastAPI route handlers
│   │   ├── projects.py  # Project management endpoints
│   │   └── streaming.py # WebSocket streaming endpoints
│   ├── database/        # DuckDB integration
│   │   ├── connection.py
│   │   ├── models.py    # Data models and schemas
│   │   └── service.py   # Database service layer
│   ├── prompts/         # LLM prompts and templates
│   │   └── react_prompts.py
│   └── utils/           # Utility functions
│       └── docker_route.py
├── main.py            # FastAPI application entry point
├── pyproject.toml     # Python dependencies (uv)
└── .env              # Environment configuration
```

## Frontend Structure (`/web`)
```
web/
├── src/
│   ├── components/      # React components
│   │   ├── ui/         # shadcn/ui components
│   │   ├── AIBuilder.tsx
│   │   ├── ChatPanel.tsx
│   │   ├── CodeEditor.tsx
│   │   ├── FileExplorer.tsx
│   │   └── ...
│   ├── pages/          # Route components
│   │   ├── ChatPage.tsx
│   │   ├── Index.tsx
│   │   └── NotFound.tsx
│   ├── services/       # API and utility services
│   │   ├── eventManager.ts
│   │   └── soundService.ts
│   ├── hooks/          # Custom React hooks
│   ├── lib/            # Utility functions
│   ├── config/         # Configuration files
│   └── main.tsx       # React application entry point
├── public/            # Static assets
├── package.json       # Node.js dependencies
├── vite.config.ts     # Vite configuration
├── tailwind.config.ts # Tailwind CSS configuration
└── .env              # Environment variables
```

## Key Architectural Patterns

### Backend Patterns
- **Layered Architecture**: API routes → Services → Database
- **Dependency Injection**: Database connections and services
- **WebSocket Streaming**: Real-time communication pattern
- **Agent Pattern**: LangChain ReAct agent with tools
- **Repository Pattern**: Database service abstraction

### Frontend Patterns
- **Component Composition**: shadcn/ui + custom components
- **Route-based Code Splitting**: React Router with page components
- **Server State Management**: TanStack Query for API state
- **Custom Hooks**: Reusable logic extraction
- **Service Layer**: API communication abstraction

### File Naming Conventions
- **Backend**: Snake case (`react_agent.py`, `database_service.py`)
- **Frontend**: PascalCase for components (`ChatPanel.tsx`), camelCase for utilities (`eventManager.ts`)
- **Configuration**: Lowercase with dots (`.env`, `vite.config.ts`)

### Import Patterns
- **Frontend**: Use `@/` alias for src imports
- **Backend**: Relative imports within app structure
- **Environment**: Use dotenv for configuration management