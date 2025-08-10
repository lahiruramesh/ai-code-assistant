import os
import subprocess
import aiofiles
from langchain.tools import Tool, tool
from typing import List
from ..config import PROJECTS_DIR
from ..utils.docker_route import execute_container_command, check_container_status, list_all_containers, restart_container

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

def get_tools_for_project(project_path: str, container_name: str = None) -> List[Tool]:
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
            
            output = f"🖥️ Host Command Executed\n"
            output += f"Command: {command}\n"
            output += f"Directory: {project_path}\n"
            output += f"Success: {'✅ Yes' if result.returncode == 0 else '❌ No'}\n"
            output += f"Return code: {result.returncode}\n"
            
            if result.stdout:
                output += f"\n📤 STDOUT:\n{result.stdout}"
            if result.stderr:
                output += f"\n📥 STDERR:\n{result.stderr}"
            
            # Provide suggestions for common issues
            if result.returncode != 0:
                if "command not found" in result.stderr.lower():
                    output += f"\n💡 Suggestion: Command not found. If this is a container-specific command, use execute_container_command instead."
                elif "permission denied" in result.stderr.lower():
                    output += f"\n💡 Suggestion: Permission denied. Check file permissions or try with appropriate privileges."
            
            return output
        except subprocess.TimeoutExpired:
            os.chdir(original_cwd)
            return "⏰ Error: Command timed out after 30 seconds"
        except Exception as e:
            os.chdir(original_cwd)
            return f"❌ Error running command: {str(e)}"

    def get_project_info_tool(dummy_input: str = "") -> str:
        """Get information about the current project"""
        try:
            info = [f"📁 Project Path: {project_path}"]
            
            if container_name:
                info.append(f"🐳 Docker Container: {container_name}")
                
                # Get detailed container status
                status = check_container_status(container_name)
                if status["exists"]:
                    info.append(f"   Status: {status['status']}")
                    info.append(f"   Running: {'✅ Yes' if status['running'] else '❌ No'}")
                    if status.get("image"):
                        info.append(f"   Image: {status['image']}")
                    if status.get("ports"):
                        info.append(f"   Ports: {status['ports']}")
                else:
                    info.append(f"   ❌ Container not found or not managed by dock-route")
            
            # Check if it's a git repository
            if os.path.exists(os.path.join(project_path, '.git')):
                info.append("📦 Git repository detected")
            
            # Check for common project files
            common_files = ['package.json', 'tsconfig.json', 'vite.config.ts', 'next.config.js']
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

    def manage_container_tool(action: str) -> str:
        """Manage the Docker container for this project"""
        if not container_name:
            return "❌ Error: No Docker container associated with this project"
        
        action = action.lower().strip()
        
        try:
            if action == "status":
                status = check_container_status(container_name)
                output = f"🐳 Container Status for '{container_name}':\n"
                output += f"Exists: {'✅ Yes' if status['exists'] else '❌ No'}\n"
                if status['exists']:
                    output += f"Status: {status['status']}\n"
                    output += f"Running: {'✅ Yes' if status['running'] else '❌ No'}\n"
                    if status.get('image'):
                        output += f"Image: {status['image']}\n"
                    if status.get('ports'):
                        output += f"Ports: {status['ports']}\n"
                else:
                    output += f"Error: {status.get('error', 'Unknown error')}\n"
                return output
                
            elif action == "restart":
                result = restart_container(container_name)
                if result["success"]:
                    return f"✅ Container '{container_name}' restarted successfully"
                else:
                    return f"❌ Failed to restart container '{container_name}': {result.get('error', 'Unknown error')}"
                    
            elif action == "list":
                result = list_all_containers()
                if result["success"]:
                    return f"📋 All Containers:\n{result['output']}"
                else:
                    return f"❌ Failed to list containers: {result.get('error', 'Unknown error')}"
                    
            else:
                return f"❌ Unknown action '{action}'. Available actions: status, restart, list"
                
        except Exception as e:
            return f"❌ Error managing container: {str(e)}"

    def wait_and_retry_tool(action: str) -> str:
        """Wait for container initialization and retry operations"""
        import time
        
        if not container_name:
            return "❌ Error: No Docker container associated with this project"
        
        try:
            if action.lower() == "wait":
                # Wait for container to fully initialize
                status = check_container_status(container_name)
                if status["exists"] and status["running"]:
                    if "up" in status["status"].lower() and "second" in status["status"].lower():
                        print("⏳ Waiting for container to fully initialize...")
                        time.sleep(10)
                        # Check status again
                        new_status = check_container_status(container_name)
                        return f"✅ Container initialization wait completed. New status: {new_status['status']}"
                    else:
                        return f"✅ Container appears to be fully initialized. Status: {status['status']}"
                else:
                    return f"❌ Container is not running. Status: {status['status']}"
            else:
                return f"❌ Unknown action '{action}'. Available actions: wait"
                
        except Exception as e:
            return f"❌ Error in wait and retry: {str(e)}"

    def execute_container_command_tool(command: str) -> str:
        """Execute a command in the Docker container for this project"""
        if not container_name:
            return "Error: No Docker container associated with this project"
        
        try:
            result = execute_container_command(container_name, command)
            
            output = f"🚀 Container Command Executed\n"
            output += f"Command: {result['command']}\n"
            output += f"Container: {container_name}\n"
            output += f"Success: {'✅ Yes' if result['success'] else '❌ No'}\n"
            output += f"Return Code: {result['return_code']}\n"
            
            if result['stdout']:
                output += f"\n📤 STDOUT:\n{result['stdout']}"
            
            if result['stderr']:
                output += f"\n📥 STDERR:\n{result['stderr']}"
            
            # Show container status if available
            if 'container_status' in result:
                status = result['container_status']
                output += f"\n🐳 Container Status: {status['status']}"
            
            # Provide helpful suggestions based on common scenarios
            if not result['success']:
                if "not found" in result['stderr'].lower() or "not running" in result['stderr'].lower():
                    output += f"\n💡 Suggestion: The container '{container_name}' might not be running."
                    output += f"\n   Try: get_project_info to check status, or restart the container"
                elif "permission denied" in result['stderr'].lower():
                    output += f"\n💡 Suggestion: Permission issue detected. Try running with appropriate permissions."
                elif "enoent" in result['stderr'].lower():
                    output += f"\n💡 Suggestion: Command or file not found. Check if the command exists in the container."
                elif "package.json" in command and "install" in command:
                    output += f"\n💡 Suggestion: Package installation failed. Check if package.json exists and is valid."
                elif "pnpm" in command and "command not found" in result['stderr'].lower():
                    output += f"\n💡 Suggestion: pnpm not found. Try using 'npm' instead or install pnpm first."
            
            return output
            
        except Exception as e:
            error_msg = f"❌ Error executing container command: {str(e)}\n"
            error_msg += f"Container: {container_name}\n"
            error_msg += f"Command: {command}\n"
            error_msg += f"\n💡 Troubleshooting steps:\n"
            error_msg += f"1. Check if container is running: list_files or get_project_info\n"
            error_msg += f"2. Verify container name is correct\n"
            error_msg += f"3. Try starting the container if it's stopped\n"
            return error_msg

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
            description="""Run a shell command on the HOST system in the project directory.
            
            🎯 WHEN TO USE: For host-level operations (file system, git, etc.)
            
            ✅ BEST FOR:
            - Git operations: 'git status', 'git add .', 'git commit -m "message"'
            - File system operations: 'find . -name "*.ts"', 'chmod +x script.sh'
            - Host-level tools: 'docker ps', system commands
            
            ❌ AVOID FOR:
            - Package management (use execute_container_command instead)
            - Running development servers (use execute_container_command instead)
            
            Input: command to run on host system""",
            func=run_command_tool
        ),
        Tool(
            name="get_project_info",
            description="Get information about the current project structure and type, including container status",
            func=get_project_info_tool
        )
    ]
    
    # Add container tools if container is available
    if container_name:
        tools.extend([
            Tool(
                name="execute_container_command",
                description=f"""Execute a command inside the Docker container '{container_name}'. 
                
                🎯 WHEN TO USE: For operations that need to run inside the containerized environment
                
                ✅ BEST FOR:
                - Package management: 'pnpm install axios', 'pnpm install --save-dev @types/node'
                - shadcn/ui installation: 'pnpm dlx shadcn@latest add card button -y'
                - Running development server: 'pnpm dev', 'npm run dev'
                - Building project: 'pnpm build', 'npm run build'
                - Running tests: 'pnpm test', 'npm test'
                - Container-specific commands: 'ls -la', 'pwd', 'cat package.json'
                
                ⚠️ NOTE: If container shows "Up X seconds" and commands fail, use wait_and_retry first.
                
                Input: command to execute (without 'dock-route exec container-name --')""",
                func=execute_container_command_tool
            ),
            Tool(
                name="manage_container",
                description=f"""Manage the Docker container '{container_name}' for this project.
                
                🎯 WHEN TO USE: For container lifecycle management and troubleshooting
                
                ✅ AVAILABLE ACTIONS:
                - 'status': Check detailed container status and health
                - 'restart': Stop and start the container (fixes most issues)
                - 'list': Show all managed containers
                
                💡 USE CASES:
                - Before running container commands, check status
                - If container commands fail, restart the container
                - Troubleshoot container issues
                
                Input: action to perform (status/restart/list)""",
                func=manage_container_tool
            ),
            Tool(
                name="wait_and_retry",
                description=f"""Wait for container '{container_name}' to fully initialize.
                
                🎯 WHEN TO USE: When container shows "Up X seconds" but commands are failing
                
                ✅ BEST FOR:
                - After container restart before running commands
                - When shadcn/ui installation fails due to container not ready
                - Before package installations on newly started containers
                
                💡 USE CASES:
                - Container just started and needs time to initialize
                - Commands failing with "container not running" despite status showing "Up"
                
                Input: 'wait' to wait for container initialization""",
                func=wait_and_retry_tool
            )
        ])
    
    return tools