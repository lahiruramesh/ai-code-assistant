import os
from fastapi import APIRouter, HTTPException
from fastapi.responses import JSONResponse
from app.database.service import db_service
from ..config import PROJECTS_DIR, MODEL_NAME
from ..utils.docker_route import ensure_container_running, get_container_status_for_project, delete_project_and_cleanup
import random
from app.utils.docker_route import deploy_app
from app.database.models import (
    ConversationMessageCreate,
    ProjectCreate,
)

router = APIRouter()

@router.get("")
async def get_projects():
    """Get all projects from database"""
    projects = db_service.get_all_projects()
    return JSONResponse(content={
        "projects": [
            {
                "id": p.id,
                "name": p.name,
                "template": p.template,
                "status": p.status,
                "docker_container": p.docker_container,
                "port": p.port,
                "url": f"http://localhost:{p.port}" if p.port else None,
                "created_at": p.created_at.isoformat() if p.created_at else None,
                "updated_at": p.updated_at.isoformat() if p.updated_at else None
            }
            for p in projects
        ]
    })

@router.post("/")
async def create_project(project_data: ProjectCreate):
    """Create a new project"""
    try:
        fancy_name = db_service.generate_fancy_project_name(project_data.message)
        project_data.name = fancy_name
        project = db_service.create_project(project_data)
        
        # Check port availability
        # TODO: Implement a more robust port checking mechanism
        port = random.randint(8084, 9999)
        try:
            deploy_result = deploy_app("react-shadcn-template", fancy_name, fancy_name.lower(), int(port))
        except Exception as e:
            return {
                "error": str(e)
            }
        container_name = deploy_result.get("container_name")
        project.docker_container = container_name
        project.name = fancy_name
        project.port = port
        project.status = "created"
        db_service.update_project(project.id, project)
        
        user_message = ConversationMessageCreate(
            project_id=project.id,
            role="user",
            content=project_data.message,
            message_type="chat",
            model=MODEL_NAME,
            provider="openrouter"
        )
        db_service.create_conversation_message(user_message)
        return JSONResponse(content={
            "message": "Project created successfully",
            "id": project.id,
            "name": project.name,
            "template": project.template,
            "docker_container": project.docker_container,
            "port": project.port
        }, status_code=201)
    
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

@router.delete("/{project_id}")
async def delete_project(project_id: str):
    """Delete a project and cleanup all associated resources"""
    try:
        # Get project details before deletion
        project = db_service.get_project_by_id(project_id)
        if not project:
            raise HTTPException(status_code=404, detail="Project not found")
        
        # Cleanup Docker container, image and project files
        cleanup_result = {"container_removed": False, "image_removed": False, "files_removed": False, "errors": []}
        
        if project.docker_container or project.name:
            project_path = os.path.join(PROJECTS_DIR, project.name) if project.name else None
            
            try:
                cleanup_result = delete_project_and_cleanup(
                    container_name=project.docker_container,
                    project_path=project_path
                )
            except Exception as e:
                cleanup_result["errors"].append(f"Cleanup failed: {str(e)}")
        
        # Delete project from database
        db_service.delete_project(project_id)
        
        return JSONResponse(content={
            "message": "Project deleted successfully",
            "project_id": project_id,
            "cleanup_result": cleanup_result
        })
        
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/{project_id}")
async def get_project(project_id: str):
    """Get a specific project by ID and ensure container is running"""
    project = db_service.get_project_by_id(project_id)
    if not project:
        raise HTTPException(status_code=404, detail="Project not found")
    
    # Get container status and ensure it's running
    container_info = None
    container_started = False
    
    if project.docker_container:
        container_status = get_container_status_for_project(project.docker_container)
        container_info = container_status
        
        # If container needs to be started, start it
        if container_status["needs_start"]:
            start_result = ensure_container_running(project.docker_container)
            container_started = start_result["success"]
            if container_started:
                container_info["running"] = True
                container_info["status"] = "Started automatically"
    
    return JSONResponse(content={
        "id": project.id,
        "name": project.name,
        "template": project.template,
        "status": project.status,
        "docker_container": project.docker_container,
        "port": project.port,
        "url": f"http://localhost:{project.port}" if project.port else None,
        "created_at": project.created_at.isoformat() if project.created_at else None,
        "updated_at": project.updated_at.isoformat() if project.updated_at else None,
        "container_info": container_info,
        "container_started": container_started
    })

@router.get("/{project_name}/preview")
async def get_project_preview(project_name: str):
    """Get project preview URL by project name or ID"""
    # Try to find project by name first
    project = db_service.get_project_by_name(project_name)    
    if not project:
        raise HTTPException(status_code=404, detail="Project not found")
    
    project_path = os.path.abspath(os.path.join(PROJECTS_DIR, project.name))
    preview_url = f"http://localhost:{project.port}" if project.port else f"http://localhost:3000/{project.name}"
    
    return JSONResponse(content={
        "preview_url": preview_url,
        "host_path": project_path,
        "project_name": project.name
    })

