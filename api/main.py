import os
from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from app.api import streaming, projects, auth, github, vercel, models, tokens
from app.database.connection import db
from app.database.service import db_service
from app.config import (
    WEB_URL
)

# Create the projects directory if it doesn't exist
if not os.path.exists("./projects"):
    os.makedirs("./projects")

# Create the data directory for database if it doesn't exist
if not os.path.exists("./data"):
    os.makedirs("./data")

@asynccontextmanager
async def lifespan(app: FastAPI):
    """Handle application lifespan events"""
    # Startup
    print("🚀 Starting API server...")
    print("✅ Server ready!")
    
    yield
    
    # Shutdown
    print("🛑 Shutting down server...")
    if hasattr(db, '_connection') and db._connection:
        db._connection.close()
    print("✅ Cleanup complete!")

app = FastAPI(
    title="Code Editing Agent Backend with Authentication & Integrations",
    description="A streaming backend for a LangChain agent with authentication, GitHub, and Vercel integrations.",
    version="0.3.0",
    lifespan=lifespan,
)

# Configure CORS to allow the frontend to connect
app.add_middleware(
    CORSMiddleware,
    allow_origins=[
        "http://localhost:8080", 
        WEB_URL
    ],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include API routers
app.include_router(streaming.router, prefix="/api/v1/chat", tags=["Chat"])
app.include_router(projects.router, prefix="/api/v1/projects", tags=["Projects"])
app.include_router(auth.router, prefix="/api/v1", tags=["Authentication"])
app.include_router(github.router, prefix="/api/v1", tags=["GitHub Integration"])
app.include_router(vercel.router, prefix="/api/v1", tags=["Vercel Integration"])
app.include_router(models.router, prefix="/api/v1/models", tags=["Models"])
app.include_router(tokens.router, prefix="/api/v1/tokens", tags=["Tokens"])



# @app.get("/api/v1/chat/{chat_id}")
# def get_chat_history(chat_id: str):
#     """Get chat history by chat ID (session ID)"""
#     try:
#         messages = db_service.get_conversation_messages(chat_id)
#         return {
#             "chat_id": chat_id,
#             "messages": [
#                 {
#                     "id": msg.id,
#                     "type": msg.role,
#                     "content": msg.content,
#                     "timestamp": msg.created_at.isoformat() if msg.created_at else None,
#                     "model": msg.model,
#                     "provider": msg.provider
#                 }
#                 for msg in messages
#             ]
#         }
#     except Exception as e:
#         raise HTTPException(status_code=404, detail=f"Chat not found: {str(e)}")

@app.post("/api/v1/chat/{session_id}/cancel")
def cancel_chat_session(session_id: str):
    """Cancel an active chat session"""
    # TODO: Implement session cancellation logic
    # For now, just return success
    return {"message": "Session cancelled", "session_id": session_id}

@app.get("/")
def read_root():
    return {
        "message": "Welcome to the Code Editing Agent Backend",
        "version": "0.3.0",
        "features": [
            "DuckDB Integration",
            "Project-aware Chat Sessions", 
            "WebSocket Streaming",
            "Token Usage Tracking",
            "Conversation History",
            "Google OAuth Authentication",
            "GitHub Integration",
            "Vercel Deployment",
            "Repository Management"
        ]
    }

@app.get("/health")
def health_check():
    """Health check endpoint"""
    try:
        # Test database connection
        conn = db.get_connection()
        conn.execute("SELECT 1").fetchone()
        return {"status": "healthy", "database": "connected"}
    except Exception as e:
        return {"status": "unhealthy", "error": str(e)}