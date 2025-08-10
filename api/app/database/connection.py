import duckdb
import os
from typing import Optional
from ..config import DATABASE_DIR, DATABASE_FILE, RESET_DB_ON_STARTUP

class DatabaseConnection:
    _instance: Optional['DatabaseConnection'] = None
    _connection: Optional[duckdb.DuckDBPyConnection] = None
    
    def __new__(cls):
        if cls._instance is None:
            cls._instance = super(DatabaseConnection, cls).__new__(cls)
        return cls._instance
    
    def __init__(self):
        if self._connection is None:
            # Ensure directory exists
            os.makedirs(DATABASE_DIR, exist_ok=True)

            # Connect to database with a small recovery routine for WAL issues
            self._connection = self._connect_with_recovery()

            # Initialize schema (optionally reset)
            self._init_tables(reset=RESET_DB_ON_STARTUP)
    
    def get_connection(self) -> duckdb.DuckDBPyConnection:
        return self._connection
    
    def reconnect(self) -> duckdb.DuckDBPyConnection:
        """Force reconnection to database"""
        try:
            if self._connection:
                self._connection.close()
        except:
            pass
        self._connection = self._connect_with_recovery()
        return self._connection
    
    def _init_tables(self, reset: bool = False):
        """Initialize all database tables.
        If reset=True, drop-and-recreate tables; otherwise, create if not exists.
        """

        if reset:
            drop_tables = [
                "DROP TABLE IF EXISTS conversation_messages",
                "DROP TABLE IF EXISTS token_usage", 
                "DROP TABLE IF EXISTS containers",
                "DROP TABLE IF EXISTS projects"
            ]
            for drop_sql in drop_tables:
                try:
                    self._connection.execute(drop_sql)
                except Exception:
                    pass  # Ignore errors if tables don't exist
        
        tables = [
            """CREATE TABLE IF NOT EXISTS projects (
                id TEXT PRIMARY KEY,
                name TEXT UNIQUE NOT NULL,
                template TEXT NOT NULL,
                docker_container TEXT UNIQUE,
                port INTEGER,
                status TEXT DEFAULT 'created',
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )""",
            """CREATE TABLE IF NOT EXISTS containers (
                id TEXT PRIMARY KEY,
                name TEXT UNIQUE NOT NULL,
                project_id TEXT,
                status TEXT DEFAULT 'created',
                port_mapping TEXT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (project_id) REFERENCES projects (id)
            )""",
            """CREATE TABLE IF NOT EXISTS token_usage (
                id TEXT PRIMARY KEY,
                session_id TEXT NOT NULL,
                project_id TEXT,
                model TEXT NOT NULL,
                provider TEXT NOT NULL,
                input_tokens INTEGER DEFAULT 0,
                output_tokens INTEGER DEFAULT 0,
                total_tokens INTEGER DEFAULT 0,
                request_type TEXT DEFAULT 'chat',
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (project_id) REFERENCES projects (id)
            )""",
            """CREATE TABLE IF NOT EXISTS conversation_messages (
                id TEXT PRIMARY KEY,
                session_id TEXT,
                project_id TEXT NOT NULL,
                role TEXT NOT NULL,
                content TEXT NOT NULL,
                message_type TEXT DEFAULT 'chat',
                model TEXT,
                provider TEXT,
                token_usage_id TEXT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (project_id) REFERENCES projects (id),
                FOREIGN KEY (token_usage_id) REFERENCES token_usage (id)
            )""",

        ]
        
        indexes = [
            "CREATE INDEX IF NOT EXISTS idx_token_usage_session ON token_usage(session_id)",
            "CREATE INDEX IF NOT EXISTS idx_token_usage_project ON token_usage(project_id)",
            "CREATE INDEX IF NOT EXISTS idx_conversation_project ON conversation_messages(project_id)",
            "CREATE INDEX IF NOT EXISTS idx_conversation_created ON conversation_messages(created_at)",
            "CREATE INDEX IF NOT EXISTS idx_projects_created ON projects(created_at)"
        ]
        
        for table_sql in tables:
            self._connection.execute(table_sql)
            
        for index_sql in indexes:
            self._connection.execute(index_sql)
        
        self._connection.commit()

    def _connect_with_recovery(self) -> duckdb.DuckDBPyConnection:
        """Attempt to connect to the DuckDB database with minimal recovery steps.
        - If a WAL-related error occurs, try removing a stale .wal file and reconnect.
        - If still failing and RESET_DB_ON_STARTUP is True, delete the DB file and recreate.
        """
        try:
            return duckdb.connect(DATABASE_FILE)
        except Exception as e:
            msg = str(e).lower()
            wal_path = f"{DATABASE_FILE}.wal"
            # Try WAL cleanup if error mentions wal/catalog issues
            if "wal" in msg or ("catalog" in msg and os.path.exists(wal_path)):
                try:
                    if os.path.exists(wal_path):
                        os.remove(wal_path)
                    return duckdb.connect(DATABASE_FILE)
                except Exception:
                    pass
            # Last resort: if flag set, delete and recreate db file
            if RESET_DB_ON_STARTUP:
                try:
                    if os.path.exists(DATABASE_FILE):
                        os.remove(DATABASE_FILE)
                    if os.path.exists(wal_path):
                        os.remove(wal_path)
                except Exception:
                    pass
                return duckdb.connect(DATABASE_FILE)
            # Re-raise if we cannot or should not recover
            raise

# Global database instance
db = DatabaseConnection()
