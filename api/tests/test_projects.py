"""
Unit tests for project-related API endpoints.
"""
import pytest
from unittest.mock import patch, Mock
from fastapi import HTTPException
import json

from app.database.models import Project, ProjectCreate


class TestProjectsAPI:
    """Test cases for projects API endpoints."""
    
    def test_get_projects_success(self, client, mock_db_service, sample_project):
        """Test successful retrieval of all projects."""
        # Arrange
        mock_db_service.get_all_projects.return_value = [sample_project]
        
        with patch('app.api.projects.db_service', mock_db_service):
            # Act
            response = client.get("/api/v1/projects")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert "projects" in data
            assert len(data["projects"]) == 1
            assert data["projects"][0]["id"] == "test-project-id"
            assert data["projects"][0]["name"] == "TestProject"
            mock_db_service.get_all_projects.assert_called_once()
    
    def test_get_projects_empty(self, client, mock_db_service):
        """Test retrieval when no projects exist."""
        # Arrange
        mock_db_service.get_all_projects.return_value = []
        
        with patch('app.api.projects.db_service', mock_db_service):
            # Act
            response = client.get("/api/v1/projects")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert data["projects"] == []
    
    def test_create_project_success(self, client, mock_db_service, mock_docker_utils, sample_project):
        """Test successful project creation."""
        # Arrange
        mock_db_service.generate_fancy_project_name.return_value = "TestProject"
        mock_db_service.create_project.return_value = sample_project
        mock_db_service.update_project.return_value = sample_project
        mock_db_service.create_conversation_message.return_value = Mock()
        
        project_data = {
            "name": "TestProject",
            "template": "reactjs",
            "message": "Create a test project"
        }
        
        with patch('app.api.projects.db_service', mock_db_service):
            # Act
            response = client.post("/api/v1/projects/", json=project_data)
            
            # Assert
            assert response.status_code == 201
            data = response.json()
            assert data["message"] == "Project created successfully"
            assert data["name"] == "TestProject"
            assert "docker_container" in data
            mock_db_service.create_project.assert_called_once()
            mock_db_service.create_conversation_message.assert_called_once()
    
    def test_create_project_docker_failure(self, client, mock_db_service):
        """Test project creation when Docker deployment fails."""
        # Arrange
        mock_db_service.generate_fancy_project_name.return_value = "TestProject"
        mock_db_service.create_project.return_value = Mock(id="test-id")
        
        project_data = {
            "name": "TestProject",
            "template": "reactjs",
            "message": "Create a test project"
        }
        
        with patch('app.api.projects.db_service', mock_db_service), \
             patch('app.api.projects.deploy_app', side_effect=Exception("Docker error")):
            # Act
            response = client.post("/api/v1/projects/", json=project_data)
            
            # Assert
            assert response.status_code == 200  # Returns error in response body
            data = response.json()
            assert "error" in data
            assert "Docker error" in data["error"]
    
    def test_get_project_by_id_success(self, client, mock_db_service, mock_docker_utils, sample_project):
        """Test successful retrieval of project by ID."""
        # Arrange
        mock_db_service.get_project_by_id.return_value = sample_project
        
        with patch('app.api.projects.db_service', mock_db_service):
            # Act
            response = client.get("/api/v1/projects/test-project-id")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert data["id"] == "test-project-id"
            assert data["name"] == "TestProject"
            assert "container_info" in data
            mock_db_service.get_project_by_id.assert_called_once_with("test-project-id")
    
    def test_get_project_by_id_not_found(self, client, mock_db_service):
        """Test retrieval of non-existent project."""
        # Arrange
        mock_db_service.get_project_by_id.return_value = None
        
        with patch('app.api.projects.db_service', mock_db_service):
            # Act
            response = client.get("/api/v1/projects/nonexistent-id")
            
            # Assert
            assert response.status_code == 404
            data = response.json()
            assert data["detail"] == "Project not found"
    
    def test_delete_project_success(self, client, mock_db_service, mock_docker_utils, sample_project):
        """Test successful project deletion."""
        # Arrange
        mock_db_service.get_project_by_id.return_value = sample_project
        mock_db_service.delete_project.return_value = True
        
        with patch('app.api.projects.db_service', mock_db_service):
            # Act
            response = client.delete("/api/v1/projects/test-project-id")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert data["message"] == "Project deleted successfully"
            assert data["project_id"] == "test-project-id"
            assert "cleanup_result" in data
            mock_db_service.delete_project.assert_called_once_with("test-project-id")
    
    def test_delete_project_not_found(self, client, mock_db_service):
        """Test deletion of non-existent project."""
        # Arrange
        mock_db_service.get_project_by_id.return_value = None
        
        with patch('app.api.projects.db_service', mock_db_service):
            # Act
            response = client.delete("/api/v1/projects/nonexistent-id")
            
            # Assert
            assert response.status_code == 404
            data = response.json()
            assert data["detail"] == "Project not found"
    
    def test_get_project_conversations_success(self, client, mock_db_service, sample_project, sample_message):
        """Test successful retrieval of project conversations."""
        # Arrange
        mock_db_service.get_project_by_id.return_value = sample_project
        mock_db_service.get_project_messages.return_value = [sample_message]
        
        with patch('app.api.projects.db_service', mock_db_service):
            # Act
            response = client.get("/api/v1/projects/test-project-id/conversations")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert data["project_id"] == "test-project-id"
            assert len(data["messages"]) == 1
            assert data["messages"][0]["content"] == "Hello, world!"
    
    def test_get_project_files_success(self, client, mock_db_service, sample_project):
        """Test successful retrieval of project files."""
        # Arrange
        mock_db_service.get_project_by_name.return_value = sample_project
        
        with patch('app.api.projects.db_service', mock_db_service), \
             patch('os.path.isdir', return_value=True), \
             patch('os.listdir', return_value=['src', 'package.json']), \
             patch('os.path.join', side_effect=lambda *args: '/'.join(args)), \
             patch('os.path.getsize', return_value=1024):
            # Act
            response = client.get("/api/v1/projects/TestProject/files")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert "files" in data
            assert len(data["files"]) == 2
    
    def test_get_project_files_not_found(self, client, mock_db_service):
        """Test retrieval of files for non-existent project."""
        # Arrange
        mock_db_service.get_project_by_name.return_value = None
        
        with patch('app.api.projects.db_service', mock_db_service):
            # Act
            response = client.get("/api/v1/projects/NonExistentProject/files")
            
            # Assert
            assert response.status_code == 404
            data = response.json()
            assert data["detail"] == "Project not found"
    
    def test_get_file_content_success(self, client, mock_db_service, sample_project):
        """Test successful retrieval of file content."""
        # Arrange
        mock_db_service.get_project_by_name.return_value = sample_project
        file_content = "console.log('Hello, world!');"
        
        with patch('app.api.projects.db_service', mock_db_service), \
             patch('os.path.exists', return_value=True), \
             patch('os.path.isfile', return_value=True), \
             patch('os.path.abspath', side_effect=lambda x: x), \
             patch('builtins.open', mock_open(read_data=file_content)):
            # Act
            response = client.get("/api/v1/projects/TestProject/files/src/index.js")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert data["content"] == file_content
            assert data["file_path"] == "src/index.js"
    
    def test_get_file_content_security_violation(self, client, mock_db_service, sample_project):
        """Test file access security check."""
        # Arrange
        mock_db_service.get_project_by_name.return_value = sample_project
        
        with patch('app.api.projects.db_service', mock_db_service), \
             patch('os.path.exists', return_value=True), \
             patch('os.path.isfile', return_value=True), \
             patch('os.path.abspath') as mock_abspath, \
             patch('os.path.join', side_effect=lambda *args: '/'.join(args)):
            
            # Mock abspath to return different paths for security check
            def mock_abspath_side_effect(path):
                if "malicious" in path:
                    return "/malicious/path/file.txt"
                elif "TestProject" in path:
                    return "/safe/projects/TestProject"
                else:
                    return "/safe/projects/TestProject/file.txt"
            
            mock_abspath.side_effect = mock_abspath_side_effect
            
            # Act
            response = client.get("/api/v1/projects/TestProject/files/../../../malicious/file.txt")
            
            # Assert
            assert response.status_code == 403
            data = response.json()
            assert data["detail"] == "Access denied"


def mock_open(read_data=""):
    """Helper function to mock file opening."""
    from unittest.mock import mock_open as _mock_open
    return _mock_open(read_data=read_data)