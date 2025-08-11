"""
Unit tests for database service operations.
"""
import pytest
from unittest.mock import Mock, patch, MagicMock
from datetime import datetime
import uuid

from app.database.service import DatabaseService
from app.database.models import (
    Project, ProjectCreate, ConversationMessage, ConversationMessageCreate,
    TokenUsage, TokenUsageCreate, User, UserCreate
)


class TestDatabaseService:
    """Test cases for database service operations."""
    
    @pytest.fixture
    def db_service(self):
        """Create a database service instance with mocked connection."""
        with patch('app.database.service.db') as mock_db:
            mock_conn = Mock()
            mock_db.get_connection.return_value = mock_conn
            mock_db.reconnect.return_value = mock_conn
            
            service = DatabaseService()
            service.conn = mock_conn
            return service
    
    def test_create_project_success(self, db_service):
        """Test successful project creation."""
        # Arrange
        project_data = ProjectCreate(
            name="TestProject",
            template="reactjs",
            docker_container="test-container",
            port=3000,
            message="Create a test project"
        )
        
        mock_result = [
            "test-project-id", "TestProject", "reactjs", "test-container", 
            3000, "created", datetime.now(), datetime.now()
        ]
        
        db_service._fetchone_with_retry = Mock(return_value=mock_result)
        
        # Act
        result = db_service.create_project(project_data)
        
        # Assert
        assert isinstance(result, Project)
        assert result.id == "test-project-id"
        assert result.name == "TestProject"
        assert result.template == "reactjs"
        assert result.docker_container == "test-container"
        assert result.port == 3000
        assert result.status == "created"
        db_service._fetchone_with_retry.assert_called()
        db_service.conn.commit.assert_called()
    
    def test_get_project_by_id_success(self, db_service):
        """Test successful project retrieval by ID."""
        # Arrange
        project_id = "test-project-id"
        mock_result = [
            project_id, "TestProject", "reactjs", "test-container",
            3000, "created", datetime.now(), datetime.now()
        ]
        
        db_service._fetchone_with_retry = Mock(return_value=mock_result)
        
        # Act
        result = db_service.get_project_by_id(project_id)
        
        # Assert
        assert isinstance(result, Project)
        assert result.id == project_id
        assert result.name == "TestProject"
        db_service._fetchone_with_retry.assert_called_once()
    
    def test_get_project_by_id_not_found(self, db_service):
        """Test project retrieval when project doesn't exist."""
        # Arrange
        project_id = "nonexistent-id"
        db_service._fetchone_with_retry = Mock(return_value=None)
        
        # Act
        result = db_service.get_project_by_id(project_id)
        
        # Assert
        assert result is None
        db_service._fetchone_with_retry.assert_called_once()
    
    def test_get_all_projects_success(self, db_service):
        """Test successful retrieval of all projects."""
        # Arrange
        mock_results = [
            ["id1", "Project1", "reactjs", "container1", 3000, "created", datetime.now(), datetime.now()],
            ["id2", "Project2", "nodejs", "container2", 3001, "created", datetime.now(), datetime.now()]
        ]
        
        db_service._fetchall_with_retry = Mock(return_value=mock_results)
        
        # Act
        result = db_service.get_all_projects()
        
        # Assert
        assert len(result) == 2
        assert all(isinstance(project, Project) for project in result)
        assert result[0].name == "Project1"
        assert result[1].name == "Project2"
        db_service._fetchall_with_retry.assert_called_once()
    
    def test_delete_project_success(self, db_service):
        """Test successful project deletion."""
        # Arrange
        project_id = "test-project-id"
        db_service._execute_with_retry = Mock()
        
        # Act
        result = db_service.delete_project(project_id)
        
        # Assert
        assert result is True
        # Should call delete for messages, tokens, and project
        assert db_service._execute_with_retry.call_count == 3
        db_service.conn.commit.assert_called()
    
    def test_delete_project_database_error(self, db_service):
        """Test project deletion with database error."""
        # Arrange
        project_id = "test-project-id"
        db_service._execute_with_retry = Mock(side_effect=Exception("Database error"))
        
        # Act & Assert
        with pytest.raises(Exception, match="Database error"):
            db_service.delete_project(project_id)
    
    def test_create_conversation_message_success(self, db_service):
        """Test successful conversation message creation."""
        # Arrange
        message_data = ConversationMessageCreate(
            project_id="test-project-id",
            role="user",
            content="Hello, world!",
            message_type="chat",
            model="gpt-4",
            provider="openai"
        )
        
        mock_result = [
            "test-message-id", "test-project-id", "user", "Hello, world!",
            "chat", "gpt-4", "openai", None, datetime.now(), datetime.now()
        ]
        
        db_service._fetchone_with_retry = Mock(return_value=mock_result)
        
        # Act
        result = db_service.create_conversation_message(message_data)
        
        # Assert
        assert isinstance(result, ConversationMessage)
        assert result.id == "test-message-id"
        assert result.project_id == "test-project-id"
        assert result.role == "user"
        assert result.content == "Hello, world!"
        db_service._fetchone_with_retry.assert_called_once()
        db_service.conn.commit.assert_called()
    
    def test_get_project_messages_success(self, db_service):
        """Test successful retrieval of project messages."""
        # Arrange
        project_id = "test-project-id"
        mock_results = [
            ["msg1", project_id, "user", "Hello", "chat", "gpt-4", "openai", None, datetime.now(), datetime.now()],
            ["msg2", project_id, "assistant", "Hi there!", "chat", "gpt-4", "openai", None, datetime.now(), datetime.now()]
        ]
        
        db_service._fetchall_with_retry = Mock(return_value=mock_results)
        
        # Act
        result = db_service.get_project_messages(project_id)
        
        # Assert
        assert len(result) == 2
        assert all(isinstance(msg, ConversationMessage) for msg in result)
        assert result[0].role == "user"
        assert result[1].role == "assistant"
        db_service._fetchall_with_retry.assert_called_once()
    
    def test_create_token_usage_success(self, db_service):
        """Test successful token usage creation."""
        # Arrange
        usage_data = TokenUsageCreate(
            session_id="test-session-id",
            project_id="test-project-id",
            model="gpt-4",
            provider="openai",
            input_tokens=100,
            output_tokens=50,
            total_tokens=150
        )
        
        mock_result = [
            "test-usage-id", "test-session-id", "test-project-id", "gpt-4",
            "openai", 100, 50, 150, "chat", datetime.now()
        ]
        
        db_service.conn.execute.return_value.fetchone.return_value = mock_result
        
        # Act
        result = db_service.create_token_usage(usage_data)
        
        # Assert
        assert isinstance(result, TokenUsage)
        assert result.id == "test-usage-id"
        assert result.session_id == "test-session-id"
        assert result.total_tokens == 150
        db_service.conn.execute.assert_called()
        db_service.conn.commit.assert_called()
    
    def test_get_session_token_usage_success(self, db_service):
        """Test successful retrieval of session token usage."""
        # Arrange
        session_id = "test-session-id"
        mock_results = [
            ["usage1", session_id, "project1", "gpt-4", "openai", 100, 50, 150, "chat", datetime.now()],
            ["usage2", session_id, "project1", "gpt-4", "openai", 80, 40, 120, "chat", datetime.now()]
        ]
        
        db_service._fetchall_with_retry = Mock(return_value=mock_results)
        
        # Act
        result = db_service.get_session_token_usage(session_id)
        
        # Assert
        assert len(result) == 2
        assert all(isinstance(usage, TokenUsage) for usage in result)
        assert result[0].total_tokens == 150
        assert result[1].total_tokens == 120
        db_service._fetchall_with_retry.assert_called_once()
    
    def test_get_global_token_stats_success(self, db_service):
        """Test successful retrieval of global token statistics."""
        # Arrange
        mock_totals = [10000, 6000, 4000, 25]  # total_tokens, input, output, sessions
        mock_models = [["gpt-4"], ["claude-3.5-sonnet"]]
        mock_providers = [["openai"], ["anthropic"]]
        mock_last_updated = [datetime.now()]
        
        db_service._fetchone_with_retry = Mock(return_value=mock_totals)
        db_service._fetchall_with_retry = Mock(side_effect=[mock_models, mock_providers])
        
        # Mock the last_updated query separately
        def mock_fetchone_side_effect(query, params=None):
            if "MAX(created_at)" in query:
                return mock_last_updated
            return mock_totals
        
        db_service._fetchone_with_retry.side_effect = mock_fetchone_side_effect
        
        # Act
        result = db_service.get_global_token_stats()
        
        # Assert
        assert result["total_tokens"] == 10000
        assert result["total_input_tokens"] == 6000
        assert result["total_output_tokens"] == 4000
        assert result["total_sessions"] == 25
        assert "gpt-4" in result["models_used"]
        assert "openai" in result["providers_used"]
        assert result["last_updated"] is not None
    
    def test_get_global_token_stats_error_handling(self, db_service):
        """Test global token stats with database error."""
        # Arrange
        db_service._fetchone_with_retry = Mock(side_effect=Exception("Database error"))
        
        # Act
        result = db_service.get_global_token_stats()
        
        # Assert
        assert result["total_tokens"] == 0
        assert result["models_used"] == []
        assert result["last_updated"] is None
    
    def test_generate_fancy_project_name(self, db_service):
        """Test fancy project name generation."""
        # Arrange
        query = "Create a React application with TypeScript"
        
        with patch('random.choice') as mock_choice, \
             patch('random.randint', return_value=42):
            mock_choice.side_effect = ["Stellar", "Hub"]
            
            # Act
            result = db_service.generate_fancy_project_name(query)
            
            # Assert
            assert "Stellar" in result
            assert "React" in result or "Application" in result
            assert "Hub" in result
            assert "42" in result
    
    def test_get_chat_summary_with_messages(self, db_service):
        """Test chat summary generation with existing messages."""
        # Arrange
        project_id = "test-project-id"
        mock_messages = [
            Mock(role="user", content="Hello"),
            Mock(role="assistant", content="Hi there!"),
            Mock(role="user", content="How are you?"),
            Mock(role="assistant", content="I'm doing well, thank you!")
        ]
        
        db_service.get_project_messages = Mock(return_value=mock_messages)
        
        # Act
        result = db_service.get_chat_summary(project_id)
        
        # Assert
        assert "User has made 2 requests" in result
        assert "Assistant has provided 2 responses" in result
        assert "Recent conversation context:" in result
        assert "Hello" in result
    
    def test_get_chat_summary_empty(self, db_service):
        """Test chat summary generation with no messages."""
        # Arrange
        project_id = "test-project-id"
        db_service.get_project_messages = Mock(return_value=[])
        
        # Act
        result = db_service.get_chat_summary(project_id)
        
        # Assert
        assert result == ""
    
    def test_execute_with_retry_success(self, db_service):
        """Test successful query execution with retry logic."""
        # Arrange
        query = "SELECT * FROM projects"
        params = ["test-id"]
        mock_result = Mock()
        db_service.conn.execute.return_value = mock_result
        
        # Act
        result = db_service._execute_with_retry(query, params)
        
        # Assert
        assert result == mock_result
        db_service.conn.execute.assert_called_once_with(query, params)
    
    def test_execute_with_retry_database_invalidation(self, db_service):
        """Test retry logic on database invalidation."""
        # Arrange
        query = "SELECT * FROM projects"
        
        # Create a proper DuckDB FatalException mock
        import duckdb
        mock_error = duckdb.FatalException("database has been invalidated")
        
        db_service.conn.execute.side_effect = [mock_error, Mock()]
        
        with patch('app.database.service.db') as mock_db:
            mock_db.reconnect.return_value = Mock()
            
            # Act
            result = db_service._execute_with_retry(query)
            
            # Assert
            assert db_service.conn.execute.call_count == 2
            mock_db.reconnect.assert_called_once()
    
    def test_execute_with_retry_max_retries_exceeded(self, db_service):
        """Test retry logic when max retries are exceeded."""
        # Arrange
        query = "SELECT * FROM projects"
        db_service.conn.execute.side_effect = Exception("Persistent error")
        
        with patch('app.database.service.db') as mock_db:
            mock_db.reconnect.return_value = Mock()
            
            # Act & Assert
            with pytest.raises(Exception, match="Persistent error"):
                db_service._execute_with_retry(query, max_retries=2)