# Product Overview

This is a Code Editing Agent platform that combines AI-powered chat capabilities with project management and code generation. The system consists of:

## Core Features
- **AI Chat Interface**: Real-time streaming chat with LLM models for code assistance
- **Project Management**: Create, manage, and organize coding projects with persistent storage
- **Multi-Model Support**: Integration with various LLM providers (OpenRouter, Anthropic, Gemini, Ollama, Bedrock)
- **Token Tracking**: Monitor and track LLM usage and costs across conversations
- **Conversation History**: Persistent chat sessions with project-specific context
- **Docker Integration**: Container management for project environments

## Architecture
The platform uses a FastAPI backend with DuckDB for persistence, connected to a React frontend via WebSocket streaming for real-time communication. The system is designed to help developers build applications through conversational AI assistance while maintaining project context and history.

## Target Use Case
Developers who want an AI-powered coding assistant that can maintain context across sessions, manage multiple projects, and provide real-time streaming responses while tracking usage and costs.