# API Test Suite

This directory contains comprehensive unit and integration tests for the Code Editing Agent API.

## Test Structure

```
tests/
├── conftest.py              # Test configuration and fixtures
├── test_projects.py         # Project management API tests
├── test_streaming.py        # WebSocket streaming and chat tests
├── test_models_tokens.py    # Models and token usage API tests
├── test_database_service.py # Database service layer tests
├── test_auth.py            # Authentication and OAuth tests
├── test_main.py            # Main application and health check tests
├── test_integration.py     # End-to-end integration tests
└── README.md               # This file
```

## Test Categories

### Unit Tests
- **Projects API** (`test_projects.py`): CRUD operations, file management, conversations, security
- **Streaming API** (`test_streaming.py`): WebSocket connections, chat sessions, agent responses
- **Models & Tokens** (`test_models_tokens.py`): Model configuration, token tracking, usage statistics
- **Database Service** (`test_database_service.py`): Database operations, error handling, retry logic
- **Authentication** (`test_auth.py`): OAuth flows, user management, integration connections
- **Main Application** (`test_main.py`): Health checks, CORS, startup/shutdown events

### Integration Tests
- **Full Workflows** (`test_integration.py`): Complete user journeys, cross-service interactions
- **Error Handling**: Consistent error responses across endpoints
- **Concurrent Operations**: Thread safety and resource management

## Test Features

### Mocking Strategy
- **Database Service**: Mocked with realistic return values and error scenarios
- **External APIs**: GitHub, Vercel, and OAuth provider responses mocked
- **Docker Operations**: Container management operations mocked
- **File System**: File operations mocked for security and isolation
- **Agent Responses**: LangChain agent streaming responses mocked

### Test Fixtures
- `client`: FastAPI test client for HTTP requests
- `mock_db_service`: Comprehensive database service mock
- `mock_agent`: ReAct agent mock with streaming responses
- `sample_project`: Sample project data for testing
- `sample_user`: Sample user data for authentication tests
- `sample_message`: Sample conversation message
- `sample_token_usage`: Sample token usage record

### Coverage Areas
- ✅ **API Endpoints**: All REST and WebSocket endpoints
- ✅ **Database Operations**: CRUD operations, error handling, retry logic
- ✅ **Authentication**: OAuth flows, token validation, user management
- ✅ **File Management**: Project files, security checks, content retrieval
- ✅ **Token Tracking**: Usage monitoring, statistics, cost calculation
- ✅ **Error Handling**: Consistent error responses, validation
- ✅ **Integration Flows**: End-to-end user workflows

## Running Tests

### Install Dependencies
```bash
uv sync
```

### Run All Tests
```bash

# Or directly with pytest
uv run pytest tests/ -v
```

### Run Specific Test Categories
```bash
# Unit tests only
python run_tests.py unit

# Integration tests only
python run_tests.py integration

# Database tests only
python run_tests.py database

# API endpoint tests
python run_tests.py api

# Authentication tests
python run_tests.py auth

# Main application tests
python run_tests.py main
```

### Run with Coverage
```bash
uv run pytest tests/ --cov=app --cov-report=html --cov-report=term-missing
```

### Run Specific Test Files
```bash
# Test projects API
uv run pytest tests/test_projects.py -v

# Test streaming functionality
uv run pytest tests/test_streaming.py -v

# Test with specific markers
uv run pytest tests/ -m "not slow" -v
```

## Test Configuration

### pytest.ini
- Test discovery patterns
- Output formatting
- Warning filters
- Test markers for categorization

### Environment Variables
Tests use environment variable mocking to avoid dependencies on actual configuration:
- `LLM_PROVIDER`: Mocked for model testing
- `MODEL_NAME`: Mocked for configuration testing
- OAuth credentials: Mocked for security

### Async Testing
Uses `pytest-asyncio` for testing async endpoints and database operations:
```python
@pytest.mark.asyncio
async def test_async_function():
    result = await some_async_function()
    assert result is not None
```

## Mock Patterns

### Database Service Mocking
```python
def test_create_project(mock_db_service, sample_project):
    mock_db_service.create_project.return_value = sample_project
    # Test logic here
    mock_db_service.create_project.assert_called_once()
```

### External API Mocking
```python
with patch('httpx.AsyncClient') as mock_client:
    mock_response = Mock()
    mock_response.status_code = 200
    mock_client.return_value.__aenter__.return_value.get = AsyncMock(return_value=mock_response)
    # Test logic here
```

### WebSocket Testing
```python
@pytest.mark.asyncio
async def test_websocket_stream(mock_agent):
    chunks = []
    async for chunk in mock_agent.stream_response("test message", "/path", "container"):
        chunks.append(chunk)
    assert len(chunks) > 0
```

## Test Data

### Sample Data Fixtures
All test fixtures provide realistic data that matches the application's data models:
- Projects with proper UUIDs, timestamps, and relationships
- Users with OAuth provider information
- Messages with proper role assignments and content
- Token usage with realistic token counts and provider information

### Error Scenarios
Tests cover various error conditions:
- Database connection failures
- Invalid input validation
- Authentication failures
- External API errors
- File system permission issues
- Container management failures

## Continuous Integration

### Test Requirements
- Python 3.11+
- All dependencies in `pyproject.toml`
- No external service dependencies (all mocked)
- No persistent state between tests

### Coverage Goals
- Minimum 80% code coverage
- 100% coverage for critical paths (authentication, data persistence)
- All API endpoints tested
- All error conditions covered

### Performance
- Tests should complete in under 30 seconds
- No actual network requests
- No actual file system operations
- No actual database connections

## Debugging Tests

### Verbose Output
```bash
uv run pytest tests/ -v -s
```

### Debug Specific Test
```bash
uv run pytest tests/test_projects.py::TestProjectsAPI::test_create_project_success -v -s
```

### Coverage Report
```bash
uv run pytest tests/ --cov=app --cov-report=html
open htmlcov/index.html
```

### Test Timing
```bash
uv run pytest tests/ --durations=10
```

## Contributing

When adding new features:
1. Add corresponding unit tests
2. Update integration tests if needed
3. Ensure all tests pass
4. Maintain coverage above 80%
5. Add appropriate test fixtures
6. Mock external dependencies
7. Test error conditions

### Test Naming Convention
- Test files: `test_<module_name>.py`
- Test classes: `Test<FeatureName>API`
- Test methods: `test_<action>_<condition>`

Example:
```python
class TestProjectsAPI:
    def test_create_project_success(self):
        # Test successful project creation
        
    def test_create_project_invalid_data(self):
        # Test project creation with invalid data
        
    def test_get_project_not_found(self):
        # Test retrieval of non-existent project
```