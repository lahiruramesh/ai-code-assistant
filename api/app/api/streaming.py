import json
import uuid
import os
from fastapi import APIRouter, WebSocket, WebSocketDisconnect
from app.agents.react_agent import ReActAgent
from app.database.service import db_service
from app.database.models import (
    ConversationMessageCreate, TokenUsageCreate, ProjectCreate, ChatRequest
)
import random
from app.utils.docker_route import deploy_app

router = APIRouter()

@router.websocket("/stream/{project_id}")
async def websocket_stream(websocket: WebSocket, project_id: int):
    await websocket.accept()
    
    # Generate session ID
    session_id = str(uuid.uuid4())
    
    # Get project details
    project = db_service.get_project_by_id(project_id)
    if not project:
        await websocket.close(code=1003, reason="Project not found")
        return
    
    # Get project path
    project_path = os.path.abspath(os.path.join(os.getenv("PROJECTS_DIR"), project.name))
    
    # Initialize agent with project context
    agent = ReActAgent(project_path=project_path)
    
    try:
        await websocket.send_json({
            "type": "session_started",
            "session_id": session_id,
            "project_id": project_id,
            "project_name": project.name
        })
        
        while True:
            data = await websocket.receive_text()
            payload = json.loads(data)
            
            message = payload.get("message", "")
            model = payload.get("model", "gpt-4")
            provider = payload.get("provider", "openai")
            
            # Store user message
            user_message = ConversationMessageCreate(
                session_id=session_id,
                project_id=project_id,
                role="user",
                content=message,
                model=model,
                provider=provider
            )
            db_service.create_conversation_message(user_message)
            
            # Send acknowledgment
            await websocket.send_json({
                "type": "message_received",
                "content": message,
                "session_id": session_id
            })
            
            # Stream agent response
            full_response = ""
            input_tokens = 0
            output_tokens = 0
            
            async for chunk in agent.stream_response(message, project_path):
                # Parse chunk for token usage and content
                if isinstance(chunk, dict):
                    content = chunk.get("content", "")
                    if content:
                        full_response += content
                        await websocket.send_json({
                            "type": "agent_chunk",
                            "content": content,
                            "session_id": session_id,
                            "project_id": project_id
                        })
                    
                    # Extract token usage if available
                    if "input_tokens" in chunk:
                        input_tokens += chunk.get("input_tokens", 0)
                    if "output_tokens" in chunk:
                        output_tokens += chunk.get("output_tokens", 0)
                else:
                    # Handle string chunks
                    full_response += str(chunk)
                    await websocket.send_json({
                        "type": "agent_chunk",
                        "content": str(chunk),
                        "session_id": session_id,
                        "project_id": project_id
                    })
            
            # Store assistant response
            assistant_message = ConversationMessageCreate(
                session_id=session_id,
                project_id=project_id,
                role="assistant",
                content=full_response,
                model=model,
                provider=provider
            )
            db_service.create_conversation_message(assistant_message)
            
            # Store token usage
            total_tokens = input_tokens + output_tokens
            if total_tokens > 0:
                token_usage = TokenUsageCreate(
                    session_id=session_id,
                    project_id=project_id,
                    model=model,
                    provider=provider,
                    input_tokens=input_tokens,
                    output_tokens=output_tokens,
                    total_tokens=total_tokens
                )
                db_service.create_token_usage(token_usage)
            
            # Send completion signal
            await websocket.send_json({
                "type": "response_complete",
                "session_id": session_id,
                "token_usage": {
                    "input_tokens": input_tokens,
                    "output_tokens": output_tokens,
                    "total_tokens": total_tokens
                }
            })
            
    except WebSocketDisconnect:
        print(f"Client disconnected from session {session_id}")
    except Exception as e:
        print(f"An error occurred in session {session_id}: {e}")
        await websocket.close(code=1011, reason=str(e))

@router.post("/create-session")
async def create_chat_session(request: ChatRequest):
    """Create a new chat session with a project"""
    
    # Generate fancy project name based on the query
    fancy_name = db_service.generate_fancy_project_name(request.message)
    
    # Check port availability
    port = random.randint(8084, 9999)
    try:
        deploy_result = deploy_app("nextjs-shadcn-template", fancy_name, fancy_name.lower(), int(port))
    except Exception as e:
        return {
            "error": str(e)
        }   
        
    project_path = deploy_result.get("project_path")
    container_name = deploy_result.get("container_name")
    # Create new project
    project_data = ProjectCreate(
        name=fancy_name,
        template="nextjs-shadcn-template",
        docker_container=container_name,
        port=port
    )
    
    project = db_service.create_project(project_data)
    
    return {
        "project_id": project.id,
        "project_name": project.name,
        "project_path": project_path,
        "url": f"http://localhost:{port}",
        "session_id": str(uuid.uuid4())
    }