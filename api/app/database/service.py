from typing import List, Optional
from datetime import datetime
import uuid
import random
import re
import duckdb
from app.database.connection import db
from app.database.models import (
    Project, ProjectCreate, ConversationMessage, ConversationMessageCreate,
    TokenUsage, TokenUsageCreate, User, UserCreate, GitHubRepository, 
    VercelDeploymentRecord
)

class DatabaseService:
    def __init__(self):
        self.conn = db.get_connection()
        self.create_tables()
    
    def _execute_with_retry(self, query: str, params: list = None, max_retries: int = 3):
        """Execute a query with automatic retry on database invalidation"""
        for attempt in range(max_retries):
            try:
                if params:
                    return self.conn.execute(query, params)
                else:
                    return self.conn.execute(query)
            except duckdb.FatalException as e:
                if "database has been invalidated" in str(e) and attempt < max_retries - 1:
                    print(f"Database invalidated, reconnecting (attempt {attempt + 1})")
                    # Reconnect to database
                    self.conn = db.reconnect()
                    continue
                raise
            except Exception as e:
                if attempt < max_retries - 1:
                    print(f"Database error, retrying (attempt {attempt + 1}): {e}")
                    try:
                        self.conn = db.reconnect()
                        continue
                    except:
                        pass
                raise
        
    def _fetchone_with_retry(self, query: str, params: list = None):
        """Execute query and fetch one result with retry logic"""
        result = self._execute_with_retry(query, params)
        return result.fetchone()
    
    def _fetchall_with_retry(self, query: str, params: list = None):
        """Execute query and fetch all results with retry logic"""
        result = self._execute_with_retry(query, params)
        return result.fetchall()
    
    def create_tables(self):
        """Create all necessary tables"""
        cursor = self.conn.cursor()
        
        # Users table
        cursor.execute("""
            CREATE TABLE IF NOT EXISTS users (
                id TEXT PRIMARY KEY,
                email TEXT UNIQUE NOT NULL,
                name TEXT NOT NULL,
                avatar_url TEXT,
                google_id TEXT,
                github_username TEXT,
                github_token TEXT,
                vercel_token TEXT,
                vercel_team_id TEXT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        """)
        
        # GitHub repositories table
        cursor.execute("""
            CREATE TABLE IF NOT EXISTS github_repositories (
                id TEXT PRIMARY KEY,
                user_id TEXT NOT NULL,
                project_id TEXT,
                repo_name TEXT NOT NULL,
                repo_url TEXT NOT NULL,
                clone_url TEXT NOT NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (user_id) REFERENCES users (id),
                FOREIGN KEY (project_id) REFERENCES projects (id)
            )
        """)
        
        # Vercel deployments table
        cursor.execute("""
            CREATE TABLE IF NOT EXISTS vercel_deployments (
                id TEXT PRIMARY KEY,
                user_id TEXT NOT NULL,
                project_id TEXT NOT NULL,
                deployment_id TEXT NOT NULL,
                deployment_url TEXT NOT NULL,
                status TEXT DEFAULT 'QUEUED',
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (user_id) REFERENCES users (id),
                FOREIGN KEY (project_id) REFERENCES projects (id)
            )
        """)
        
        # Check if projects table needs updating
        try:
            # Try to add new columns to existing projects table
            cursor.execute("ALTER TABLE projects ADD COLUMN user_id TEXT")
        except:
            pass  # Column might already exist
            
        try:
            cursor.execute("ALTER TABLE projects ADD COLUMN github_repo_id TEXT")
        except:
            pass
            
        try:
            cursor.execute("ALTER TABLE projects ADD COLUMN vercel_deployment_id TEXT")
        except:
            pass
        
        self.conn.commit()
    
    # User operations
    async def create_user(self, user_data: UserCreate) -> User:
        user_id = str(uuid.uuid4())
        
        query = """
        INSERT INTO users (id, email, name, avatar_url, google_id, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        """
        self.conn.execute(
            query, 
            [user_id, user_data.email, user_data.name, user_data.avatar_url, user_data.google_id]
        )
        self.conn.commit()
        
        return await self.get_user_by_id(user_id)
    
    async def get_user_by_id(self, user_id: str) -> Optional[User]:
        query = "SELECT * FROM users WHERE id = ?"
        result = self.conn.execute(query, [user_id]).fetchone()
        if result:
            return User(
                id=result[0], email=result[1], name=result[2], avatar_url=result[3],
                google_id=result[4], github_username=result[5], github_token=result[6],
                vercel_token=result[7], vercel_team_id=result[8], 
                created_at=result[9], updated_at=result[10]
            )
        return None
    
    async def get_user_by_email(self, email: str) -> Optional[User]:
        query = "SELECT * FROM users WHERE email = ?"
        result = self.conn.execute(query, [email]).fetchone()
        if result:
            return User(
                id=result[0], email=result[1], name=result[2], avatar_url=result[3],
                google_id=result[4], github_username=result[5], github_token=result[6],
                vercel_token=result[7], vercel_team_id=result[8],
                created_at=result[9], updated_at=result[10]
            )
        return None
    
    async def update_user_github(self, user_id: str, github_username: str, github_token: str):
        query = """
        UPDATE users 
        SET github_username = ?, github_token = ?, updated_at = CURRENT_TIMESTAMP 
        WHERE id = ?
        """
        self.conn.execute(query, [github_username, github_token, user_id])
        self.conn.commit()
    
    async def update_user_vercel(self, user_id: str, vercel_token: str, vercel_team_id: Optional[str] = None):
        query = """
        UPDATE users 
        SET vercel_token = ?, vercel_team_id = ?, updated_at = CURRENT_TIMESTAMP 
        WHERE id = ?
        """
        self.conn.execute(query, [vercel_token, vercel_team_id, user_id])
        self.conn.commit()
    
    # GitHub repository operations
    async def create_github_repository(self, repo: GitHubRepository) -> GitHubRepository:
        query = """
        INSERT INTO github_repositories (id, user_id, project_id, repo_name, repo_url, clone_url, created_at)
        VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
        """
        self.conn.execute(
            query,
            [repo.id, repo.user_id, repo.project_id, repo.repo_name, repo.repo_url, repo.clone_url]
        )
        self.conn.commit()
        return repo
    
    async def get_github_repository_by_name(self, user_id: str, repo_name: str) -> Optional[GitHubRepository]:
        query = "SELECT * FROM github_repositories WHERE user_id = ? AND repo_name = ?"
        result = self.conn.execute(query, [user_id, repo_name]).fetchone()
        if result:
            return GitHubRepository(
                id=result[0], user_id=result[1], project_id=result[2],
                repo_name=result[3], repo_url=result[4], clone_url=result[5],
                created_at=result[6]
            )
        return None
    
    async def get_github_repository_by_project(self, project_id: str) -> Optional[GitHubRepository]:
        query = "SELECT * FROM github_repositories WHERE project_id = ?"
        result = self.conn.execute(query, [project_id]).fetchone()
        if result:
            return GitHubRepository(
                id=result[0], user_id=result[1], project_id=result[2],
                repo_name=result[3], repo_url=result[4], clone_url=result[5],
                created_at=result[6]
            )
        return None
    
    async def update_github_repository_project(self, repo_id: str, project_id: str):
        query = "UPDATE github_repositories SET project_id = ? WHERE id = ?"
        self.conn.execute(query, [project_id, repo_id])
        self.conn.commit()
    
    async def delete_github_repository_by_name(self, user_id: str, repo_name: str):
        query = "DELETE FROM github_repositories WHERE user_id = ? AND repo_name = ?"
        self.conn.execute(query, [user_id, repo_name])
        self.conn.commit()
    
    # Vercel deployment operations
    async def create_vercel_deployment(self, deployment: VercelDeploymentRecord) -> VercelDeploymentRecord:
        query = """
        INSERT INTO vercel_deployments (id, user_id, project_id, deployment_id, deployment_url, status, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        """
        self.conn.execute(
            query,
            [deployment.id, deployment.user_id, deployment.project_id, 
             deployment.deployment_id, deployment.deployment_url, deployment.status]
        )
        self.conn.commit()
        return deployment
    
    async def get_vercel_deployment_by_deployment_id(self, deployment_id: str) -> Optional[VercelDeploymentRecord]:
        query = "SELECT * FROM vercel_deployments WHERE deployment_id = ?"
        result = self.conn.execute(query, [deployment_id]).fetchone()
        if result:
            return VercelDeploymentRecord(
                id=result[0], user_id=result[1], project_id=result[2],
                deployment_id=result[3], deployment_url=result[4], status=result[5],
                created_at=result[6], updated_at=result[7]
            )
        return None
    
    async def update_vercel_deployment_status(self, deployment_id: str, status: str):
        query = """
        UPDATE vercel_deployments 
        SET status = ?, updated_at = CURRENT_TIMESTAMP 
        WHERE deployment_id = ?
        """
        self.conn.execute(query, [status, deployment_id])
        self.conn.commit()
    
    async def delete_vercel_deployment_by_deployment_id(self, deployment_id: str):
        query = "DELETE FROM vercel_deployments WHERE deployment_id = ?"
        self.conn.execute(query, [deployment_id])
        self.conn.commit()
    
    # Update project operations to include user and integration relations
    async def update_project_github_repo(self, project_id: str, github_repo_id: str):
        query = """
        UPDATE projects 
        SET github_repo_id = ?, updated_at = CURRENT_TIMESTAMP 
        WHERE id = ?
        """
        self.conn.execute(query, [github_repo_id, project_id])
        self.conn.commit()
    
    async def update_project_vercel_deployment(self, project_id: str, vercel_deployment_id: str):
        query = """
        UPDATE projects 
        SET vercel_deployment_id = ?, updated_at = CURRENT_TIMESTAMP 
        WHERE id = ?
        """
        self.conn.execute(query, [vercel_deployment_id, project_id])
        self.conn.commit()
    
    # Project operations
    def create_project(self, project_data: ProjectCreate) -> Project:
        import uuid
        project_id = str(uuid.uuid4())
        
        query = """
        INSERT INTO projects (id, name, template, docker_container, port, status, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, 'created', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        RETURNING *
        """
        result = self._fetchone_with_retry(
            query, 
            [project_id, project_data.name, project_data.template, project_data.docker_container, project_data.port]
        )
        self.conn.commit()
        
        return Project(
            id=result[0],
            name=result[1],
            template=result[2],
            docker_container=result[3],
            port=result[4],
            status=result[5],
            created_at=result[6],
            updated_at=result[7]
        )
    
    # Update the project data
    def update_project(self, project_id: str, project_data: ProjectCreate) -> Project:
        query = """
        UPDATE projects 
        SET name = ?, template = ?, docker_container = ?, port = ?, updated_at = CURRENT_TIMESTAMP 
        WHERE id = ?
        RETURNING *
        """
        result = self._fetchone_with_retry(
            query, 
            [project_data.name, project_data.template, project_data.docker_container, project_data.port, project_id]
        )
        self.conn.commit()
        
        return Project(
            id=result[0],
            name=result[1],
            template=result[2],
            docker_container=result[3],
            port=result[4],
            status=result[5],
            created_at=result[6],
            updated_at=result[7]
        )
    
    
    def get_project_by_id(self, project_id: str) -> Optional[Project]:
        query = "SELECT * FROM projects WHERE id = ?"
        result = self._fetchone_with_retry(query, [project_id])
        if result:
            return Project(
                id=result[0],
                name=result[1],
                template=result[2],
                docker_container=result[3],
                port=result[4],
                status=result[5],
                created_at=result[6],
                updated_at=result[7]
            )
        return None
    
    def get_project_by_name(self, name: str) -> Optional[Project]:
        query = "SELECT * FROM projects WHERE name = ?"
        result = self._fetchone_with_retry(query, [name])
        if result:
            return Project(
                id=result[0],
                name=result[1],
                template=result[2],
                docker_container=result[3],
                port=result[4],
                status=result[5],
                created_at=result[6],
                updated_at=result[7]
            )
        return None
    
    def get_all_projects(self) -> List[Project]:
        query = "SELECT * FROM projects ORDER BY created_at DESC"
        results = self._fetchall_with_retry(query)
        return [
            Project(
                id=row[0],
                name=row[1],
                template=row[2],
                docker_container=row[3],
                port=row[4],
                status=row[5],
                created_at=row[6],
                updated_at=row[7]
            )
            for row in results
        ]
    
    def delete_project(self, project_id: str) -> bool:
        """Delete a project and all associated data"""
        try:
            # Delete associated conversation messages first (foreign key constraint)
            delete_messages_query = "DELETE FROM conversation_messages WHERE project_id = ?"
            self._execute_with_retry(delete_messages_query, [project_id])
            
            # Delete associated token usage records
            delete_tokens_query = "DELETE FROM token_usage WHERE project_id = ?"
            self._execute_with_retry(delete_tokens_query, [project_id])
            
            # Delete the project
            delete_project_query = "DELETE FROM projects WHERE id = ?"
            result = self._execute_with_retry(delete_project_query, [project_id])
            
            self.conn.commit()
            return True
        except Exception as e:
            print(f"Error deleting project {project_id}: {e}")
            raise
    
    # Conversation operations
    def create_conversation_message(self, message_data: ConversationMessageCreate) -> ConversationMessage:
        import uuid
        message_id = str(uuid.uuid4())
        
        query = """
        INSERT INTO conversation_messages (id, project_id, role, content, message_type, model, provider, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        RETURNING *
        """
        result = self._fetchone_with_retry(
            query,
            [
                message_id, message_data.project_id, message_data.role, message_data.content,
                message_data.message_type, message_data.model, message_data.provider
            ]
        )
        self.conn.commit()
        
        return ConversationMessage(
            id=result[0],
            project_id=result[1],
            role=result[2],
            content=result[3],
            message_type=result[4],
            model=result[5],
            provider=result[6],
            token_usage_id=result[7],
            created_at=result[8],
            updated_at=result[9]
        )
    
    def get_project_messages(self, project_id: str) -> List[ConversationMessage]:
        query = """
        SELECT id, session_id, project_id, role, content, message_type, model, provider, token_usage_id, created_at, updated_at 
        FROM conversation_messages 
        WHERE project_id = ? AND message_type = 'chat'
        ORDER BY created_at ASC
        """
        results = self._fetchall_with_retry(query, [project_id])
        return [
            ConversationMessage(
                id=row[0],
                project_id=row[2],  # project_id is at index 2
                role=row[3],        # role is at index 3
                content=row[4],     # content is at index 4
                message_type=row[5], # message_type is at index 5
                model=row[6],       # model is at index 6
                provider=row[7],    # provider is at index 7
                token_usage_id=row[8], # token_usage_id is at index 8
                created_at=row[9],  # created_at is at index 9
                updated_at=row[10]  # updated_at is at index 10
            )
            for row in results
        ]
    
    def get_conversation_messages(self, session_id: str) -> List[ConversationMessage]:
        """Legacy method - kept for backward compatibility"""
        query = """
        SELECT * FROM conversation_messages 
        WHERE session_id = ? 
        ORDER BY created_at ASC
        """
        results = self.conn.execute(query, [session_id]).fetchall()
        return [
            ConversationMessage(
                id=row[0],
                project_id=row[1],
                role=row[2],
                content=row[3],
                message_type=row[4],
                model=row[5],
                provider=row[6],
                token_usage_id=row[7],
                created_at=row[8],
                updated_at=row[9]
            )
            for row in results
        ]
    
    # Token usage operations
    def create_token_usage(self, usage_data: TokenUsageCreate) -> TokenUsage:
        import uuid
        usage_id = str(uuid.uuid4())
        
        query = """
        INSERT INTO token_usage (id, session_id, project_id, model, provider, input_tokens, output_tokens, total_tokens, request_type, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
        RETURNING *
        """
        result = self.conn.execute(
            query,
            [
                usage_id, usage_data.session_id, usage_data.project_id, usage_data.model,
                usage_data.provider, usage_data.input_tokens, usage_data.output_tokens,
                usage_data.total_tokens, usage_data.request_type
            ]
        ).fetchone()
        self.conn.commit()
        
        return TokenUsage(
            id=result[0],
            session_id=result[1],
            project_id=result[2],
            model=result[3],
            provider=result[4],
            input_tokens=result[5],
            output_tokens=result[6],
            total_tokens=result[7],
            request_type=result[8],
            created_at=result[9]
        )
    
    def get_token_usage_by_session(self, session_id: str) -> List[TokenUsage]:
        query = """
        SELECT * FROM token_usage 
        WHERE session_id = ? 
        ORDER BY created_at DESC
        """
        results = self.conn.execute(query, [session_id]).fetchall()
        return [
            TokenUsage(
                id=row[0],
                session_id=row[1],
                project_id=row[2],
                model=row[3],
                provider=row[4],
                input_tokens=row[5],
                output_tokens=row[6],
                total_tokens=row[7],
                request_type=row[8],
                created_at=row[9]
            )
            for row in results
        ]
    
    def get_session_token_usage(self, session_id: str) -> List[TokenUsage]:
        """Get token usage records for a specific session"""
        query = """
        SELECT * FROM token_usage 
        WHERE session_id = ? 
        ORDER BY created_at DESC
        """
        results = self._fetchall_with_retry(query, [session_id])
        return [
            TokenUsage(
                id=row[0],
                session_id=row[1],
                project_id=row[2],
                model=row[3],
                provider=row[4],
                input_tokens=row[5],
                output_tokens=row[6],
                total_tokens=row[7],
                request_type=row[8],
                created_at=row[9]
            )
            for row in results
        ]
    
    def get_project_token_usage(self, project_id: str) -> List[TokenUsage]:
        """Get token usage records for a specific project"""
        query = """
        SELECT * FROM token_usage 
        WHERE project_id = ? 
        ORDER BY created_at DESC
        """
        results = self._fetchall_with_retry(query, [project_id])
        return [
            TokenUsage(
                id=row[0],
                session_id=row[1],
                project_id=row[2],
                model=row[3],
                provider=row[4],
                input_tokens=row[5],
                output_tokens=row[6],
                total_tokens=row[7],
                request_type=row[8],
                created_at=row[9]
            )
            for row in results
        ]
    
    def get_global_token_stats(self) -> dict:
        """Get global token usage statistics"""
        try:
            # Get total token counts
            totals_query = """
            SELECT 
                COALESCE(SUM(total_tokens), 0) as total_tokens,
                COALESCE(SUM(input_tokens), 0) as total_input_tokens,
                COALESCE(SUM(output_tokens), 0) as total_output_tokens,
                COUNT(DISTINCT session_id) as total_sessions
            FROM token_usage
            """
            totals_result = self._fetchone_with_retry(totals_query)
            
            # Get unique models used
            models_query = "SELECT DISTINCT model FROM token_usage WHERE model IS NOT NULL"
            models_results = self._fetchall_with_retry(models_query)
            models_used = [row[0] for row in models_results]
            
            # Get unique providers used
            providers_query = "SELECT DISTINCT provider FROM token_usage WHERE provider IS NOT NULL"
            providers_results = self._fetchall_with_retry(providers_query)
            providers_used = [row[0] for row in providers_results]
            
            # Get last updated timestamp
            last_updated_query = "SELECT MAX(created_at) FROM token_usage"
            last_updated_result = self._fetchone_with_retry(last_updated_query)
            last_updated = last_updated_result[0] if last_updated_result and last_updated_result[0] else None
            
            return {
                "total_tokens": totals_result[0] if totals_result else 0,
                "total_input_tokens": totals_result[1] if totals_result else 0,
                "total_output_tokens": totals_result[2] if totals_result else 0,
                "total_sessions": totals_result[3] if totals_result else 0,
                "models_used": models_used,
                "providers_used": providers_used,
                "last_updated": last_updated.isoformat() if last_updated else None
            }
        except Exception as e:
            print(f"Error getting global token stats: {e}")
            return {
                "total_tokens": 0,
                "total_input_tokens": 0,
                "total_output_tokens": 0,
                "total_sessions": 0,
                "models_used": [],
                "providers_used": [],
                "last_updated": None
            }

    def get_chat_summary(self, project_id: str) -> str:
        """Generate a summary of the chat history for a project"""
        messages = self.get_project_messages(project_id)
        
        if len(messages) < 2:  # No meaningful conversation yet
            return ""
        
        # Create a concise summary of the conversation
        summary_parts = []
        user_messages = [msg for msg in messages if msg.role == "user"]
        assistant_messages = [msg for msg in messages if msg.role == "assistant"]
        
        if user_messages:
            summary_parts.append(f"User has made {len(user_messages)} requests")
            
        if assistant_messages:
            summary_parts.append(f"Assistant has provided {len(assistant_messages)} responses")
            
        # Get the last few exchanges for context
        recent_messages = messages[-6:]  # Last 6 messages (3 exchanges)
        if recent_messages:
            summary_parts.append("Recent conversation context:")
            for msg in recent_messages:
                role = "User" if msg.role == "user" else "Assistant"
                content_preview = msg.content[:100] + "..." if len(msg.content) > 100 else msg.content
                summary_parts.append(f"- {role}: {content_preview}")
        
        return "\n".join(summary_parts)

    def generate_fancy_project_name(self, query: str) -> str:
        """Generate a fancy project name based on the user query"""
        # Extract meaningful words from the query
        words = re.findall(r'\b\w+\b', query.lower())
        meaningful_words = [word for word in words if len(word) > 3 and word not in ['with', 'using', 'create', 'make', 'build', 'develop']]
        
        # Adjectives for fancy names
        adjectives = [
            'stellar', 'cosmic', 'quantum', 'nexus', 'prime', 'apex', 'zen', 'flux',
            'epic', 'vivid', 'swift', 'noble', 'crystal', 'golden', 'silver', 'phoenix'
        ]
        
        # Project type suffixes
        suffixes = ['hub', 'forge', 'studio', 'lab', 'works', 'craft', 'core', 'space']
        
        if meaningful_words:
            base_word = meaningful_words[0].capitalize()
        else:
            base_word = "Project"
        
        adjective = random.choice(adjectives).capitalize()
        suffix = random.choice(suffixes).capitalize()
        
        return f"{adjective}{base_word}{suffix}-{random.randint(10, 100)}"

# Global database service instance
db_service = DatabaseService()
