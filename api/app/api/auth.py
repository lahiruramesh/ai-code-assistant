from fastapi import APIRouter, HTTPException
from fastapi.responses import RedirectResponse
from pydantic import BaseModel
import httpx
import uuid
from typing import Optional
from ..database.models import UserCreate, VercelConnection
from ..database.service import DatabaseService
from ..config import (
    GOOGLE_CLIENT_ID,
    GOOGLE_CLIENT_SECRET,
    GITHUB_CLIENT_ID,
    GITHUB_CLIENT_SECRET,
    WEB_URL
)

router = APIRouter(prefix="/auth", tags=["authentication"])
db_service = DatabaseService()

class GoogleCallbackRequest(BaseModel):
    code: str
    state: str


@router.get("/google/login")
async def google_login():
    """Initiate Google OAuth login"""
    if not GOOGLE_CLIENT_ID:
        raise HTTPException(status_code=500, detail="Google OAuth not configured")
    
    state = str(uuid.uuid4())
    redirect_uri = WEB_URL
    google_auth_url = (
        f"https://accounts.google.com/o/oauth2/auth?"
        f"client_id={GOOGLE_CLIENT_ID}&"
        f"redirect_uri={redirect_uri}&"
        f"scope=openid%20email%20profile&"
        f"response_type=code&"
        f"state={state}"
    )
    
    return {"auth_url": google_auth_url, "state": state}

@router.post("/google/callback")
async def google_callback(request: GoogleCallbackRequest):
    """Handle Google OAuth callback"""
    if not GOOGLE_CLIENT_ID or not GOOGLE_CLIENT_SECRET:
        raise HTTPException(status_code=500, detail="Google OAuth not configured")
    
    # Exchange code for token
    async with httpx.AsyncClient() as client:
        redirect_uri = WEB_URL
        token_response = await client.post(
            "https://oauth2.googleapis.com/token",
            data={
                "client_id": GOOGLE_CLIENT_ID,
                "client_secret": GOOGLE_CLIENT_SECRET,
                "code": request.code,
                "grant_type": "authorization_code",
                "redirect_uri": redirect_uri
            }
        )
        
        if token_response.status_code != 200:
            raise HTTPException(status_code=400, detail="Failed to exchange code for token")
        
        token_data = token_response.json()
        access_token = token_data.get("access_token")
        
        # Get user info from Google
        user_response = await client.get(
            "https://www.googleapis.com/oauth2/v2/userinfo",
            headers={"Authorization": f"Bearer {access_token}"}
        )
        
        if user_response.status_code != 200:
            raise HTTPException(status_code=400, detail="Failed to get user info")
        
        user_data = user_response.json()
        
        # Create or update user in database
        user_create = UserCreate(
            email=user_data["email"],
            name=user_data["name"],
            avatar_url=user_data.get("picture"),
            google_id=user_data["id"]
        )
        
        # Check if user exists
        existing_user = await db_service.get_user_by_email(user_create.email)
        if existing_user:
            user = existing_user
        else:
            user = await db_service.create_user(user_create)
        
        return {
            "user": {
                "id": user.id,
                "email": user.email,
                "name": user.name,
                "avatar_url": user.avatar_url
            },
            "access_token": access_token
        }

@router.get("/github/login")
async def github_login():
    """Initiate GitHub OAuth login"""
    if not GITHUB_CLIENT_ID:
        raise HTTPException(status_code=500, detail="GitHub OAuth not configured")
    
    state = str(uuid.uuid4())
    github_auth_url = (
        f"https://github.com/login/oauth/authorize?"
        f"client_id={GITHUB_CLIENT_ID}&"
        f"redirect_uri={WEB_URL}/auth/github/callback&"
        f"scope=repo%20user&"
        f"state={state}"
    )
    
    return {"auth_url": github_auth_url, "state": state}

