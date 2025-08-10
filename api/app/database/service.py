from typing import List, Optional
from datetime import datetime
import uuid
import random
import re
from app.database.connection import db
from app.database.models import (
    Project, ProjectCreate, ConversationMessage, ConversationMessageCreate,
    TokenUsage, TokenUsageCreate
)

class DatabaseService:
    def __init__(self):
        self.conn = db.get_connection()
    
    # Project operations
    def create_project(self, project_data: ProjectCreate) -> Project:
        import uuid
        project_id = str(uuid.uuid4())
        
        query = """
        INSERT INTO projects (id, name, template, docker_container, port, status, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, 'created', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        RETURNING *
        """
        result = self.conn.execute(
            query, 
            [project_id, project_data.name, project_data.template, project_data.docker_container, project_data.port]
        ).fetchone()
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
        result = self.conn.execute(
            query, 
            [project_data.name, project_data.template, project_data.docker_container, project_data.port, project_id]
        ).fetchone()
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
        result = self.conn.execute(query, [project_id]).fetchone()
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
        result = self.conn.execute(query, [name]).fetchone()
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
        results = self.conn.execute(query).fetchall()
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
    
    # Conversation operations
    def create_conversation_message(self, message_data: ConversationMessageCreate) -> ConversationMessage:
        import uuid
        message_id = str(uuid.uuid4())
        
        query = """
        INSERT INTO conversation_messages (id, project_id, role, content, message_type, model, provider, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        RETURNING *
        """
        result = self.conn.execute(
            query,
            [
                message_id, message_data.project_id, message_data.role, message_data.content,
                message_data.message_type, message_data.model, message_data.provider
            ]
        ).fetchone()
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
        SELECT * FROM conversation_messages 
        WHERE project_id = ? AND message_type = 'chat'
        ORDER BY created_at ASC
        """
        results = self.conn.execute(query, [project_id]).fetchall()
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
