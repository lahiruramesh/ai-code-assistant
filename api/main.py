import os
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from app.api import streaming, projects
from app.database.connection import db
from app.database.service import db_service
from dotenv import load_dotenv

load_dotenv()

# Create the projects directory if it doesn't exist
if not os.path.exists("./projects"):
    os.makedirs("./projects")

# Create the data directory for database if it doesn't exist
if not os.path.exists("./data"):
    os.makedirs("./data")

app = FastAPI(
    title="LangChain ReAct Agent Backend with DuckDB",
    description="A streaming backend for a LangChain agent with tool-calling capabilities and persistent storage.",
    version="0.2.0",
)

# Configure CORS to allow the frontend to connect
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:8080", "http://localhost:3000", "http://localhost:5173"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include API routers
app.include_router(streaming.router, prefix="/api/v1/chat", tags=["Chat"])
app.include_router(projects.router, prefix="/api/v1/projects", tags=["Projects"])

@app.get("/api/v1/models")
def get_models():
    """Get available models and current provider"""
    return {
        "provider": os.getenv("LLM_PROVIDER", "openai"),
        "models": [
            "gpt-4",
            "gpt-4-turbo",
            "gpt-3.5-turbo"
        ]
    }

@app.get("/api/v1/chat/{chat_id}")
def get_chat_history(chat_id: str):
    """Get chat history by chat ID (session ID)"""
    try:
        messages = db_service.get_conversation_messages(chat_id)
        return {
            "chat_id": chat_id,
            "messages": [
                {
                    "id": msg.id,
                    "type": msg.role,
                    "content": msg.content,
                    "timestamp": msg.created_at.isoformat() if msg.created_at else None,
                    "model": msg.model,
                    "provider": msg.provider
                }
                for msg in messages
            ]
        }
    except Exception as e:
        raise HTTPException(status_code=404, detail=f"Chat not found: {str(e)}")

@app.post("/api/v1/chat/{session_id}/cancel")
def cancel_chat_session(session_id: str):
    """Cancel an active chat session"""
    # TODO: Implement session cancellation logic
    # For now, just return success
    return {"message": "Session cancelled", "session_id": session_id}

@app.get("/")
def read_root():
    return {
        "message": "Welcome to the LangChain ReAct Agent Backend",
        "version": "0.2.0",
        "features": [
            "DuckDB Integration",
            "Project-aware Chat Sessions",
            "WebSocket Streaming",
            "Token Usage Tracking",
            "Conversation History"
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

@app.on_event("startup")
async def startup_event():
    """Initialize database on startup"""
    print("🚀 Starting API server...")
    print("📁 Projects directory:", os.path.abspath("./projects"))
    print("🗄️  Database path:", os.path.abspath("./data/database.db"))
    print("✅ Server ready!")

@app.on_event("shutdown")
async def shutdown_event():
    """Cleanup on shutdown"""
    print("🛑 Shutting down server...")
    if hasattr(db, '_connection') and db._connection:
        db._connection.close()
    print("✅ Cleanup complete!")