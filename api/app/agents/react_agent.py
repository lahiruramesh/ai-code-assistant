import os
from dotenv import load_dotenv
from langchain_openai import ChatOpenAI
from langchain.agents import AgentExecutor, create_react_agent
from app.prompts.react_prompts import react_prompt
from app.agents.tools import get_tools_for_project
from ..config import MODEL_NAME, OPENROUTER_API_KEY, OPENROUTER_API_BASE

load_dotenv()

class ReActAgent:
    def __init__(self, project_path: str = None, container_name: str = None):
        self.project_path = project_path or "/tmp/projects"
        self.container_name = container_name
        
        # Initialize the LLM from OpenRouter
        self.llm = ChatOpenAI(
            model=MODEL_NAME,
            openai_api_key=OPENROUTER_API_KEY,
            openai_api_base=OPENROUTER_API_BASE,
            streaming=True,
            temperature=0.1,
        )
        
        # Get tools with project context
        self.tools = get_tools_for_project(self.project_path, self.container_name)
        
        # Create the agent with project-aware prompt
        self.prompt = self._get_project_aware_prompt()
        self.agent = create_react_agent(self.llm, self.tools, self.prompt)
        self.agent_executor = AgentExecutor(
            agent=self.agent,
            tools=self.tools,
            verbose=True,
            handle_parsing_errors=True # Helps with robustness
        )

    def _get_project_aware_prompt(self):
        """Get a prompt that includes project context"""
        project_context = f"""
You are working on a project located at: {self.project_path}

When using tools, always consider the project context and work within the project directory.
If you need to create, edit, or analyze files, they should be relative to the project path.
"""
        return react_prompt.partial(project_context=project_context)

    async def stream_response(self, user_input: str, project_path: str = None, container_name: str = None):
        """Streams the agent's thoughts and actions with project context."""
        if project_path:
            self.project_path = project_path
            if container_name:
                self.container_name = container_name
            # Update tools with new project path and container
            self.tools = get_tools_for_project(self.project_path, self.container_name)
            self.agent = create_react_agent(self.llm, self.tools, self._get_project_aware_prompt())
            self.agent_executor = AgentExecutor(
                agent=self.agent,
                tools=self.tools,
                verbose=True,
                handle_parsing_errors=True
            )
        
        # Add project context to user input
        enhanced_input = f"""
Project Path: {self.project_path}
User Request: {user_input}

Please help with this request in the context of the project at the specified path.
"""
        
        # The `astream_log` method provides detailed, structured output
        async for chunk in self.agent_executor.astream_log(
            {"input": enhanced_input},
            include_names=["ChatOpenAI"], # Filter for LLM outputs if needed
        ):
            # Process and format the chunk for better frontend consumption
            processed_chunk = self._process_chunk(chunk)
            if processed_chunk:
                yield processed_chunk
    
    def _process_chunk(self, chunk):
        """Process and format chunks for better frontend consumption"""
        if not chunk:
            return None
        
        # Extract meaningful content from the chunk
        if hasattr(chunk, 'ops') and chunk.ops:
            for op in chunk.ops:
                if op.get('op') == 'add' and 'content' in op.get('value', {}):
                    content = op['value']['content']
                    if isinstance(content, str) and content.strip():
                        return {
                            "type": "content",
                            "content": content,
                            "source": "agent"
                        }
        
        # Handle other chunk formats
        if isinstance(chunk, dict):
            if 'content' in chunk:
                return {
                    "type": "content", 
                    "content": chunk['content'],
                    "source": "agent"
                }
        
        # Fallback for any other format
        return {
            "type": "raw",
            "content": str(chunk),
            "source": "agent"
        }