from fastapi import APIRouter, HTTPException, Depends
import httpx
import os
import tempfile
import shutil
from pathlib import Path
from git import Repo
from typing import Optional
from ..database.models import GitHubRepoCreate, GitHubRepository
from ..database.service import DatabaseService

router = APIRouter(prefix="/github", tags=["github"])
db_service = DatabaseService()

@router.post("/repositories")
async def create_github_repository(user_id: str, repo_data: GitHubRepoCreate):
    """Create a new GitHub repository and link it to a project"""
    # Get user's GitHub token
    user = await db_service.get_user_by_id(user_id)
    if not user or not user.github_token:
        raise HTTPException(status_code=400, detail="GitHub not connected")
    
    async with httpx.AsyncClient() as client:
        # Create repository on GitHub
        create_repo_response = await client.post(
            "https://api.github.com/user/repos",
            headers={
                "Authorization": f"Bearer {user.github_token}",
                "Accept": "application/vnd.github.v3+json"
            },
            json={
                "name": repo_data.name,
                "description": repo_data.description,
                "private": repo_data.private,
                "auto_init": True,
                "gitignore_template": "Node"
            }
        )
        
        if create_repo_response.status_code != 201:
            error_detail = create_repo_response.json().get("message", "Failed to create repository")
            raise HTTPException(status_code=400, detail=f"GitHub API error: {error_detail}")
        
        repo_info = create_repo_response.json()
        
        # Save repository info to database
        github_repo = GitHubRepository(
            id=f"gh_{repo_info['id']}",
            user_id=user_id,
            project_id="",  # Will be updated when linked to a project
            repo_name=repo_info["name"],
            repo_url=repo_info["html_url"],
            clone_url=repo_info["clone_url"]
        )
        
        saved_repo = await db_service.create_github_repository(github_repo)
        
        return {
            "id": saved_repo.id,
            "name": repo_info["name"],
            "url": repo_info["html_url"],
            "clone_url": repo_info["clone_url"],
            "private": repo_info["private"]
        }

@router.get("/repositories")
async def list_github_repositories(user_id: str):
    """List user's GitHub repositories"""
    user = await db_service.get_user_by_id(user_id)
    if not user or not user.github_token:
        raise HTTPException(status_code=400, detail="GitHub not connected")
    
    async with httpx.AsyncClient() as client:
        response = await client.get(
            "https://api.github.com/user/repos",
            headers={
                "Authorization": f"Bearer {user.github_token}",
                "Accept": "application/vnd.github.v3+json"
            },
            params={"sort": "updated", "per_page": 50}
        )
        
        if response.status_code != 200:
            raise HTTPException(status_code=400, detail="Failed to fetch repositories")
        
        repos = response.json()
        return [
            {
                "id": repo["id"],
                "name": repo["name"],
                "full_name": repo["full_name"],
                "url": repo["html_url"],
                "clone_url": repo["clone_url"],
                "private": repo["private"],
                "updated_at": repo["updated_at"]
            }
            for repo in repos
        ]

@router.post("/repositories/{repo_name}/commit")
async def commit_and_push_to_github(
    user_id: str, 
    repo_name: str, 
    project_id: str,
    commit_message: str,
    files_to_commit: Optional[list] = None
):
    """Commit project files to GitHub repository"""
    user = await db_service.get_user_by_id(user_id)
    if not user or not user.github_token:
        raise HTTPException(status_code=400, detail="GitHub not connected")
    
    # Get project details
    project = await db_service.get_project_by_id(project_id)
    if not project:
        raise HTTPException(status_code=404, detail="Project not found")
    
    # Get GitHub repository info
    github_repo = await db_service.get_github_repository_by_name(user_id, repo_name)
    if not github_repo:
        raise HTTPException(status_code=404, detail="GitHub repository not found")
    
    try:
        # Create temporary directory
        with tempfile.TemporaryDirectory() as temp_dir:
            temp_path = Path(temp_dir)
            
            # Clone the repository
            repo_url_with_token = github_repo.clone_url.replace(
                "https://", f"https://{user.github_token}@"
            )
            
            repo = Repo.clone_from(repo_url_with_token, temp_path)
            
            # Copy project files to repository
            project_path = Path(f"./projects/{project_id}")
            if project_path.exists():
                # Copy all files from project to repo
                for item in project_path.rglob("*"):
                    if item.is_file() and not item.name.startswith('.'):
                        relative_path = item.relative_to(project_path)
                        dest_path = temp_path / relative_path
                        dest_path.parent.mkdir(parents=True, exist_ok=True)
                        shutil.copy2(item, dest_path)
            
            # Add README.md if it doesn't exist
            readme_path = temp_path / "README.md"
            if not readme_path.exists():
                readme_content = f"""# {project.name}

This project was created using the Code Editing Agent.

## Template
- **Template**: {project.template}
- **Created**: {project.created_at}

## Getting Started

Follow the instructions below to run this project locally.

### Prerequisites
- Node.js (for web projects)
- Docker (if using containerized setup)

### Installation
1. Clone this repository
2. Install dependencies
3. Start the development server

## Deployment
This project can be deployed to Vercel or other hosting platforms.
"""
                readme_path.write_text(readme_content)
            
            # Add all files to git
            repo.git.add(A=True)
            
            # Check if there are changes to commit
            if repo.is_dirty() or len(repo.untracked_files) > 0:
                # Commit changes
                repo.index.commit(commit_message)
                
                # Push to remote
                origin = repo.remote(name='origin')
                origin.push()
                
                # Update database with commit info
                await db_service.update_project_github_repo(project_id, github_repo.id)
                
                return {
                    "success": True,
                    "message": "Files committed and pushed successfully",
                    "repository_url": github_repo.repo_url,
                    "commit_message": commit_message
                }
            else:
                return {
                    "success": True,
                    "message": "No changes to commit",
                    "repository_url": github_repo.repo_url
                }
                
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to commit to GitHub: {str(e)}")

@router.post("/repositories/{repo_name}/link")
async def link_repository_to_project(user_id: str, repo_name: str, project_id: str):
    """Link an existing GitHub repository to a project"""
    github_repo = await db_service.get_github_repository_by_name(user_id, repo_name)
    if not github_repo:
        raise HTTPException(status_code=404, detail="GitHub repository not found")
    
    await db_service.update_github_repository_project(github_repo.id, project_id)
    await db_service.update_project_github_repo(project_id, github_repo.id)
    
    return {
        "success": True,
        "message": f"Repository {repo_name} linked to project",
        "repository_url": github_repo.repo_url
    }

@router.delete("/repositories/{repo_name}")
async def delete_github_repository(user_id: str, repo_name: str):
    """Delete a GitHub repository"""
    user = await db_service.get_user_by_id(user_id)
    if not user or not user.github_token:
        raise HTTPException(status_code=400, detail="GitHub not connected")
    
    async with httpx.AsyncClient() as client:
        response = await client.delete(
            f"https://api.github.com/repos/{user.github_username}/{repo_name}",
            headers={
                "Authorization": f"Bearer {user.github_token}",
                "Accept": "application/vnd.github.v3+json"
            }
        )
        
        if response.status_code == 204:
            # Remove from database
            await db_service.delete_github_repository_by_name(user_id, repo_name)
            return {"success": True, "message": "Repository deleted successfully"}
        else:
            raise HTTPException(status_code=400, detail="Failed to delete repository")
