"""
Unit tests for models and tokens API endpoints.
"""
import pytest
from unittest.mock import patch, Mock
from datetime import datetime


class TestModelsAPI:
    """Test cases for models API endpoints."""
    
    def test_get_all_models_openrouter(self, client):
        """Test getting all models with OpenRouter provider."""
        # Arrange
        with patch.dict('os.environ', {'LLM_PROVIDER': 'openrouter', 'MODEL_NAME': 'anthropic/claude-3.5-sonnet'}):
            # Act
            response = client.get("/api/v1/models/all")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert data["provider"] == "openrouter"
            assert "anthropic/claude-3.5-sonnet" in data["models"]
            assert "openai/gpt-4o" in data["models"]
            assert data["current_model"] == "anthropic/claude-3.5-sonnet"
    
    def test_get_all_models_openai(self, client):
        """Test getting all models with OpenAI provider."""
        # Arrange
        with patch.dict('os.environ', {'LLM_PROVIDER': 'openai', 'MODEL_NAME': 'gpt-4'}):
            # Act
            response = client.get("/api/v1/models/all")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert data["provider"] == "openai"
            assert "gpt-4" in data["models"]
            assert "gpt-3.5-turbo" in data["models"]
            assert data["current_model"] == "gpt-4"
    
    def test_get_all_models_anthropic(self, client):
        """Test getting all models with Anthropic provider."""
        # Arrange
        with patch.dict('os.environ', {'LLM_PROVIDER': 'anthropic', 'MODEL_NAME': 'claude-3-5-sonnet-20241022'}):
            # Act
            response = client.get("/api/v1/models/all")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert data["provider"] == "anthropic"
            assert "claude-3-5-sonnet-20241022" in data["models"]
            assert "claude-3-haiku-20240307" in data["models"]
    
    def test_get_all_models_google(self, client):
        """Test getting all models with Google provider."""
        # Arrange
        with patch.dict('os.environ', {'LLM_PROVIDER': 'google', 'MODEL_NAME': 'gemini-pro'}):
            # Act
            response = client.get("/api/v1/models/all")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert data["provider"] == "google"
            assert "gemini-pro" in data["models"]
            assert "gemini-pro-vision" in data["models"]
    
    def test_get_all_models_unknown_provider(self, client):
        """Test getting all models with unknown provider defaults to OpenRouter."""
        # Arrange
        with patch.dict('os.environ', {'LLM_PROVIDER': 'unknown-provider'}):
            # Act
            response = client.get("/api/v1/models/all")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert data["provider"] == "unknown-provider"
            # Should default to OpenRouter models
            assert "anthropic/claude-3.5-sonnet" in data["models"]
    
    def test_get_models_legacy_endpoint(self, client):
        """Test legacy models endpoint."""
        # Arrange
        with patch.dict('os.environ', {'LLM_PROVIDER': 'openrouter', 'MODEL_NAME': 'anthropic/claude-3.5-sonnet'}):
            # Act
            response = client.get("/api/v1/models")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert data["provider"] == "openrouter"
            assert data["current_model"] == "anthropic/claude-3.5-sonnet"


