import os
from fastapi import APIRouter, HTTPException
from fastapi.responses import JSONResponse
from app.database.service import db_service
from app.database.models import ProjectCreate

router = APIRouter()
PROJECTS_DIR = os.getenv("PROJECTS_DIR")

@router.get("/")
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
        db_service.create_project(project_data)
        return JSONResponse(content={
            "message": "Project created successfully",
            "project": {
                "name": project_data.name,
                "template": project_data.template,
                "docker_container": project_data.docker_container,
                "port": project_data.port
            }
        }, status_code=201)
    
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

@router.get("/{project_id}")
async def get_project(project_id: int):
    """Get a specific project by ID"""
    project = db_service.get_project_by_id(project_id)
    if not project:
        raise HTTPException(status_code=404, detail="Project not found")
    
    return JSONResponse(content={
        "id": project.id,
        "name": project.name,
        "template": project.template,
        "status": project.status,
        "docker_container": project.docker_container,
        "port": project.port,
        "created_at": project.created_at.isoformat() if project.created_at else None,
        "updated_at": project.updated_at.isoformat() if project.updated_at else None
    })

@router.get("/{project_name}/preview")
async def get_project_preview(project_name: str):
    """Get project preview URL by project name or ID"""
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
async def get_project_conversations(project_id: int):
    """Get all conversations for a project"""
    project = db_service.get_project_by_id(project_id)
    if not project:
        raise HTTPException(status_code=404, detail="Project not found")
    
    # Get unique session IDs for this project
    conn = db_service.conn
    query = """
    SELECT DISTINCT session_id, MIN(created_at) as first_message
    FROM conversation_messages 
    WHERE project_id = ? 
    GROUP BY session_id
    ORDER BY first_message DESC
    """
    sessions = conn.execute(query, [project_id]).fetchall()
    
    conversations = []
    for session_id, first_message in sessions:
        messages = db_service.get_conversation_messages(session_id)
        if messages:
            conversations.append({
                "session_id": session_id,
                "first_message_time": first_message,
                "message_count": len(messages),
                "last_message": messages[-1].content[:100] + "..." if len(messages[-1].content) > 100 else messages[-1].content
            })
    
    return JSONResponse(content={"conversations": conversations})

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