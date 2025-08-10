import os
import subprocess
import aiofiles
from langchain.tools import Tool, tool
from typing import List

PROJECTS_DIR = os.getenv("PROJECTS_DIR", "./projects")

@tool
async def write_file(project_name: str, file_path: str, content: str) -> str:
    """
    Writes content to a specified file within a project directory.
    Useful for creating or updating code files, configurations, etc.
    'project_name' is the name of the project folder.
    'file_path' is the relative path of the file within the project.
    'content' is the data to be written to the file.
    """
    try:
        full_path = os.path.join(PROJECTS_DIR, project_name, file_path)
        os.makedirs(os.path.dirname(full_path), exist_ok=True)
        async with aiofiles.open(full_path, "w", encoding="utf-8") as f:
            await f.write(content)
        return f"File '{file_path}' has been written successfully in project '{project_name}'."
    except Exception as e:
        return f"Error writing file: {str(e)}"

def get_tools_for_project(project_path: str) -> List[Tool]:
    """Get tools that are aware of the project context"""
    
    def read_file_tool(file_path: str) -> str:
        """Read a file from the project directory"""
        try:
            full_path = os.path.join(project_path, file_path)
            # Security check
            if not full_path.startswith(os.path.abspath(project_path)):
                return "Error: Access denied - file outside project directory"
            
            with open(full_path, 'r', encoding='utf-8') as f:
                content = f.read()
            return f"Content of {file_path}:\n{content}"
        except FileNotFoundError:
            return f"Error: File {file_path} not found"
        except Exception as e:
            return f"Error reading file: {str(e)}"

    def write_file_tool(input_str: str) -> str:
        """Write content to a file in the project directory
        Input format: filename|content"""
        try:
            parts = input_str.split('|', 1)
            if len(parts) != 2:
                return "Error: Input must be in format 'filename|content'"
            
            file_path, content = parts
            full_path = os.path.join(project_path, file_path)
            
            # Security check
            if not full_path.startswith(os.path.abspath(project_path)):
                return "Error: Access denied - file outside project directory"
            
            # Create directory if it doesn't exist
            os.makedirs(os.path.dirname(full_path), exist_ok=True)
            
            with open(full_path, 'w', encoding='utf-8') as f:
                f.write(content)
            return f"Successfully wrote to {file_path}"
        except Exception as e:
            return f"Error writing file: {str(e)}"

    def list_files_tool(directory: str = ".") -> str:
        """List files and directories in the project"""
        try:
            full_path = os.path.join(project_path, directory)
            
            # Security check
            if not full_path.startswith(os.path.abspath(project_path)):
                return "Error: Access denied - directory outside project"
            
            if not os.path.exists(full_path):
                return f"Error: Directory {directory} not found"
            
            items = []
            for item in sorted(os.listdir(full_path)):
                if item.startswith('.'):
                    continue  # Skip hidden files
                item_path = os.path.join(full_path, item)
                if os.path.isdir(item_path):
                    items.append(f"📁 {item}/")
                else:
                    size = os.path.getsize(item_path)
                    items.append(f"📄 {item} ({size} bytes)")
            
            return f"Contents of {directory}:\n" + "\n".join(items)
        except Exception as e:
            return f"Error listing directory: {str(e)}"

    def run_command_tool(command: str) -> str:
        """Run a shell command in the project directory"""
        try:
            # Change to project directory
            original_cwd = os.getcwd()
            os.chdir(project_path)
            
            # Run command with timeout
            result = subprocess.run(
                command,
                shell=True,
                capture_output=True,
                text=True,
                timeout=30  # 30 second timeout
            )
            
            # Restore original directory
            os.chdir(original_cwd)
            
            output = ""
            if result.stdout:
                output += f"STDOUT:\n{result.stdout}\n"
            if result.stderr:
                output += f"STDERR:\n{result.stderr}\n"
            output += f"Return code: {result.returncode}"
            
            return output
        except subprocess.TimeoutExpired:
            os.chdir(original_cwd)
            return "Error: Command timed out after 30 seconds"
        except Exception as e:
            os.chdir(original_cwd)
            return f"Error running command: {str(e)}"

    def get_project_info_tool(dummy_input: str = "") -> str:
        """Get information about the current project"""
        try:
            info = [f"Project Path: {project_path}"]
            
            # Check if it's a git repository
            if os.path.exists(os.path.join(project_path, '.git')):
                info.append("📦 Git repository detected")
            
            # Check for common project files
            common_files = ['package.json', 'requirements.txt', 'Cargo.toml', 'go.mod', 'pom.xml', 'composer.json']
            for file in common_files:
                if os.path.exists(os.path.join(project_path, file)):
                    info.append(f"📄 Found {file}")
            
            # Count files and directories
            total_files = 0
            total_dirs = 0
            for root, dirs, files in os.walk(project_path):
                dirs[:] = [d for d in dirs if not d.startswith('.')]
                total_dirs += len(dirs)
                total_files += len([f for f in files if not f.startswith('.')])
            
            info.append(f"📊 {total_files} files, {total_dirs} directories")
            
            return "\n".join(info)
        except Exception as e:
            return f"Error getting project info: {str(e)}"

    # Create the tools list
    tools = [
        Tool(
            name="read_file",
            description="Read the contents of a file. Input: file path relative to project root",
            func=read_file_tool
        ),
        Tool(
            name="write_file",
            description="Write content to a file. Input format: 'filename|content'",
            func=write_file_tool
        ),
        Tool(
            name="list_files",
            description="List files and directories. Input: directory path (default: current directory)",
            func=list_files_tool
        ),
        Tool(
            name="run_command",
            description="Run a shell command in the project directory. Input: command to run",
            func=run_command_tool
        ),
        Tool(
            name="get_project_info",
            description="Get information about the current project structure and type",
            func=get_project_info_tool
        )
    ]
    
    return tools

# A list of all available tools for the agent (including the original write_file tool)
available_tools = [write_file] + get_tools_for_project("./projects")