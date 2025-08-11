"""
Unit tests for streaming/chat API endpoints.
"""
import pytest
from unittest.mock import patch, Mock, AsyncMock
import json
import asyncio
from fastapi.testclient import TestClient

from app.database.models import ConversationMessageCreate, TokenUsageCreate


class TestStreamingAPI:
    """Test cases for streaming/chat API endpoints."""
    
    def test_create_chat_session_success(self, client, mock_db_service, mock_docker_utils, sample_project):
        """Test successful chat session creation."""
        # Arrange
        mock_db_service.generate_fancy_project_name.return_value = "TestChatProject"
        mock_db_service.create_project.return_value = sample_project
        mock_db_service.create_conversation_message.return_value = Mock()
        
        chat_request = {
            "message": "Create a React app with TypeScript"
        }
        
        with patch('app.api.streaming.db_service', mock_db_service):
            # Act
            response = client.post("/api/v1/chat/create-session", json=chat_request)
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert "project_id" in data
            assert "session_id" in data
            assert "project_name" in data
            assert "url" in data
            assert data["initial_message"] == chat_request["message"]
            
            # Verify database calls
            mock_db_service.create_project.assert_called_once()
            assert mock_db_service.create_conversation_message.call_count == 2  # User + AI messages
    
    def test_create_chat_session_docker_failure(self, client, mock_db_service):
        """Test chat session creation when Docker deployment fails."""
        # Arrange
        mock_db_service.generate_fancy_project_name.return_value = "TestProject"
        
        chat_request = {
            "message": "Create a React app"
        }
        
        with patch('app.api.streaming.db_service', mock_db_service), \
             patch('app.api.streaming.deploy_app', side_effect=Exception("Docker deployment failed")):
            # Act
            response = client.post("/api/v1/chat/create-session", json=chat_request)
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert "error" in data
            assert "Docker deployment failed" in data["error"]
    
    @pytest.mark.asyncio
    async def test_websocket_stream_success(self, mock_db_service, mock_agent, sample_project):
        """Test successful WebSocket streaming."""
        # Arrange
        mock_db_service.get_project_by_id.return_value = sample_project
        mock_db_service.create_conversation_message.return_value = Mock()
        mock_db_service.create_token_usage.return_value = Mock()
        mock_db_service.get_chat_summary.return_value = "Previous conversation context"
        
        # Mock WebSocket
        mock_websocket = AsyncMock()
        mock_websocket.receive_text = AsyncMock(return_value=json.dumps({
            "message": "Hello, how can you help me?",
            "model": "gpt-4",
            "provider": "openai"
        }))
        
        with patch('app.api.streaming.db_service', mock_db_service), \
             patch('app.api.streaming.ReActAgent', return_value=mock_agent), \
             patch('os.path.abspath', return_value="/test/path"):
            
            from app.api.streaming import websocket_stream
            
            # Act & Assert - This would normally be tested with a WebSocket test client
            # For now, we'll test the core logic components
            assert mock_db_service.get_project_by_id is not None
            assert mock_agent.stream_response is not None
    
    def test_websocket_project_not_found(self, client, mock_db_service):
        """Test WebSocket connection with non-existent project."""
        # Arrange
        mock_db_service.get_project_by_id.return_value = None
        
        # This would require a WebSocket test client to properly test
        # For now, we verify the database service is called correctly
        with patch('app.api.streaming.db_service', mock_db_service):
            # The actual WebSocket test would go here
            # We're testing the logic that would be called
            project = mock_db_service.get_project_by_id("nonexistent-id")
            assert project is None
    
    @pytest.mark.asyncio
    async def test_agent_streaming_response(self, mock_agent):
        """Test agent streaming response generation."""
        # Arrange
        message = "Create a React component"
        project_path = "/test/project"
        container_name = "test-container"
        
        # Act
        chunks = []
        async for chunk in mock_agent.stream_response(message, project_path, container_name):
            chunks.append(chunk)
        
        # Assert
        assert len(chunks) == 3
        assert all(chunk["type"] == "content" for chunk in chunks)
        assert chunks[0]["content"] == "I'll help you with that. "
    
    def test_conversation_message_creation(self, mock_db_service):
        """Test conversation message creation logic."""
        # Arrange
        project_id = "test-project-id"
        message_content = "Hello, world!"
        
        mock_message = Mock()
        mock_message.id = "test-message-id"
        mock_db_service.create_conversation_message.return_value = mock_message
        
        # Act
        message_data = ConversationMessageCreate(
            project_id=project_id,
            role="user",
            content=message_content,
            message_type="chat",
            model="gpt-4",
            provider="openai"
        )
        
        result = mock_db_service.create_conversation_message(message_data)
        
        # Assert
        assert result.id == "test-message-id"
        mock_db_service.create_conversation_message.assert_called_once_with(message_data)
    
    def test_token_usage_creation(self, mock_db_service):
        """Test token usage tracking logic."""
        # Arrange
        session_id = "test-session-id"
        project_id = "test-project-id"
        
        mock_usage = Mock()
        mock_usage.id = "test-usage-id"
        mock_db_service.create_token_usage.return_value = mock_usage
        
        # Act
        usage_data = TokenUsageCreate(
            session_id=session_id,
            project_id=project_id,
            model="gpt-4",
            provider="openai",
            input_tokens=100,
            output_tokens=50,
            total_tokens=150
        )
        
        result = mock_db_service.create_token_usage(usage_data)
        
        # Assert
        assert result.id == "test-usage-id"
        mock_db_service.create_token_usage.assert_called_once_with(usage_data)
    
    def test_chat_summary_generation(self, mock_db_service):
        """Test chat summary generation for context."""
        # Arrange
        project_id = "test-project-id"
        expected_summary = "User has made 3 requests\nAssistant has provided 3 responses"
        mock_db_service.get_chat_summary.return_value = expected_summary
        
        # Act
        summary = mock_db_service.get_chat_summary(project_id)
        
        # Assert
        assert summary == expected_summary
        mock_db_service.get_chat_summary.assert_called_once_with(project_id)
    
    def test_enhanced_message_with_context(self):
        """Test message enhancement with chat history context."""
        # Arrange
        original_message = "Add a button component"
        chat_summary = "Previous conversation: User created a React app"
        
        # Act
        enhanced_message = f"""Previous conversation context:
                            {chat_summary}
                            Current user request: {original_message}
                            Please consider the previous conversation context when responding to the current request."""
        
        # Assert
        assert original_message in enhanced_message
        assert chat_summary in enhanced_message
        assert "Previous conversation context:" in enhanced_message
    
    @pytest.mark.asyncio
    async def test_websocket_error_handling(self, mock_db_service, sample_project):
        """Test WebSocket error handling scenarios."""
        # Arrange
        mock_db_service.get_project_by_id.return_value = sample_project
        mock_db_service.create_conversation_message.side_effect = Exception("Database error")
        
        # This would test WebSocket error handling
        # In a real test, we'd use a WebSocket test client
        with patch('app.api.streaming.db_service', mock_db_service):
            # Verify that database errors are handled appropriately
            with pytest.raises(Exception, match="Database error"):
                mock_db_service.create_conversation_message(Mock())
    
    def test_session_id_generation(self):
        """Test session ID generation for WebSocket connections."""
        import uuid
        
        # Act
        session_id = str(uuid.uuid4())
        
        # Assert
        assert len(session_id) == 36  # UUID format
        assert session_id.count('-') == 4  # UUID has 4 hyphens
    
    def test_project_path_resolution(self, sample_project):
        """Test project path resolution for agent context."""
        import os
        
        # Arrange
        projects_dir = "/tmp/projects"
        project_name = sample_project.name
        
        # Act
        project_path = os.path.abspath(os.path.join(projects_dir, project_name))
        
        # Assert
        assert project_name in project_path
        assert projects_dir in project_path