def build_file_tree(base_path, current_path, name):
    """Build file tree structure"""
    full_path = os.path.join(base_path, current_path) if current_path else base_path
    node = {"name": name, "path": current_path or ""}
    
    if os.path.isdir(full_path):
        node["type"] = "folder"
        try:
            children = []
            for child in sorted(os.listdir(full_path)):
                if not child.startswith('.'):  # Skip hidden files
                    child_path = os.path.join(current_path, child) if current_path else child
                    children.append(build_file_tree(base_path, child_path, child))
            node["children"] = children
        except PermissionError:
            node["children"] = []
    else:
        node["type"] = "file"
        try:
            node["size"] = f"{os.path.getsize(full_path) / 1024:.2f} KB"
        except (OSError, FileNotFoundError):
            node["size"] = "0 KB"
    
    return node

@router.get("/{project_name}/files")
async def get_project_files(project_name: str, source: str = None):
    """Get project file structure by project name"""
    # Try to find project by name first
    project = db_service.get_project_by_name(project_name)
    if not project:
        # If not found by name, try by ID if it's numeric
        try:
            project_id = int(project_name)
            project = db_service.get_project_by_id(project_id)
        except ValueError:
            pass
    
    if not project:
        raise HTTPException(status_code=404, detail="Project not found")
    
    project_path = os.path.join(PROJECTS_DIR, project.name)
    if not os.path.isdir(project_path):
        raise HTTPException(status_code=404, detail="Project directory not found")
    
    try:
        files = []
        for entry in sorted(os.listdir(project_path)):
            if not entry.startswith('.'):  # Skip hidden files
                files.append(build_file_tree(project_path, entry, entry))
        
        return JSONResponse(content={"files": files})
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error reading project files: {str(e)}")

@router.get("/{project_name}/files/{file_path:path}")
async def get_file_content(project_name: str, file_path: str, source: str = None):
    """Get content of a specific file by project name"""
    # Try to find project by name first
    project = db_service.get_project_by_name(project_name)
    if not project:
        # If not found by name, try by ID if it's numeric
        try:
            project_id = int(project_name)
            project = db_service.get_project_by_id(project_id)
        except ValueError:
            pass
    
    if not project:
        raise HTTPException(status_code=404, detail="Project not found")
    
    full_path = os.path.join(PROJECTS_DIR, project.name, file_path)
    
    if not os.path.exists(full_path) or not os.path.isfile(full_path):
        raise HTTPException(status_code=404, detail="File not found")
    
    # Security check: ensure file is within project directory
    project_path = os.path.abspath(os.path.join(PROJECTS_DIR, project.name))
    full_path = os.path.abspath(full_path)
    if not full_path.startswith(project_path):
        raise HTTPException(status_code=403, detail="Access denied")
    
    try:
        with open(full_path, "r", encoding="utf-8") as f:
            content = f.read()
        return JSONResponse(content={"content": content, "file_path": file_path})
    except UnicodeDecodeError:
        # If it's a binary file, return info instead of content
        return JSONResponse(content={
            "content": "[Binary file - cannot display]",
            "file_path": file_path,
            "is_binary": True
        })
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error reading file: {str(e)}")

@router.get("/{project_id}/conversations")
@router.get("/{project_id}/messages")
async def get_project_messages(project_id: str):
    """Get all chat messages for a project"""
    project = db_service.get_project_by_id(project_id)
    if not project:
        raise HTTPException(status_code=404, detail="Project not found")
    
    messages = db_service.get_project_messages(project_id)
    
    return JSONResponse(content={
        "project_id": project_id,
        "project_name": project.name,
        "messages": [
            {
                "id": msg.id,
                "role": msg.role,
                "content": msg.content,
                "message_type": msg.message_type,
                "model": msg.model,
                "provider": msg.provider,
                "created_at": msg.created_at.isoformat() if msg.created_at else None
            }
            for msg in messages
        ]
    })

async def get_project_conversations(project_id: str):
    """Get all conversations for a project - Legacy endpoint"""
    project = db_service.get_project_by_id(project_id)
    if not project:
        raise HTTPException(status_code=404, detail="Project not found")
    
    messages = db_service.get_project_messages(project_id)
    
    return JSONResponse(content={
        "project_id": project_id,
        "conversations": [{
            "project_id": project_id,
            "message_count": len(messages),
            "last_message": messages[-1].content[:100] + "..." if messages and len(messages[-1].content) > 100 else (messages[-1].content if messages else ""),
            "last_updated": messages[-1].created_at.isoformat() if messages else None
        }] if messages else []
    })

@router.get("/{project_id}/conversations/{session_id}")
async def get_conversation_messages(project_id: int, session_id: str):
    """Get all messages for a specific conversation"""
    project = db_service.get_project_by_id(project_id)
    if not project:
        raise HTTPException(status_code=404, detail="Project not found")
    
    messages = db_service.get_conversation_messages(session_id)
    
    return JSONResponse(content={
        "session_id": session_id,
        "project_id": project_id,
        "messages": [
            {
                "id": msg.id,
                "role": msg.role,
                "content": msg.content,
                "model": msg.model,
                "provider": msg.provider,
                "created_at": msg.created_at.isoformat() if msg.created_at else None
            }
            for msg in messages
        ]
    })