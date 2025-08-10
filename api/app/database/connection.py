import duckdb
import os
from typing import Optional

class DatabaseConnection:
    _instance: Optional['DatabaseConnection'] = None
    _connection: Optional[duckdb.DuckDBPyConnection] = None
    
    def __new__(cls):
        if cls._instance is None:
            cls._instance = super(DatabaseConnection, cls).__new__(cls)
        return cls._instance
    
    def __init__(self):
        if self._connection is None:
            db_path = os.getenv("DATABASE_PATH", "./data/database.db")
            os.makedirs(os.path.dirname(db_path), exist_ok=True)
            self._connection = duckdb.connect(db_path)
            self._init_tables()
    
    def get_connection(self) -> duckdb.DuckDBPyConnection:
        return self._connection
    
    def _init_tables(self):
        """Initialize all database tables"""
        tables = [
            """CREATE TABLE IF NOT EXISTS projects (
                id INTEGER PRIMARY KEY,
                name TEXT UNIQUE NOT NULL,
                template TEXT NOT NULL,
                docker_container TEXT UNIQUE,
                port INTEGER,
                status TEXT DEFAULT 'created',
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )""",
            """CREATE SEQUENCE IF NOT EXISTS projects_seq START 1""",
            """CREATE TABLE IF NOT EXISTS containers (
                id INTEGER PRIMARY KEY,
                name TEXT UNIQUE NOT NULL,
                project_id INTEGER,
                status TEXT DEFAULT 'created',
                port_mapping TEXT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (project_id) REFERENCES projects (id)
            )""",
            """CREATE SEQUENCE IF NOT EXISTS containers_seq START 1""",
            """CREATE TABLE IF NOT EXISTS token_usage (
                id INTEGER PRIMARY KEY,
                session_id TEXT NOT NULL,
                project_id INTEGER,
                model TEXT NOT NULL,
                provider TEXT NOT NULL,
                input_tokens INTEGER DEFAULT 0,
                output_tokens INTEGER DEFAULT 0,
                total_tokens INTEGER DEFAULT 0,
                request_type TEXT DEFAULT 'chat',
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (project_id) REFERENCES projects (id)
            )""",
            """CREATE SEQUENCE IF NOT EXISTS token_usage_seq START 1""",
            """CREATE TABLE IF NOT EXISTS conversation_messages (
                id INTEGER PRIMARY KEY,
                session_id TEXT NOT NULL,
                project_id INTEGER,
                role TEXT NOT NULL,
                content TEXT NOT NULL,
                model TEXT,
                provider TEXT,
                token_usage_id INTEGER,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (project_id) REFERENCES projects (id),
                FOREIGN KEY (token_usage_id) REFERENCES token_usage (id)
            )""",
            """CREATE SEQUENCE IF NOT EXISTS conversation_messages_seq START 1"""
        ]
        
        indexes = [
            "CREATE INDEX IF NOT EXISTS idx_token_usage_session ON token_usage(session_id)",
            "CREATE INDEX IF NOT EXISTS idx_token_usage_project ON token_usage(project_id)",
            "CREATE INDEX IF NOT EXISTS idx_conversation_session ON conversation_messages(session_id)",
            "CREATE INDEX IF NOT EXISTS idx_conversation_project ON conversation_messages(project_id)"
        ]
        
        for table_sql in tables:
            self._connection.execute(table_sql)
            
        for index_sql in indexes:
            self._connection.execute(index_sql)
        
        self._connection.commit()

# Global database instance
db = DatabaseConnection()
