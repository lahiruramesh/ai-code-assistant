"""
Centralized configuration management for the application.
Loads environment variables from .env file and provides them as constants.

Notes on Database config:
- Prefer DATABASE_DIR (a folder). We'll create database.db inside it.
- Back-compat: if DATABASE_PATH is provided and ends with .db, we treat it as a file path.
	If it doesn't end with .db, we treat it as a directory.
"""
import os
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()

# API Configuration
OPENROUTER_API_KEY = os.getenv("OPENROUTER_API_KEY")
OPENROUTER_API_BASE = os.getenv("OPENROUTER_API_BASE", "https://openrouter.ai/api/v1")
MODEL_NAME = os.getenv("MODEL_NAME", "anthropic/claude-3.5-sonnet")

# Project Configuration
PROJECTS_DIR = os.getenv("PROJECTS_DIR", "/tmp/projects")
PROJECTS_TEMPLATE_DIR = os.getenv("PROJECTS_TEMPLATE_DIR", "/tmp/projects/templates")

# Docker Configuration
DOCK_ROUTE_PATH = os.getenv("DOCK_ROUTE_PATH", "/usr/local/bin/dock-route")

# Database Configuration
# Preferred: provide DATABASE_DIR (directory). We'll write database.db in there.
# Back-compat optional vars: DATABASE_PATH (dir or file) or DATABASE_FILE (explicit file path).
_DATABASE_DIR_ENV = os.getenv("DATABASE_DIR")
_DATABASE_PATH_ENV = os.getenv("DATABASE_PATH")
_DATABASE_FILE_ENV = os.getenv("DATABASE_FILE")

# Resolve database directory and file consistently
if _DATABASE_FILE_ENV:
	# Explicit file path
	DATABASE_FILE = _DATABASE_FILE_ENV
	DATABASE_DIR = os.path.dirname(DATABASE_FILE) or "."
else:
	if _DATABASE_DIR_ENV:
		DATABASE_DIR = _DATABASE_DIR_ENV
		DATABASE_FILE = os.path.join(DATABASE_DIR, "database.db")
	elif _DATABASE_PATH_ENV:
		# If ends with .db -> treat as file, else as directory
		if _DATABASE_PATH_ENV.lower().endswith(".db"):
			DATABASE_FILE = _DATABASE_PATH_ENV
			DATABASE_DIR = os.path.dirname(DATABASE_FILE) or "."
		else:
			DATABASE_DIR = _DATABASE_PATH_ENV
			DATABASE_FILE = os.path.join(DATABASE_DIR, "database.db")
	else:
		# Defaults
		DATABASE_DIR = os.getenv("DATABASE_DEFAULT_DIR", "./data")
		DATABASE_FILE = os.path.join(DATABASE_DIR, "database.db")

# Feature flags
RESET_DB_ON_STARTUP = os.getenv("RESET_DB_ON_STARTUP", "false").strip().lower() in ("1", "true", "yes", "on")