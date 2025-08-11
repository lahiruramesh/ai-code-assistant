# Code Editing Agent Platform

A full-stack AI-powered code editing and project management platform that combines conversational AI with real-time project development capabilities.

## 🚀 Overview

The Code Editing Agent Platform is a comprehensive solution for AI-assisted software development, featuring:

- **AI-Powered Chat Interface**: Real-time streaming conversations with LLM models for code assistance
- **Project Management**: Create, manage, and organize coding projects with persistent storage
- **Multi-Model Support**: Integration with various LLM providers (OpenRouter, Anthropic, Gemini, Ollama, Bedrock)
- **Token Tracking**: Monitor and track LLM usage and costs across conversations
- **Docker Integration**: Automated container management for project environments
- **OAuth Authentication**: Secure authentication with Google, GitHub, and Vercel integrations
- **Real-time Collaboration**: WebSocket-based streaming for instant AI responses

## 🏗️ Architecture

```
code-editing-agent/
├── api/                 # FastAPI backend with DuckDB
├── web/                 # React frontend with TypeScript
├── bin/                 # Binary executables
├── dock-route/          # Docker routing utilities
├── scripts/             # Setup and utility scripts
└── .kiro/               # Kiro AI assistant configuration
```

### Backend (FastAPI + Python)
- **Framework**: FastAPI with Python 3.11+
- **Database**: DuckDB for persistent storage
- **AI Integration**: LangChain with ReAct agent pattern
- **Real-time**: WebSocket streaming communication
- **Authentication**: OAuth with Google, GitHub, Vercel
- **Package Manager**: uv (Python package manager)

### Frontend (React + TypeScript)
- **Framework**: React 18 with TypeScript
- **Build Tool**: Vite for fast development
- **UI Library**: shadcn/ui with Radix UI primitives
- **Styling**: Tailwind CSS for modern design
- **State Management**: TanStack Query (React Query)
- **Code Editor**: Monaco Editor integration
- **Package Manager**: npm/pnpm

## 🚀 Quick Start

### Prerequisites
- Python 3.11+
- Node.js 18+
- Docker (for project containers)
- uv (Python package manager)

### 1. Clone the Repository
```bash
git clone <repository-url>
cd code-editing-agent
```

### 2. Backend Setup
```bash
cd api

# Install dependencies
uv sync

# Copy environment configuration
cp .env.sample .env

# Edit .env with your API keys
nano .env

# Start the backend server
uv run uvicorn main:app --host localhost --port 8084 --reload
```

### 3. Frontend Setup
```bash
cd web

# Install dependencies
npm install
# or
pnpm install

# Copy environment configuration
cp .env.example .env

# Start the development server
npm run dev
```

### 4. Access the Application
- **Frontend**: http://localhost:8080
- **Backend API**: http://localhost:8084
- **API Documentation**: http://localhost:8084/docs

## 🔧 Configuration

### Environment Variables

#### Backend (.env)
```bash
# LLM Configuration
OPENROUTER_API_KEY=your-openrouter-key
MODEL_NAME=anthropic/claude-3.5-sonnet

# Database
DATABASE_DIR=./data

# Project Management
PROJECTS_DIR=/tmp/codeagent/projects
DOCK_ROUTE_PATH=/path/to/dock-route

# OAuth Integration
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret

# CORS
WEB_URL=http://localhost:8080
```

#### Frontend (.env)
```bash
VITE_API_BASE_URL=http://localhost:8084/api/v1
```

## 🎯 Key Features

### AI-Powered Development
- **Conversational Interface**: Natural language interaction with AI models
- **Code Generation**: Automated code creation and modification
- **Project Scaffolding**: Instant project setup with templates
- **Real-time Streaming**: Live AI responses via WebSocket

### Project Management
- **Multi-Project Support**: Manage multiple projects simultaneously
- **File Management**: Browse, edit, and organize project files
- **Container Integration**: Automated Docker container management
- **Template System**: Pre-configured project templates

### Authentication & Integrations
- **OAuth Support**: Google, GitHub, Vercel authentication
- **GitHub Integration**: Repository creation and management
- **Vercel Deployment**: One-click deployment to Vercel
- **Token Tracking**: Monitor API usage and costs

### Developer Experience
- **Modern UI**: Clean, responsive interface with dark/light themes
- **Code Editor**: Full-featured Monaco editor integration
- **Real-time Updates**: Live project status and container management
- **Comprehensive Testing**: Full test coverage with mocks

