from fastapi import APIRouter, HTTPException
from fastapi.responses import JSONResponse
from app.database.service import db_service

router = APIRouter()

@router.get("/usage/{session_id}")
def get_session_usage(session_id: str):
    """Get token usage for a specific session"""
    try:
        usage_records = db_service.get_session_token_usage(session_id)
        
        if not usage_records:
            return JSONResponse(content={
                "session_id": session_id,
                "total_tokens": 0,
                "input_tokens": 0,
                "output_tokens": 0,
                "records": []
            })
        
        total_tokens = sum(record.total_tokens for record in usage_records)
        total_input = sum(record.input_tokens for record in usage_records)
        total_output = sum(record.output_tokens for record in usage_records)
        
        return JSONResponse(content={
            "session_id": session_id,
            "total_tokens": total_tokens,
            "input_tokens": total_input,
            "output_tokens": total_output,
            "records": [
                {
                    "id": record.id,
                    "model": record.model,
                    "provider": record.provider,
                    "input_tokens": record.input_tokens,
                    "output_tokens": record.output_tokens,
                    "total_tokens": record.total_tokens,
                    "created_at": record.created_at.isoformat() if record.created_at else None
                }
                for record in usage_records
            ]
        })
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error fetching session usage: {str(e)}")

@router.get("/stats")
def get_global_stats():
    """Get global token usage statistics"""
    try:
        stats = db_service.get_global_token_stats()
        
        return JSONResponse(content={
            "total_tokens": stats.get("total_tokens", 0),
            "total_input_tokens": stats.get("total_input_tokens", 0),
            "total_output_tokens": stats.get("total_output_tokens", 0),
            "total_sessions": stats.get("total_sessions", 0),
            "models_used": stats.get("models_used", []),
            "providers_used": stats.get("providers_used", []),
            "last_updated": stats.get("last_updated")
        })
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error fetching global stats: {str(e)}")

@router.get("/project/{project_id}")
def get_project_usage(project_id: str):
    """Get token usage for a specific project"""
    try:
        usage_records = db_service.get_project_token_usage(project_id)
        
        if not usage_records:
            return JSONResponse(content={
                "project_id": project_id,
                "total_tokens": 0,
                "input_tokens": 0,
                "output_tokens": 0,
                "records": []
            })
        
        total_tokens = sum(record.total_tokens for record in usage_records)
        total_input = sum(record.input_tokens for record in usage_records)
        total_output = sum(record.output_tokens for record in usage_records)
        
        return JSONResponse(content={
            "project_id": project_id,
            "total_tokens": total_tokens,
            "input_tokens": total_input,
            "output_tokens": total_output,
            "records": [
                {
                    "id": record.id,
                    "session_id": record.session_id,
                    "model": record.model,
                    "provider": record.provider,
                    "input_tokens": record.input_tokens,
                    "output_tokens": record.output_tokens,
                    "total_tokens": record.total_tokens,
                    "created_at": record.created_at.isoformat() if record.created_at else None
                }
                for record in usage_records
            ]
        })
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error fetching project usage: {str(e)}")