class TestTokensAPI:
    """Test cases for tokens API endpoints."""
    
    def test_get_session_usage_success(self, client, mock_db_service, sample_token_usage):
        """Test successful retrieval of session token usage."""
        # Arrange
        session_id = "test-session-id"
        mock_db_service.get_session_token_usage.return_value = [sample_token_usage]
        
        with patch('app.api.tokens.db_service', mock_db_service):
            # Act
            response = client.get(f"/api/v1/tokens/usage/{session_id}")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert data["session_id"] == session_id
            assert data["total_tokens"] == 150
            assert data["input_tokens"] == 100
            assert data["output_tokens"] == 50
            assert len(data["records"]) == 1
            assert data["records"][0]["model"] == "gpt-4"
            mock_db_service.get_session_token_usage.assert_called_once_with(session_id)
    
    def test_get_session_usage_empty(self, client, mock_db_service):
        """Test retrieval of session usage when no records exist."""
        # Arrange
        session_id = "empty-session-id"
        mock_db_service.get_session_token_usage.return_value = []
        
        with patch('app.api.tokens.db_service', mock_db_service):
            # Act
            response = client.get(f"/api/v1/tokens/usage/{session_id}")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert data["session_id"] == session_id
            assert data["total_tokens"] == 0
            assert data["input_tokens"] == 0
            assert data["output_tokens"] == 0
            assert data["records"] == []
    
    def test_get_session_usage_database_error(self, client, mock_db_service):
        """Test session usage retrieval with database error."""
        # Arrange
        session_id = "error-session-id"
        mock_db_service.get_session_token_usage.side_effect = Exception("Database connection failed")
        
        with patch('app.api.tokens.db_service', mock_db_service):
            # Act
            response = client.get(f"/api/v1/tokens/usage/{session_id}")
            
            # Assert
            assert response.status_code == 500
            data = response.json()
            assert "Database connection failed" in data["detail"]
    
    def test_get_global_stats_success(self, client, mock_db_service):
        """Test successful retrieval of global token statistics."""
        # Arrange
        mock_stats = {
            "total_tokens": 10000,
            "total_input_tokens": 6000,
            "total_output_tokens": 4000,
            "total_sessions": 25,
            "models_used": ["gpt-4", "claude-3.5-sonnet"],
            "providers_used": ["openai", "anthropic"],
            "last_updated": "2024-01-15T10:30:00"
        }
        mock_db_service.get_global_token_stats.return_value = mock_stats
        
        with patch('app.api.tokens.db_service', mock_db_service):
            # Act
            response = client.get("/api/v1/tokens/stats")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert data["total_tokens"] == 10000
            assert data["total_input_tokens"] == 6000
            assert data["total_output_tokens"] == 4000
            assert data["total_sessions"] == 25
            assert "gpt-4" in data["models_used"]
            assert "openai" in data["providers_used"]
            assert data["last_updated"] == "2024-01-15T10:30:00"
            mock_db_service.get_global_token_stats.assert_called_once()
    
    def test_get_global_stats_empty(self, client, mock_db_service):
        """Test global stats when no data exists."""
        # Arrange
        mock_stats = {
            "total_tokens": 0,
            "total_input_tokens": 0,
            "total_output_tokens": 0,
            "total_sessions": 0,
            "models_used": [],
            "providers_used": [],
            "last_updated": None
        }
        mock_db_service.get_global_token_stats.return_value = mock_stats
        
        with patch('app.api.tokens.db_service', mock_db_service):
            # Act
            response = client.get("/api/v1/tokens/stats")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert data["total_tokens"] == 0
            assert data["models_used"] == []
            assert data["last_updated"] is None
    
    def test_get_global_stats_database_error(self, client, mock_db_service):
        """Test global stats retrieval with database error."""
        # Arrange
        mock_db_service.get_global_token_stats.side_effect = Exception("Database query failed")
        
        with patch('app.api.tokens.db_service', mock_db_service):
            # Act
            response = client.get("/api/v1/tokens/stats")
            
            # Assert
            assert response.status_code == 500
            data = response.json()
            assert "Database query failed" in data["detail"]
    
    def test_get_project_usage_success(self, client, mock_db_service, sample_token_usage):
        """Test successful retrieval of project token usage."""
        # Arrange
        project_id = "test-project-id"
        usage_records = [sample_token_usage]
        mock_db_service.get_project_token_usage.return_value = usage_records
        
        with patch('app.api.tokens.db_service', mock_db_service):
            # Act
            response = client.get(f"/api/v1/tokens/project/{project_id}")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert data["project_id"] == project_id
            assert data["total_tokens"] == 150
            assert data["input_tokens"] == 100
            assert data["output_tokens"] == 50
            assert len(data["records"]) == 1
            assert data["records"][0]["session_id"] == "test-session-id"
            mock_db_service.get_project_token_usage.assert_called_once_with(project_id)
    
    def test_get_project_usage_empty(self, client, mock_db_service):
        """Test project usage retrieval when no records exist."""
        # Arrange
        project_id = "empty-project-id"
        mock_db_service.get_project_token_usage.return_value = []
        
        with patch('app.api.tokens.db_service', mock_db_service):
            # Act
            response = client.get(f"/api/v1/tokens/project/{project_id}")
            
            # Assert
            assert response.status_code == 200
            data = response.json()
            assert data["project_id"] == project_id
            assert data["total_tokens"] == 0
            assert data["records"] == []
    
    def test_token_usage_aggregation(self, sample_token_usage):
        """Test token usage aggregation logic."""
        # Arrange
        usage_records = [
            sample_token_usage,
            Mock(input_tokens=50, output_tokens=25, total_tokens=75),
            Mock(input_tokens=200, output_tokens=100, total_tokens=300)
        ]
        
        # Act
        total_tokens = sum(record.total_tokens for record in usage_records)
        total_input = sum(record.input_tokens for record in usage_records)
        total_output = sum(record.output_tokens for record in usage_records)
        
        # Assert
        assert total_tokens == 525  # 150 + 75 + 300
        assert total_input == 350   # 100 + 50 + 200
        assert total_output == 175  # 50 + 25 + 100
    
    def test_token_usage_record_serialization(self, sample_token_usage):
        """Test token usage record serialization for API response."""
        # Act
        serialized = {
            "id": sample_token_usage.id,
            "model": sample_token_usage.model,
            "provider": sample_token_usage.provider,
            "input_tokens": sample_token_usage.input_tokens,
            "output_tokens": sample_token_usage.output_tokens,
            "total_tokens": sample_token_usage.total_tokens,
            "created_at": sample_token_usage.created_at.isoformat() if sample_token_usage.created_at else None
        }
        
        # Assert
        assert serialized["id"] == "test-usage-id"
        assert serialized["model"] == "gpt-4"
        assert serialized["provider"] == "openai"
        assert serialized["input_tokens"] == 100
        assert serialized["output_tokens"] == 50
        assert serialized["total_tokens"] == 150
        assert serialized["created_at"] is not None