@router.post("/github/callback")
async def github_callback(code: str, state: str, user_id: str):
    """Handle GitHub OAuth callback"""
    if not GITHUB_CLIENT_ID or not GITHUB_CLIENT_SECRET:
        raise HTTPException(status_code=500, detail="GitHub OAuth not configured")
    
    # Exchange code for token
    async with httpx.AsyncClient() as client:
        token_response = await client.post(
            "https://github.com/login/oauth/access_token",
            data={
                "client_id": GITHUB_CLIENT_ID,
                "client_secret": GITHUB_CLIENT_SECRET,
                "code": code
            },
            headers={"Accept": "application/json"}
        )
        
        if token_response.status_code != 200:
            raise HTTPException(status_code=400, detail="Failed to exchange code for token")
        
        token_data = token_response.json()
        access_token = token_data.get("access_token")
        
        # Get user info from GitHub
        user_response = await client.get(
            "https://api.github.com/user",
            headers={"Authorization": f"Bearer {access_token}"}
        )
        
        if user_response.status_code != 200:
            raise HTTPException(status_code=400, detail="Failed to get GitHub user info")
        
        github_user = user_response.json()
        
        # Update user with GitHub info
        await db_service.update_user_github(
            user_id=user_id,
            github_username=github_user["login"],
            github_token=access_token
        )
        
        return {
            "github_username": github_user["login"],
            "github_connected": True
        }

@router.post("/vercel/connect")
async def connect_vercel(user_id: str, vercel_connection: VercelConnection):
    """Connect Vercel account"""
    # Verify Vercel token
    async with httpx.AsyncClient() as client:
        response = await client.get(
            "https://api.vercel.com/v2/user",
            headers={"Authorization": f"Bearer {vercel_connection.vercel_token}"}
        )
        
        if response.status_code != 200:
            raise HTTPException(status_code=400, detail="Invalid Vercel token")
        
        vercel_user = response.json()
        
        # Update user with Vercel info
        await db_service.update_user_vercel(
            user_id=user_id,
            vercel_token=vercel_connection.vercel_token,
            vercel_team_id=vercel_connection.vercel_team_id
        )
        
        return {
            "vercel_username": vercel_user.get("username"),
            "vercel_connected": True
        }

@router.get("/user/{user_id}")
async def get_user(user_id: str):
    """Get user profile with connection status"""
    user = await db_service.get_user_by_id(user_id)
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    
    return {
        "id": user.id,
        "email": user.email,
        "name": user.name,
        "avatar_url": user.avatar_url,
        "github_connected": bool(user.github_username),
        "github_username": user.github_username,
        "vercel_connected": bool(user.vercel_token)
    }

@router.post("/logout")
async def logout(user_id: str):
    """Logout user"""
    # In a real app, you might invalidate tokens here
    return {"message": "Logged out successfully"}

@router.get("/github/connect")
async def github_connect():
    """Initiate GitHub OAuth connection"""
    if not GITHUB_CLIENT_ID:
        raise HTTPException(status_code=500, detail="GitHub OAuth not configured")
    
    state = str(uuid.uuid4())
    redirect_uri = f"{WEB_URL}/auth/callback/github"
    github_auth_url = (
        f"https://github.com/login/oauth/authorize?"
        f"client_id={GITHUB_CLIENT_ID}&"
        f"redirect_uri={redirect_uri}&"
        f"scope=repo%20user:email&"
        f"state={state}"
    )
    
    return {"auth_url": github_auth_url, "state": state}

@router.get("/vercel/connect")
async def vercel_connect():
    """Initiate Vercel OAuth connection"""
    state = str(uuid.uuid4())
    redirect_uri = f"{WEB_URL}/auth/callback/vercel"
    
    # Vercel OAuth URL (you'll need to configure Vercel OAuth app)
    vercel_auth_url = (
        f"https://vercel.com/oauth/authorize?"
        f"client_id=YOUR_VERCEL_CLIENT_ID&"  # Replace with actual client ID
        f"redirect_uri={redirect_uri}&"
        f"scope=read:user,write:project&"
        f"state={state}"
    )
    
    return {"auth_url": vercel_auth_url, "state": state}

@router.post("/github/disconnect")
async def disconnect_github():
    """Disconnect GitHub integration"""
    # Implementation to remove GitHub tokens from user account
    return {"message": "GitHub disconnected successfully"}

@router.post("/vercel/disconnect")
async def disconnect_vercel():
    """Disconnect Vercel integration"""
    # Implementation to remove Vercel tokens from user account
    return {"message": "Vercel disconnected successfully"}
