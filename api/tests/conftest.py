"""
Test configuration and fixtures for the API tests.
"""
import pytest
import asyncio
from unittest.mock import Mock, AsyncMock, patch
from fastapi.testclient import TestClient
from datetime import datetime
import uuid

# Import the main app
from main import app
from app.database.models import (
    Project, ProjectCreate, ConversationMessage, ConversationMessageCreate,
    TokenUsage, TokenUsageCreate, User, UserCreate
)

@pytest.fixture
def client():
    """Create a test client for the FastAPI app."""
    return TestClient(app)

@pytest.fixture
def mock_db_service():
    """Mock database service with common methods."""
    mock_service = Mock()
    
    # Mock project methods
    mock_service.create_project = Mock()
    mock_service.get_project_by_id = Mock()
    mock_service.get_project_by_name = Mock()
    mock_service.get_all_projects = Mock()
    mock_service.update_project = Mock()
    mock_service.delete_project = Mock()
    
    # Mock conversation methods
    mock_service.create_conversation_message = Mock()
    mock_service.get_project_messages = Mock()
    mock_service.get_conversation_messages = Mock()
    mock_service.get_chat_summary = Mock()
    
    # Mock token usage methods
    mock_service.create_token_usage = Mock()
    mock_service.get_session_token_usage = Mock()
    mock_service.get_project_token_usage = Mock()
    mock_service.get_global_token_stats = Mock()
    
    # Mock user methods
    mock_service.create_user = AsyncMock()
    mock_service.get_user_by_id = AsyncMock()
    mock_service.get_user_by_email = AsyncMock()
    mock_service.update_user_github = AsyncMock()
    mock_service.update_user_vercel = AsyncMock()
    
    # Mock utility methods
    mock_service.generate_fancy_project_name = Mock()
    
    return mock_service

@pytest.fixture
def mock_agent():
    """Mock ReAct agent for testing streaming responses."""
    mock_agent = Mock()
    
    async def mock_stream_response(message, project_path, container_name):
        """Mock streaming response generator."""
        chunks = [
            {"type": "content", "content": "I'll help you with that. "},
            {"type": "content", "content": "Let me analyze your request. "},
            {"type": "content", "content": "Here's what I'll do..."},
        ]
        for chunk in chunks:
            yield chunk
    
    mock_agent.stream_response = mock_stream_response
    return mock_agent

@pytest.fixture
def sample_project():
    """Sample project data for testing."""
    return Project(
        id="test-project-id",
        name="TestProject",
        template="reactjs",
        docker_container="test-container",
        port=3000,
        status="created",
        created_at=datetime.now(),
        updated_at=datetime.now()
    )

@pytest.fixture
def sample_project_create():
    """Sample project creation data."""
    return ProjectCreate(
        name="TestProject",
        template="reactjs",
        docker_container="test-container",
        port=3000,
        message="Create a test project"
    )

@pytest.fixture
def sample_user():
    """Sample user data for testing."""
    return User(
        id="test-user-id",
        email="test@example.com",
        name="Test User",
        avatar_url="https://example.com/avatar.jpg",
        google_id="google-123",
        created_at=datetime.now(),
        updated_at=datetime.now()
    )

@pytest.fixture
def sample_message():
    """Sample conversation message for testing."""
    return ConversationMessage(
        id="test-message-id",
        project_id="test-project-id",
        role="user",
        content="Hello, world!",
        message_type="chat",
        model="gpt-4",
        provider="openai",
        created_at=datetime.now(),
        updated_at=datetime.now()
    )

@pytest.fixture
def sample_token_usage():
    """Sample token usage data for testing."""
    return TokenUsage(
        id="test-usage-id",
        session_id="test-session-id",
        project_id="test-project-id",
        model="gpt-4",
        provider="openai",
        input_tokens=100,
        output_tokens=50,
        total_tokens=150,
        request_type="chat",
        created_at=datetime.now()
    )

@pytest.fixture
def mock_docker_utils():
    """Mock Docker utility functions."""
    with patch('app.utils.docker_route.deploy_app') as mock_deploy, \
         patch('app.utils.docker_route.ensure_container_running') as mock_ensure, \
         patch('app.utils.docker_route.get_container_status_for_project') as mock_status, \
         patch('app.utils.docker_route.delete_project_and_cleanup') as mock_cleanup:
        
        mock_deploy.return_value = {
            "container_name": "test-container",
            "project_path": "/tmp/test-project"
        }
        
        mock_ensure.return_value = {"success": True}
        
        mock_status.return_value = {
            "running": True,
            "needs_start": False,
            "status": "running"
        }
        
        mock_cleanup.return_value = {
            "container_removed": True,
            "image_removed": True,
            "files_removed": True,
            "errors": []
        }
        
        yield {
            "deploy_app": mock_deploy,
            "ensure_container_running": mock_ensure,
            "get_container_status_for_project": mock_status,
            "delete_project_and_cleanup": mock_cleanup
        }

@pytest.fixture
def mock_external_apis():
    """Mock external API calls (GitHub, Vercel, etc.)."""
    with patch('httpx.AsyncClient') as mock_client:
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = {"success": True}
        
        mock_client.return_value.__aenter__.return_value.get = AsyncMock(return_value=mock_response)
        mock_client.return_value.__aenter__.return_value.post = AsyncMock(return_value=mock_response)
        mock_client.return_value.__aenter__.return_value.delete = AsyncMock(return_value=mock_response)
        
        yield mock_client

@pytest.fixture(scope="session")
def event_loop():
    """Create an instance of the default event loop for the test session."""
    loop = asyncio.get_event_loop_policy().new_event_loop()
    yield loop
    loop.close()