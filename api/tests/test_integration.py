"""
Integration tests for API endpoints with database and external services.
"""
import pytest
from unittest.mock import patch, Mock, AsyncMock
import json
import asyncio


class TestAPIIntegration:
    """Integration test cases for API workflows."""
    
    def test_full_project_lifecycle(self, client, mock_db_service, mock_docker_utils, sample_project):
        """Test complete project lifecycle: create, retrieve, update, delete."""
        # Arrange
        project_data = {
            "name": "IntegrationTestProject",
            "template": "reactjs",
            "message": "Create a test project for integration testing"
        }
        
        # Mock database responses for project lifecycle
        mock_db_service.generate_fancy_project_name.return_value = "IntegrationTestProject"
        mock_db_service.create_project.return_value = sample_project
        mock_db_service.update_project.return_value = sample_project
        mock_db_service.get_project_by_id.return_value = sample_project
        mock_db_service.get_all_projects.return_value = [sample_project]
        mock_db_service.delete_project.return_value = True
        mock_db_service.create_conversation_message.return_value = Mock()
        
        with patch('app.api.projects.db_service', mock_db_service):
            # Act & Assert - Create project
            create_response = client.post("/api/v1/projects/", json=project_data)
            assert create_response.status_code == 201
            created_project = create_response.json()
            project_id = created_project["id"]
            
            # Act & Assert - Get all projects
            list_response = client.get("/api/v1/projects")
            assert list_response.status_code == 200
            projects_list = list_response.json()
            assert len(projects_list["projects"]) == 1
            
            # Act & Assert - Get specific project
            get_response = client.get(f"/api/v1/projects/{project_id}")
            assert get_response.status_code == 200
            retrieved_project = get_response.json()
            assert retrieved_project["id"] == project_id
            
            # Act & Assert - Delete project
            delete_response = client.delete(f"/api/v1/projects/{project_id}")
            assert delete_response.status_code == 200
            delete_result = delete_response.json()
            assert delete_result["message"] == "Project deleted successfully"
    
    def test_chat_session_with_token_tracking(self, client, mock_db_service, mock_docker_utils, sample_project, sample_token_usage):
        """Test chat session creation with token usage tracking."""
        # Arrange
        chat_request = {
            "message": "Create a React component with TypeScript"
        }
        
        mock_db_service.generate_fancy_project_name.return_value = "ReactComponentProject"
        mock_db_service.create_project.return_value = sample_project
        mock_db_service.create_conversation_message.return_value = Mock()
        mock_db_service.get_session_token_usage.return_value = [sample_token_usage]
        
        with patch('app.api.streaming.db_service', mock_db_service):
            # Act - Create chat session
            session_response = client.post("/api/v1/chat/create-session", json=chat_request)
            assert session_response.status_code == 200
            session_data = session_response.json()
            session_id = session_data["session_id"]
            
            # Act - Check token usage
            with patch('app.api.tokens.db_service', mock_db_service):
                usage_response = client.get(f"/api/v1/tokens/usage/{session_id}")
                assert usage_response.status_code == 200
                usage_data = usage_response.json()
                assert usage_data["session_id"] == session_id
                assert usage_data["total_tokens"] == 150
    
    def test_project_files_and_content_workflow(self, client, mock_db_service, sample_project):
        """Test project file listing and content retrieval workflow."""
        # Arrange
        project_name = "TestProject"
        mock_db_service.get_project_by_name.return_value = sample_project
        
        # Mock file system operations
        mock_files = ['src', 'package.json', 'README.md']
        file_content = "console.log('Hello, world!');"
        
        with patch('app.api.projects.db_service', mock_db_service), \
             patch('os.path.isdir', return_value=True), \
             patch('os.listdir', return_value=mock_files), \
             patch('os.path.join', side_effect=lambda *args: '/'.join(args)), \
             patch('os.path.getsize', return_value=1024), \
             patch('os.path.exists', return_value=True), \
             patch('os.path.isfile', return_value=True), \
             patch('os.path.abspath', side_effect=lambda x: f"/safe/projects/{x}" if not x.startswith('/') else x), \
             patch('builtins.open', mock_open(read_data=file_content)):
            
            # Act - Get project files
            files_response = client.get(f"/api/v1/projects/{project_name}/files")
            assert files_response.status_code == 200
            files_data = files_response.json()
            assert "files" in files_data
            assert len(files_data["files"]) == 3
            
            # Act - Get file content
            content_response = client.get(f"/api/v1/projects/{project_name}/files/src/index.js")
            assert content_response.status_code == 200
            content_data = content_response.json()
            assert content_data["content"] == file_content
            assert content_data["file_path"] == "src/index.js"
    
    def test_models_and_tokens_integration(self, client, mock_db_service):
        """Test models and tokens API integration."""
        # Arrange
        mock_stats = {
            "total_tokens": 5000,
            "total_input_tokens": 3000,
            "total_output_tokens": 2000,
            "total_sessions": 10,
            "models_used": ["gpt-4", "claude-3.5-sonnet"],
            "providers_used": ["openai", "anthropic"],
            "last_updated": "2024-01-15T10:30:00"
        }
        mock_db_service.get_global_token_stats.return_value = mock_stats
        
        with patch('app.api.tokens.db_service', mock_db_service), \
             patch.dict('os.environ', {'LLM_PROVIDER': 'openrouter', 'MODEL_NAME': 'anthropic/claude-3.5-sonnet'}):
            
            # Act - Get available models
            models_response = client.get("/api/v1/models/all")
            assert models_response.status_code == 200
            models_data = models_response.json()
            assert models_data["provider"] == "openrouter"
            assert "anthropic/claude-3.5-sonnet" in models_data["models"]
            
            # Act - Get global token stats
            stats_response = client.get("/api/v1/tokens/stats")
            assert stats_response.status_code == 200
            stats_data = stats_response.json()
            assert stats_data["total_tokens"] == 5000
            assert "gpt-4" in stats_data["models_used"]
    
    def test_error_handling_across_endpoints(self, client, mock_db_service):
        """Test error handling consistency across different endpoints."""
        # Arrange - Mock database errors
        mock_db_service.get_project_by_id.return_value = None
        mock_db_service.get_session_token_usage.side_effect = Exception("Database error")
        mock_db_service.get_global_token_stats.side_effect = Exception("Connection failed")
        
        with patch('app.api.projects.db_service', mock_db_service), \
             patch('app.api.tokens.db_service', mock_db_service):
            
            # Act & Assert - Project not found
            project_response = client.get("/api/v1/projects/nonexistent-id")
            assert project_response.status_code == 404
            assert "Project not found" in project_response.json()["detail"]
            
            # Act & Assert - Token usage database error
            usage_response = client.get("/api/v1/tokens/usage/test-session")
            assert usage_response.status_code == 500
            assert "Database error" in usage_response.json()["detail"]
            
            # Act & Assert - Global stats database error
            stats_response = client.get("/api/v1/tokens/stats")
            assert stats_response.status_code == 500
            assert "Connection failed" in stats_response.json()["detail"]
    
    def test_concurrent_project_operations(self, client, mock_db_service, mock_docker_utils):
        """Test concurrent project operations don't interfere with each other."""
        # Arrange
        project_data_1 = {
            "name": "Project1",
            "template": "reactjs",
            "message": "Create first project"
        }
        project_data_2 = {
            "name": "Project2",
            "template": "nodejs",
            "message": "Create second project"
        }
        
        # Mock different projects
        project_1 = Mock(id="project-1", name="Project1", template="reactjs")
        project_2 = Mock(id="project-2", name="Project2", template="nodejs")
        
        mock_db_service.generate_fancy_project_name.side_effect = ["Project1", "Project2"]
        mock_db_service.create_project.side_effect = [project_1, project_2]
        mock_db_service.update_project.side_effect = [project_1, project_2]
        mock_db_service.create_conversation_message.return_value = Mock()
        
        with patch('app.api.projects.db_service', mock_db_service):
            # Act - Create projects concurrently (simulated)
            response_1 = client.post("/api/v1/projects/", json=project_data_1)
            response_2 = client.post("/api/v1/projects/", json=project_data_2)
            
            # Assert
            assert response_1.status_code == 201
            assert response_2.status_code == 201
            assert response_1.json()["name"] == "Project1"
            assert response_2.json()["name"] == "Project2"
    
    def test_health_check_integration(self, client):
        """Test health check integration with actual database connection."""
        # This tests the health check with mocked database
        with patch('app.database.connection.db') as mock_db:
            mock_conn = Mock()
            mock_conn.execute.return_value.fetchone.return_value = [1]
            mock_db.get_connection.return_value = mock_conn
            
            # Act
            response = client.get("/health")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert data["status"] == "healthy"
            assert data["database"] == "connected"
    
    def test_api_documentation_endpoints(self, client):
        """Test API documentation endpoints are accessible."""
        # FastAPI automatically generates OpenAPI docs
        
        # Act - Get OpenAPI schema
        openapi_response = client.get("/openapi.json")
        
        # Assert
        assert openapi_response.status_code == 200
        openapi_data = openapi_response.json()
        assert "openapi" in openapi_data
        assert "info" in openapi_data
        assert openapi_data["info"]["title"] == "Code Editing Agent Backend with Authentication & Integrations"
    
    def test_request_validation_integration(self, client):
        """Test request validation across different endpoints."""
        # Test invalid project creation data
        invalid_project_data = {
            "template": "reactjs"
            # Missing required 'name' field
        }
        
        response = client.post("/api/v1/projects/", json=invalid_project_data)
        assert response.status_code == 422  # Validation error
        
        # Test invalid chat session data
        invalid_chat_data = {}  # Missing required 'message' field
        
        response = client.post("/api/v1/chat/create-session", json=invalid_chat_data)
        assert response.status_code == 422  # Validation error
    
    def test_response_format_consistency(self, client, mock_db_service, sample_project):
        """Test response format consistency across endpoints."""
        # Arrange
        mock_db_service.get_all_projects.return_value = [sample_project]
        mock_db_service.get_project_by_id.return_value = sample_project
        
        with patch('app.api.projects.db_service', mock_db_service), \
             patch.dict('os.environ', {'LLM_PROVIDER': 'openai'}):
            
            # Act - Get projects list
            projects_response = client.get("/api/v1/projects")
            projects_data = projects_response.json()
            
            # Act - Get single project
            project_response = client.get("/api/v1/projects/test-id")
            project_data = project_response.json()
            
            # Act - Get models
            models_response = client.get("/api/v1/models/all")
            models_data = models_response.json()
            
            # Assert - All responses should be JSON with consistent structure
            assert isinstance(projects_data, dict)
            assert isinstance(project_data, dict)
            assert isinstance(models_data, dict)
            
            # Assert - Timestamp formats should be consistent
            if "created_at" in project_data and project_data["created_at"]:
                assert "T" in project_data["created_at"]  # ISO format


def mock_open(read_data=""):
    """Helper function to mock file opening."""
    from unittest.mock import mock_open as _mock_open
    return _mock_open(read_data=read_data)