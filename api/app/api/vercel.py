from fastapi import APIRouter, HTTPException
import httpx
import os
from typing import Optional, Dict, Any
from ..database.models import VercelDeployment, VercelDeploymentRecord
from ..database.service import DatabaseService

router = APIRouter(prefix="/vercel", tags=["vercel"])
db_service = DatabaseService()

@router.post("/deployments")
async def create_vercel_deployment(
    user_id: str,
    project_id: str,
    deployment_data: VercelDeployment
):
    """Create a new Vercel deployment from GitHub repository"""
    user = await db_service.get_user_by_id(user_id)
    if not user or not user.vercel_token:
        raise HTTPException(status_code=400, detail="Vercel not connected")
    
    # Get project details
    project = await db_service.get_project_by_id(project_id)
    if not project:
        raise HTTPException(status_code=404, detail="Project not found")
    
    # Get GitHub repository info
    github_repo = await db_service.get_github_repository_by_project(project_id)
    if not github_repo:
        raise HTTPException(status_code=400, detail="Project not linked to GitHub repository")
    
    async with httpx.AsyncClient() as client:
        # Create deployment on Vercel
        deployment_config = {
            "name": deployment_data.name,
            "gitSource": {
                "type": "github",
                "repo": deployment_data.github_repo,
                "ref": deployment_data.branch
            },
            "projectSettings": {
                "framework": _get_framework_from_template(project.template),
                "buildCommand": _get_build_command(project.template),
                "outputDirectory": _get_output_directory(project.template),
                "installCommand": _get_install_command(project.template)
            }
        }
        
        headers = {
            "Authorization": f"Bearer {user.vercel_token}",
            "Content-Type": "application/json"
        }
        
        if user.vercel_team_id:
            headers["X-Vercel-Team-Id"] = user.vercel_team_id
        
        # Create project on Vercel
        create_project_response = await client.post(
            "https://api.vercel.com/v9/projects",
            headers=headers,
            json=deployment_config
        )
        
        if create_project_response.status_code != 200:
            error_detail = create_project_response.json().get("error", {}).get("message", "Failed to create Vercel project")
            raise HTTPException(status_code=400, detail=f"Vercel API error: {error_detail}")
        
        project_info = create_project_response.json()
        
        # Trigger initial deployment
        deploy_response = await client.post(
            f"https://api.vercel.com/v13/deployments",
            headers=headers,
            json={
                "name": deployment_data.name,
                "gitSource": {
                    "type": "github",
                    "repo": deployment_data.github_repo,
                    "ref": deployment_data.branch
                },
                "projectSettings": deployment_config["projectSettings"]
            }
        )
        
        if deploy_response.status_code != 200:
            error_detail = deploy_response.json().get("error", {}).get("message", "Failed to create deployment")
            raise HTTPException(status_code=400, detail=f"Vercel deployment error: {error_detail}")
        
        deployment_info = deploy_response.json()
        
        # Save deployment info to database
        vercel_deployment = VercelDeploymentRecord(
            id=f"vrc_{deployment_info['id']}",
            user_id=user_id,
            project_id=project_id,
            deployment_id=deployment_info["id"],
            deployment_url=f"https://{deployment_info['url']}",
            status=deployment_info.get("readyState", "QUEUED")
        )
        
        saved_deployment = await db_service.create_vercel_deployment(vercel_deployment)
        await db_service.update_project_vercel_deployment(project_id, saved_deployment.id)
        
        return {
            "id": saved_deployment.id,
            "deployment_id": deployment_info["id"],
            "url": f"https://{deployment_info['url']}",
            "status": deployment_info.get("readyState", "QUEUED"),
            "project_name": deployment_data.name
        }

@router.get("/deployments")
async def list_vercel_deployments(user_id: str, project_id: Optional[str] = None):
    """List user's Vercel deployments"""
    user = await db_service.get_user_by_id(user_id)
    if not user or not user.vercel_token:
        raise HTTPException(status_code=400, detail="Vercel not connected")
    
    async with httpx.AsyncClient() as client:
        headers = {
            "Authorization": f"Bearer {user.vercel_token}",
            "Content-Type": "application/json"
        }
        
        if user.vercel_team_id:
            headers["X-Vercel-Team-Id"] = user.vercel_team_id
        
        response = await client.get(
            "https://api.vercel.com/v6/deployments",
            headers=headers,
            params={"limit": 20}
        )
        
        if response.status_code != 200:
            raise HTTPException(status_code=400, detail="Failed to fetch deployments")
        
        deployments_data = response.json()
        deployments = []
        
        for deployment in deployments_data.get("deployments", []):
            deployment_info = {
                "id": deployment["uid"],
                "url": f"https://{deployment['url']}",
                "status": deployment.get("readyState", "UNKNOWN"),
                "created_at": deployment.get("createdAt"),
                "project_name": deployment.get("name", ""),
                "source": deployment.get("source", ""),
                "target": deployment.get("target", "production")
            }
            
            # If filtering by project_id, only include matching deployments
            if project_id:
                db_deployment = await db_service.get_vercel_deployment_by_deployment_id(deployment["uid"])
                if db_deployment and db_deployment.project_id == project_id:
                    deployments.append(deployment_info)
            else:
                deployments.append(deployment_info)
        
        return deployments