## 🧪 Testing

### Backend Tests
```bash
cd api

# Run all tests
python run_tests.py

# Run specific test categories
python run_tests.py unit
python run_tests.py integration
python run_tests.py database

# Run with coverage
uv run pytest tests/ --cov=app --cov-report=html
```

### Frontend Tests
```bash
cd web

# Run tests
npm test

# Run with coverage
npm run test:coverage
```

## 📚 API Documentation

### Core Endpoints
- `GET /` - Welcome message and feature list
- `GET /health` - Health check with database status
- `POST /api/v1/chat/create-session` - Create new chat session
- `WS /api/v1/chat/stream/{project_id}` - WebSocket streaming
- `GET /api/v1/projects` - List all projects
- `POST /api/v1/projects/` - Create new project
- `GET /api/v1/projects/{id}` - Get project details
- `GET /api/v1/projects/{id}/conversations` - Get project conversations
- `GET /api/v1/models/all` - Get available AI models
- `GET /api/v1/tokens/stats` - Get token usage statistics

### Authentication Endpoints
- `GET /api/v1/auth/google/login` - Google OAuth login
- `GET /api/v1/auth/github/connect` - GitHub integration
- `GET /api/v1/auth/vercel/connect` - Vercel integration

## 🔄 Development Workflow

### 1. Start Development Environment
```bash
# Terminal 1: Backend
cd api && uv run uvicorn main:app --reload

# Terminal 2: Frontend
cd web && npm run dev

# Terminal 3: Watch tests
cd api && uv run pytest tests/ --watch
```

### 2. Make Changes
- Backend changes auto-reload with uvicorn
- Frontend changes auto-reload with Vite
- Tests run automatically on file changes

### 3. Test Changes
```bash
# Run backend tests
cd api && python run_tests.py

# Run frontend tests
cd web && npm test

# Integration testing
# Test full workflow through UI
```

## 🚢 Deployment

### Backend Deployment
```bash
cd api

# Build for production
uv build

# Deploy with Docker
docker build -t code-editing-agent-api .
docker run -p 8084:8084 code-editing-agent-api
```

### Frontend Deployment
```bash
cd web

# Build for production
npm run build

# Deploy to Vercel
vercel deploy

# Or deploy to any static hosting
# Serve the dist/ directory
```

## 🤝 Contributing

### Development Setup
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

### Code Standards
- **Backend**: Follow PEP 8 Python style guide
- **Frontend**: Use TypeScript strict mode
- **Testing**: Maintain >80% code coverage
- **Documentation**: Update README and API docs

### Commit Convention
```bash
feat: add new feature
fix: bug fix
docs: documentation update
test: add or update tests
refactor: code refactoring
style: formatting changes
```

## 📖 Documentation

- **API Documentation**: `/api/README.md`
- **Frontend Documentation**: `/web/README.md`
- **Test Documentation**: `/api/tests/README.md`
- **Deployment Guide**: `/docs/deployment.md`
- **Architecture Guide**: `/docs/architecture.md`

## 🐛 Troubleshooting

### Common Issues

#### Backend Issues
- **Import errors**: Run `uv sync` to install dependencies
- **Database errors**: Delete `data/database.db` to reset
- **Port conflicts**: Check if port 8084 is available
- **Docker issues**: Ensure Docker is running

#### Frontend Issues
- **Build errors**: Clear node_modules and reinstall
- **API connection**: Verify backend is running on port 8084
- **CORS errors**: Check WEB_URL in backend .env

#### Integration Issues
- **OAuth failures**: Verify client IDs and secrets
- **Container startup**: Check Docker daemon status
- **File permissions**: Ensure proper directory permissions

### Getting Help
1. Check the troubleshooting section in component READMEs
2. Review logs in browser console and terminal
3. Verify environment variable configuration
4. Test with minimal configuration first

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- **FastAPI** - Modern, fast web framework for building APIs
- **React** - A JavaScript library for building user interfaces
- **DuckDB** - In-process SQL OLAP database management system
- **LangChain** - Framework for developing applications with LLMs
- **shadcn/ui** - Beautifully designed components built with Radix UI
- **Tailwind CSS** - Utility-first CSS framework

## 🔗 Links

- **Live Demo**: [Coming Soon]
- **Documentation**: [API Docs](http://localhost:8084/docs)
- **Issues**: [GitHub Issues]
- **Discussions**: [GitHub Discussions]

---

**Built with ❤️ for developers who want AI-powered coding assistance**