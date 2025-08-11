---
inclusion: always
---

# Database Architecture & Patterns

## Database Technology
- **Engine**: DuckDB embedded database with persistent storage
- **Connection**: Singleton pattern with automatic reconnection and WAL recovery
- **Location**: `api/data/database.db` with optional reset on startup
- **Schema**: Project-scoped operations with foreign key relationships

## Core Tables & Relationships

### Projects Table
- **Primary Key**: `id` (UUID)
- **Required Fields**: `name` (unique), `template`, `status`
- **Optional Fields**: `docker_container`, `port`, `user_id`, `github_repo_id`, `vercel_deployment_id`
- **Timestamps**: `created_at`, `updated_at` (auto-managed)

### Conversation Messages Table
- **Primary Key**: `id` (UUID)
- **Foreign Keys**: `project_id` (required), `token_usage_id` (optional)
- **Required Fields**: `role` (user/assistant), `content`, `message_type`
- **Optional Fields**: `model`, `provider`, `session_id` (legacy)
- **Project Scoping**: ALL queries MUST filter by `project_id`

### Token Usage Table
- **Primary Key**: `id` (UUID)
- **Foreign Key**: `project_id` (optional for backward compatibility)
- **Required Fields**: `session_id`, `model`, `provider`, `request_type`
- **Token Fields**: `input_tokens`, `output_tokens`, `total_tokens`

### Users Table (Authentication)
- **Primary Key**: `id` (UUID)
- **Required Fields**: `email` (unique), `name`
- **Integration Fields**: `github_username`, `github_token`, `vercel_token`, `vercel_team_id`
- **OAuth Fields**: `google_id`, `avatar_url`

### Integration Tables
- **GitHub Repositories**: Links projects to GitHub repos with clone URLs
- **Vercel Deployments**: Tracks deployment status and URLs

## Critical Database Rules

### Project Scoping
- **MANDATORY**: Every database operation MUST be project-scoped using `project_id`
- **Conversation History**: Always filter messages by `project_id`, never by `session_id` alone
- **Token Tracking**: Associate usage records with specific projects for accurate billing
- **Data Isolation**: Projects are completely isolated from each other

### Connection Management
- **Singleton Pattern**: Use global `db_service` instance from `app.database.service`
- **Auto-Recovery**: Connection automatically handles WAL corruption and database invalidation
- **Retry Logic**: All queries use `_execute_with_retry()` for resilience
- **Transaction Safety**: Commit after write operations, rollback on errors

### Model Patterns
- **Pydantic Integration**: Use Pydantic models for API validation (`ChatRequest`, `ProjectCreate`)
- **Internal Models**: Use plain Python classes for database entities (`Project`, `ConversationMessage`)
- **Create/Update Separation**: Separate models for creation vs. retrieval operations
- **UUID Generation**: All primary keys use `str(uuid.uuid4())`

## Database Service Patterns

### Async Operations
```python
# Correct: Use async for user operations
async def create_user(self, user_data: UserCreate) -> User:
    # Implementation with async/await

# Correct: Sync for project operations (legacy compatibility)
def create_project(self, project_data: ProjectCreate) -> Project:
    # Implementation without async
```

### Error Handling
- **Retry Logic**: Use `_execute_with_retry()` for all database operations
- **Connection Recovery**: Automatic reconnection on database invalidation
- **Foreign Key Constraints**: Handle cascading deletes properly
- **Transaction Rollback**: Ensure data consistency on errors

### Query Patterns
- **Project Messages**: `WHERE project_id = ? AND message_type = 'chat' ORDER BY created_at ASC`
- **Token Usage**: `WHERE project_id = ? ORDER BY created_at DESC`
- **Global Stats**: Aggregate functions with `COALESCE()` for null safety
- **Indexes**: Use existing indexes on `project_id`, `session_id`, `created_at`

## Integration Guidelines

### Service Layer Usage
```python
from app.database.service import db_service

# Project operations
project = db_service.create_project(project_data)
messages = db_service.get_project_messages(project_id)

# Token tracking
usage = db_service.create_token_usage(usage_data)
stats = db_service.get_project_token_usage(project_id)
```

### Model Imports
```python
from app.database.models import (
    Project, ProjectCreate,
    ConversationMessage, ConversationMessageCreate,
    TokenUsage, TokenUsageCreate
)
```

## Performance Considerations
- **Batch Operations**: Use transactions for multiple related operations
- **Index Usage**: Leverage existing indexes on frequently queried columns
- **Connection Reuse**: Single connection instance across application lifecycle
- **Memory Management**: DuckDB handles memory efficiently for embedded use
- **Query Optimization**: Use prepared statements through parameterized queries


# Existing tables
- `projects` (primary key: `id`)
- `conversation_messages` (primary key: `id`, foreign key: `project_id`)
- `token_usage` (primary key: `id`, foreign key: `project_id`)
- `users` (primary key: `id`)
- `github_repositories` (primary key: `id`, foreign key: `project_id`)
- `vercel_deployments` (primary key: `id`, foreign key: `project_id`)
- `integrations` (primary key: `id`, foreign key: `project_id`)
