from typing import Optional
from datetime import datetime

try:
    from pydantic import BaseModel
    PYDANTIC_AVAILABLE = True
except ImportError:
    PYDANTIC_AVAILABLE = False
    BaseModel = object

# For FastAPI compatibility, create Pydantic models if available
if PYDANTIC_AVAILABLE:
    class ChatRequest(BaseModel):
        message: str
        project_id: Optional[str] = None
        session_id: Optional[str] = None
        model: Optional[str] = None
        provider: Optional[str] = None
    
    class ProjectCreate(BaseModel):
        name: str
        template: str
        docker_container: Optional[str] = None
        port: Optional[int] = None
else:
    # Fallback classes without Pydantic
    class ChatRequest:
        def __init__(self, message: str, project_id: Optional[str] = None, 
                     session_id: Optional[str] = None, model: Optional[str] = None, 
                     provider: Optional[str] = None):
            self.message = message
            self.project_id = project_id
            self.session_id = session_id
            self.model = model
            self.provider = provider
    
    class ProjectCreate:
        def __init__(self, name: str, template: str, docker_container: Optional[str] = None, port: Optional[int] = None):
            self.name = name
            self.template = template
            self.docker_container = docker_container
            self.port = port

# Regular classes for internal use
class Project:
    def __init__(self, id: str, name: str, template: str, docker_container: Optional[str] = None, 
                 port: Optional[int] = None, status: str = "created", created_at: datetime = None, 
                 updated_at: datetime = None):
        self.id = id
        self.name = name
        self.template = template
        self.docker_container = docker_container
        self.port = port
        self.status = status
        self.created_at = created_at
        self.updated_at = updated_at

class ConversationMessageCreate:
    def __init__(self, project_id: str, role: str, content: str, 
                 message_type: str = "chat", model: Optional[str] = None, 
                 provider: Optional[str] = None):
        self.project_id = project_id
        self.role = role
        self.content = content
        self.message_type = message_type
        self.model = model
        self.provider = provider

class ConversationMessage:
    def __init__(self, id: str, project_id: str, role: str, content: str, 
                 message_type: str = "chat", model: Optional[str] = None, 
                 provider: Optional[str] = None, token_usage_id: Optional[str] = None, 
                 created_at: datetime = None, updated_at: datetime = None):
        self.id = id
        self.project_id = project_id
        self.role = role
        self.content = content
        self.message_type = message_type
        self.model = model
        self.provider = provider
        self.token_usage_id = token_usage_id
        self.created_at = created_at
        self.updated_at = updated_at

class TokenUsageCreate:
    def __init__(self, session_id: str, project_id: Optional[str] = None, model: str = "", 
                 provider: str = "", input_tokens: int = 0, output_tokens: int = 0, 
                 total_tokens: int = 0, request_type: str = "chat"):
        self.session_id = session_id
        self.project_id = project_id
        self.model = model
        self.provider = provider
        self.input_tokens = input_tokens
        self.output_tokens = output_tokens
        self.total_tokens = total_tokens
        self.request_type = request_type

class TokenUsage:
    def __init__(self, id: str, session_id: str, project_id: Optional[str] = None, 
                 model: str = "", provider: str = "", input_tokens: int = 0, 
                 output_tokens: int = 0, total_tokens: int = 0, request_type: str = "chat", 
                 created_at: datetime = None):
        self.id = id
        self.session_id = session_id
        self.project_id = project_id
        self.model = model
        self.provider = provider
        self.input_tokens = input_tokens
        self.output_tokens = output_tokens
        self.total_tokens = total_tokens
        self.request_type = request_type
        self.created_at = created_at

class ChatResponse:
    def __init__(self, type: str, content: str, session_id: str, project_id: Optional[str] = None):
        self.type = type
        self.content = content
        self.session_id = session_id
        self.project_id = project_id