@router.get("/deployments/{deployment_id}")
async def get_vercel_deployment(user_id: str, deployment_id: str):
    """Get specific Vercel deployment details"""
    user = await db_service.get_user_by_id(user_id)
    if not user or not user.vercel_token:
        raise HTTPException(status_code=400, detail="Vercel not connected")
    
    async with httpx.AsyncClient() as client:
        headers = {
            "Authorization": f"Bearer {user.vercel_token}",
            "Content-Type": "application/json"
        }
        
        if user.vercel_team_id:
            headers["X-Vercel-Team-Id"] = user.vercel_team_id
        
        response = await client.get(
            f"https://api.vercel.com/v13/deployments/{deployment_id}",
            headers=headers
        )
        
        if response.status_code != 200:
            raise HTTPException(status_code=404, detail="Deployment not found")
        
        deployment = response.json()
        
        return {
            "id": deployment["uid"],
            "url": f"https://{deployment['url']}",
            "status": deployment.get("readyState", "UNKNOWN"),
            "created_at": deployment.get("createdAt"),
            "project_name": deployment.get("name", ""),
            "source": deployment.get("source", ""),
            "target": deployment.get("target", "production"),
            "build_logs": deployment.get("buildLogs", [])
        }

@router.post("/deployments/{deployment_id}/redeploy")
async def redeploy_vercel_deployment(user_id: str, deployment_id: str):
    """Redeploy a Vercel deployment"""
    user = await db_service.get_user_by_id(user_id)
    if not user or not user.vercel_token:
        raise HTTPException(status_code=400, detail="Vercel not connected")
    
    # Get deployment info from database
    deployment = await db_service.get_vercel_deployment_by_deployment_id(deployment_id)
    if not deployment:
        raise HTTPException(status_code=404, detail="Deployment not found")
    
    # Get project and GitHub repo info
    project = await db_service.get_project_by_id(deployment.project_id)
    github_repo = await db_service.get_github_repository_by_project(deployment.project_id)
    
    if not project or not github_repo:
        raise HTTPException(status_code=400, detail="Project or GitHub repository not found")
    
    async with httpx.AsyncClient() as client:
        headers = {
            "Authorization": f"Bearer {user.vercel_token}",
            "Content-Type": "application/json"
        }
        
        if user.vercel_team_id:
            headers["X-Vercel-Team-Id"] = user.vercel_team_id
        
        # Trigger new deployment
        deploy_response = await client.post(
            f"https://api.vercel.com/v13/deployments",
            headers=headers,
            json={
                "name": project.name,
                "gitSource": {
                    "type": "github",
                    "repo": github_repo.repo_name,
                    "ref": "main"
                },
                "projectSettings": {
                    "framework": _get_framework_from_template(project.template),
                    "buildCommand": _get_build_command(project.template),
                    "outputDirectory": _get_output_directory(project.template),
                    "installCommand": _get_install_command(project.template)
                }
            }
        )
        
        if deploy_response.status_code != 200:
            error_detail = deploy_response.json().get("error", {}).get("message", "Failed to redeploy")
            raise HTTPException(status_code=400, detail=f"Vercel redeploy error: {error_detail}")
        
        new_deployment = deploy_response.json()
        
        return {
            "id": new_deployment["id"],
            "url": f"https://{new_deployment['url']}",
            "status": new_deployment.get("readyState", "QUEUED"),
            "message": "Redeployment started successfully"
        }

@router.delete("/deployments/{deployment_id}")
async def delete_vercel_deployment(user_id: str, deployment_id: str):
    """Delete a Vercel deployment"""
    user = await db_service.get_user_by_id(user_id)
    if not user or not user.vercel_token:
        raise HTTPException(status_code=400, detail="Vercel not connected")
    
    async with httpx.AsyncClient() as client:
        headers = {
            "Authorization": f"Bearer {user.vercel_token}",
            "Content-Type": "application/json"
        }
        
        if user.vercel_team_id:
            headers["X-Vercel-Team-Id"] = user.vercel_team_id
        
        response = await client.delete(
            f"https://api.vercel.com/v13/deployments/{deployment_id}",
            headers=headers
        )
        
        if response.status_code == 200:
            # Remove from database
            await db_service.delete_vercel_deployment_by_deployment_id(deployment_id)
            return {"success": True, "message": "Deployment deleted successfully"}
        else:
            raise HTTPException(status_code=400, detail="Failed to delete deployment")

def _get_framework_from_template(template: str) -> str:
    """Get Vercel framework setting from project template"""
    template_lower = template.lower()
    if "nextjs" in template_lower or "next" in template_lower:
        return "nextjs"
    elif "react" in template_lower:
        return "create-react-app"
    elif "vue" in template_lower:
        return "vue"
    elif "angular" in template_lower:
        return "angular"
    elif "svelte" in template_lower:
        return "svelte"
    elif "nuxt" in template_lower:
        return "nuxtjs"
    else:
        return "other"

def _get_build_command(template: str) -> str:
    """Get build command from project template"""
    template_lower = template.lower()
    if "nextjs" in template_lower or "next" in template_lower:
        return "pnpm run build"
    elif "react" in template_lower:
        return "pnpm run build"
    elif "vue" in template_lower:
        return "npm run build"
    else:
        return "npm run build"

def _get_output_directory(template: str) -> str:
    """Get output directory from project template"""
    template_lower = template.lower()
    if "nextjs" in template_lower or "next" in template_lower:
        return ".next"
    elif "react" in template_lower:
        return "build"
    elif "vue" in template_lower:
        return "dist"
    else:
        return "dist"

def _get_install_command(template: str) -> str:
    """Get install command from project template"""
    return "npm install"
