---
inclusion: always
---

# Product Overview

This is a Code Editing Agent platform that enables developers to build applications through AI-powered conversations with persistent project context and real-time streaming responses.

## Core Features & Behavior
- **AI Chat Interface**: Real-time streaming chat with LLM models for code assistance
- **Project Management**: Create, manage, and organize coding projects with persistent DuckDB storage
- **Multi-Model Support**: Integration with OpenRouter, Anthropic, Gemini, Ollama, and Bedrock
- **Token Tracking**: Monitor LLM usage and costs across conversations
- **Conversation History**: Persistent chat sessions with project-specific context
- **Docker Integration**: Container management for project environments with template support

## Key Product Conventions
- **Streaming First**: All AI responses use WebSocket streaming for real-time feedback
- **Project Context**: Every conversation is tied to a specific project for context persistence
- **Cost Awareness**: Token usage and costs are tracked and displayed to users
- **Multi-Model Flexibility**: Users can switch between different LLM providers within conversations
- **Docker Templates**: Support for React.js, Node.js, and Next.js project templates with containerization

## User Experience Patterns
- **Conversational Development**: Users interact through natural language to build and modify code
- **Session Persistence**: Chat history and project state persist across browser sessions
- **Real-time Feedback**: Streaming responses provide immediate visual feedback during AI processing
- **Project Isolation**: Each project maintains separate conversation history and context
- **Template-Based Setup**: New projects can be initialized from predefined Docker templates

## Technical Constraints
- **WebSocket Required**: All AI interactions must use WebSocket streaming endpoints
- **Project Scoping**: All database operations should be scoped to specific projects
- **Token Limits**: Respect model-specific token limits and provide usage feedback
- **Docker Dependencies**: Project environments rely on Docker for isolation and consistency