import os
from fastapi import APIRouter, HTTPException
from fastapi.responses import JSONResponse

router = APIRouter()

@router.get("/all")
def get_all_models():
    """Get all available models and current provider"""
    provider = os.getenv("LLM_PROVIDER", "openrouter")
    
    # Define available models based on provider
    models_by_provider = {
        "openrouter": [
            "anthropic/claude-3.5-sonnet",
            "anthropic/claude-3-haiku",
            "openai/gpt-4o",
            "openai/gpt-4o-mini",
            "google/gemini-pro-1.5",
            "meta-llama/llama-3.1-8b-instruct",
            "mistralai/mistral-7b-instruct"
        ],
        "openai": [
            "gpt-4",
            "gpt-4-turbo",
            "gpt-3.5-turbo"
        ],
        "anthropic": [
            "claude-3-5-sonnet-20241022",
            "claude-3-haiku-20240307"
        ],
        "google": [
            "gemini-pro",
            "gemini-pro-vision"
        ]
    }
    
    available_models = models_by_provider.get(provider, models_by_provider["openrouter"])
    
    return JSONResponse(content={
        "provider": provider,
        "models": available_models,
        "current_model": os.getenv("MODEL_NAME", "anthropic/claude-3.5-sonnet")
    })

@router.get("")
def get_models():
    """Get current provider info - legacy endpoint"""
    return {
        "provider": os.getenv("LLM_PROVIDER", "openrouter"),
        "current_model": os.getenv("MODEL_NAME", "anthropic/claude-3.5-sonnet")
    }