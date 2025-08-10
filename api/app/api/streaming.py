import json
import uuid
import os
from fastapi import APIRouter, WebSocket, WebSocketDisconnect
from app.agents.react_agent import ReActAgent
from ..config import PROJECTS_DIR, MODEL_NAME
from app.database.service import db_service
from app.database.models import (
    ConversationMessageCreate, TokenUsageCreate, ProjectCreate, ChatRequest
)
import random
from app.utils.docker_route import deploy_app

router = APIRouter()

@router.websocket("/stream/{project_id}")
async def websocket_stream(websocket: WebSocket, project_id: str):
    await websocket.accept()
    
    # Generate session ID
    session_id = str(uuid.uuid4())
    
    # Get project details
    project = db_service.get_project_by_id(project_id)
    if not project:
        await websocket.close(code=1003, reason="Project not found")
        return
    
    # Get project path
    project_path = os.path.abspath(os.path.join(PROJECTS_DIR, project.name))
    
    # Initialize agent with project context and container name
    agent = ReActAgent(project_path=project_path, container_name=project.docker_container)
    
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
            model = payload.get("model", MODEL_NAME)
            provider = payload.get("provider", "openai")
            
            # Store user message
            user_message = ConversationMessageCreate(
                project_id=project_id,
                role="user",
                content=message,
                message_type="chat",
                model=model,
                provider=provider
            )
            db_service.create_conversation_message(user_message)
            
            # Get chat history summary for context
            chat_summary = db_service.get_chat_summary(project_id)
            
            # Enhance the message with chat history context if available
            enhanced_message = message
            if chat_summary:
                enhanced_message = f"""Previous conversation context:
                                    {chat_summary}
                                    Current user request: {message}
                                    Please consider the previous conversation context when responding to the current request."""
                
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
            
            await websocket.send_json({
                "type": "status",
                "content": "AI agent is thinking...",
                "session_id": session_id,
                "project_id": project_id
            })
            
            async for chunk in agent.stream_response(enhanced_message, project_path, project.docker_container):
                try:
                    # Process LangChain streaming chunks
                    if isinstance(chunk, dict):
                        # Handle different chunk types from LangChain
                        if chunk.get("type") == "content":
                            content = chunk.get("content", "")
                            if content and content.strip():
                                full_response += content
                                await websocket.send_json({
                                    "type": "agent_response",
                                    "content": content,
                                    "session_id": session_id,
                                    "project_id": project_id,
                                    "agent_type": "react"
                                })
                        
                        # Extract token usage if available
                        if "input_tokens" in chunk:
                            input_tokens += chunk.get("input_tokens", 0)
                        if "output_tokens" in chunk:
                            output_tokens += chunk.get("output_tokens", 0)
                    
                    # Handle raw string content
                    elif isinstance(chunk, str) and chunk.strip():
                        full_response += chunk
                        await websocket.send_json({
                            "type": "agent_response",
                            "content": chunk,
                            "session_id": session_id,
                            "project_id": project_id,
                            "agent_type": "react"
                        })
                    
                    # Handle LangChain log patches
                    elif hasattr(chunk, 'ops') and chunk.ops:
                        for op in chunk.ops:
                            if op.get('op') == 'add' and 'content' in op.get('value', {}):
                                content = op['value']['content']
                                if isinstance(content, str) and content.strip():
                                    full_response += content
                                    await websocket.send_json({
                                        "type": "agent_response",
                                        "content": content,
                                        "session_id": session_id,
                                        "project_id": project_id,
                                        "agent_type": "react"
                                    })
                
                except Exception as chunk_error:
                    print(f"Error processing chunk: {chunk_error}")
                    # Send the raw chunk for debugging if needed
                    await websocket.send_json({
                        "type": "debug",
                        "content": f"Debug: {str(chunk)[:200]}...",
                        "session_id": session_id,
                        "project_id": project_id
                    })
            
            # Store assistant response (only if it's actual content, not status messages)
            if full_response.strip():
                assistant_message = ConversationMessageCreate(
                    project_id=project_id,
                    role="assistant",
                    content=full_response,
                    message_type="chat",
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
        deploy_result = deploy_app("react-shadcn-template", fancy_name, fancy_name.lower(), int(port))
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
    
    session_id = str(uuid.uuid4())
    
    # Store the initial user message
    user_message = ConversationMessageCreate(
        project_id=project.id,
        role="user",
        content=request.message,
        message_type="chat",
        model="anthropic/claude-3.5-sonnet",
        provider="openrouter"
    )
    db_service.create_conversation_message(user_message)
    
    # Store initial AI response indicating project creation
    initial_ai_response = ConversationMessageCreate(
        project_id=project.id,
        role="assistant",
        content=f"I've created your project '{project.name}' and set up the development environment. The container is starting up and will be ready shortly. I'll help you build your application step by step.",
        message_type="chat",
        model="anthropic/claude-3.5-sonnet",
        provider="openrouter"
    )
    db_service.create_conversation_message(initial_ai_response)
    
    return {
        "project_id": project.id,
        "project_name": project.name,
        "project_path": project_path,
        "url": f"http://localhost:{port}",
        "session_id": session_id,
        "initial_message": request.message
    }