# This file deploy function template and return the project path, container name, and port
import os
import shutil
from ..config import PROJECTS_DIR, PROJECTS_TEMPLATE_DIR, DOCK_ROUTE_PATH

def deploy_app(template_name: str,project_name: str, container_name: str, port: int) -> dict:
    """Deploy the application and return deployment details."""
    try:
        # Define the project path
        template_path = os.path.join(PROJECTS_TEMPLATE_DIR, template_name)
        project_path = os.path.join(PROJECTS_DIR, project_name)
        
        # Copy template files to the project directory
        shutil.copytree(template_path, project_path)
        
        # Define the command and its arguments as a list
        command_as_list = [
            DOCK_ROUTE_PATH,
            "deploy",
            "reactjs",
            container_name,
            project_path,
            "--host-port",
            str(port),
            "--image",
            container_name
        ]
        execute_command(command_as_list)
        
        deployment_details = {
            "project_path": project_path,
            "container_name": container_name,
            "port": port
        }
        
        return deployment_details
    except Exception as e:
        raise RuntimeError(f"Deployment failed: {str(e)}")


def check_container_status(container_name: str) -> dict:
    """
    Check the status of a specific Docker container.
    
    Args:
        container_name: Name of the Docker container
        
    Returns:
        dict: Container status information with detailed parsing
    """
    import subprocess
    import re
    
    try:
        # Use dock-route list to check container status
        command_as_list = [
            DOCK_ROUTE_PATH,
            "list",
            "containers"
        ]
        
        result = subprocess.run(
            command_as_list,
            capture_output=True,
            text=True,
            timeout=30
        )
        
        if result.returncode == 0:
            # Parse the output to find our specific container
            lines = result.stdout.split('\n')
            container_found = False
            container_info = {
                "exists": False,
                "status": "Not found",
                "running": False,
                "image": "",
                "ports": "",
                "subdomain": ""
            }
            
            i = 0
            while i < len(lines):
                line = lines[i].strip()
                
                # Look for container name pattern: - **container-name**
                if f"**{container_name}**" in line:
                    container_found = True
                    container_info["exists"] = True
                    
                    # Parse the next few lines for container details
                    j = i + 1
                    while j < len(lines) and j < i + 6:  # Look ahead max 5 lines
                        detail_line = lines[j].strip()
                        
                        if detail_line.startswith("Image:"):
                            container_info["image"] = detail_line.replace("Image:", "").strip()
                        elif detail_line.startswith("Status:"):
                            status_text = detail_line.replace("Status:", "").strip()
                            container_info["status"] = status_text
                            # Check if container is running - "Up" indicates running state
                            container_info["running"] = ("running" in status_text.lower() or 
                                                        "up" in status_text.lower())
                        elif detail_line.startswith("Ports:"):
                            container_info["ports"] = detail_line.replace("Ports:", "").strip()
                        elif detail_line.startswith("Subdomain:"):
                            container_info["subdomain"] = detail_line.replace("Subdomain:", "").strip()
                        elif detail_line.startswith("- **") and "**" in detail_line:
                            # Hit another container, stop parsing this one
                            break
                        elif detail_line == "":
                            # Empty line might indicate end of this container's info
                            break
                        
                        j += 1
                    
                    break  # Found our container, no need to continue
                
                i += 1
            
            if not container_found:
                return {
                    "exists": False,
                    "status": "Container not found in dock-route managed containers",
                    "running": False,
                    "error": f"Container '{container_name}' not found in list"
                }
            
            return container_info
            
        else:
            return {
                "exists": False,
                "status": "Error listing containers",
                "running": False,
                "error": result.stderr or "Failed to list containers"
            }
            
    except subprocess.TimeoutExpired:
        return {
            "exists": False,
            "status": "Timeout",
            "running": False,
            "error": "Timeout while checking container status"
        }
    except Exception as e:
        return {
            "exists": False,
            "status": "Error",
            "running": False,
            "error": str(e)
        }


def execute_container_command(container_name: str, command: str) -> dict:
    """
    Execute a command in a running Docker container using dock-route exec.
    
    Args:
        container_name: Name of the Docker container
        command: Command to execute in the container
        
    Returns:
        dict: Result containing success status, stdout, stderr, and return code
    """
    import subprocess
    
    # First check if container exists and is running
    status = check_container_status(container_name)
    if not status["exists"]:
        return {
            "success": False,
            "stdout": "",
            "stderr": f"Container '{container_name}' not found",
            "return_code": -1,
            "command": command,
            "container_status": status
        }
    
    if not status["running"]:
        return {
            "success": False,
            "stdout": "",
            "stderr": f"Container '{container_name}' is not running. Status: {status['status']}",
            "return_code": -1,
            "command": command,
            "container_status": status
        }
    
    # If container just started (Up X seconds), wait a bit for it to fully initialize
    if "up" in status["status"].lower() and ("second" in status["status"].lower() or "minute" in status["status"].lower()):
        import time
        # Extract the time value to determine wait time
        if "second" in status["status"].lower():
            # If it's been up for less than 30 seconds, wait a bit more
            try:
                import re
                time_match = re.search(r'up (\d+) second', status["status"].lower())
                if time_match and int(time_match.group(1)) < 30:
                    print(f"Container recently started, waiting 5 seconds for initialization...")
                    time.sleep(5)
            except:
                # If parsing fails, just wait 3 seconds
                time.sleep(3)
    
    try:
        # Build the dock-route exec command
        command_as_list = [
            DOCK_ROUTE_PATH,
            "exec",
            container_name,
            "--"
        ] + command.split()
        
        print(f"🚀 Running container command: {' '.join(command_as_list)}")
        
        result = subprocess.run(
            command_as_list,
            capture_output=True,
            text=True,
            encoding='utf-8',
            errors='replace',  # Replace invalid UTF-8 characters
            timeout=300  # 5 minute timeout for package installations
        )
        
        return {
            "success": result.returncode == 0,
            "stdout": result.stdout,
            "stderr": result.stderr,
            "return_code": result.returncode,
            "command": command,
            "container_status": status
        }
        
    except subprocess.TimeoutExpired:
        return {
            "success": False,
            "stdout": "",
            "stderr": "Command timed out after 5 minutes",
            "return_code": -1,
            "command": command,
            "container_status": status
        }
    except FileNotFoundError:
        return {
            "success": False,
            "stdout": "",
            "stderr": f"dock-route executable not found at {DOCK_ROUTE_PATH}",
            "return_code": -1,
            "command": command,
            "container_status": status
        }
    except Exception as e:
        return {
            "success": False,
            "stdout": "",
            "stderr": str(e),
            "return_code": -1,
            "command": command,
            "container_status": status
        }


def execute_command(command_as_list: dict) -> bool:
    import subprocess

    try:
        # Execute the command
        print(f"🚀 Running command: {' '.join(command_as_list)}")
        
        result = subprocess.run(
            command_as_list, 
            check=True,          # Raise an exception if the command fails
            capture_output=True, # Capture stdout and stderr
            text=True            # Decode stdout/stderr as text
        )

        # Print the standard output and standard error
        print("\n✅ Command executed successfully!")
        if result.stdout:
            print("\n--- STDOUT ---")
            print(result.stdout)
        if result.stderr:
            print("\n--- STDERR ---")
            print(result.stderr)
            
        return True

    except FileNotFoundError:
        print(f"❌ Error: The command '{command_as_list[0]}' was not found.")
        print("Please ensure the path to the executable is correct.")
        return False

    except subprocess.CalledProcessError as e:
        # This block will run if the command returns a non-zero exit code (an error)
        print(f"\n❌ Command failed with exit code {e.returncode}")
        if e.stdout:
            print("\n--- STDOUT ---")
            print(e.stdout)
        if e.stderr:
            print("\n--- STDERR ---")
            print(e.stderr)
        return False

def list_all_containers() -> dict:
    """
    List all Docker containers managed by dock-route.
    
    Returns:
        dict: All containers information
    """
    import subprocess
    
    try:
        command_as_list = [
            DOCK_ROUTE_PATH,
            "list",
            "containers"
        ]
        
        result = subprocess.run(
            command_as_list,
            capture_output=True,
            text=True,
            timeout=30
        )
        
        return {
            "success": result.returncode == 0,
            "output": result.stdout,
            "error": result.stderr if result.returncode != 0 else None
        }
        
    except Exception as e:
        return {
            "success": False,
            "output": "",
            "error": str(e)
        }


def restart_container(container_name: str) -> dict:
    """
    Restart a container by stopping and starting it.
    
    Args:
        container_name: Name of the Docker container
        
    Returns:
        dict: Result of restart operation
    """
    import subprocess
    
    try:
        # First stop the container
        stop_result = subprocess.run(
            [DOCK_ROUTE_PATH, "stop", container_name],
            capture_output=True,
            text=True,
            timeout=60
        )
        
        # Then start it
        start_result = subprocess.run(
            [DOCK_ROUTE_PATH, "start", container_name],
            capture_output=True,
            text=True,
            timeout=60
        )
        
        return {
            "success": start_result.returncode == 0,
            "stop_output": stop_result.stdout,
            "start_output": start_result.stdout,
            "error": start_result.stderr if start_result.returncode != 0 else None
        }
        
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


def ensure_container_running(container_name: str) -> dict:
    """
    Ensure a container is running, start it if it's not.
    This is for project management, not agent tool usage.
    
    Args:
        container_name: Name of the Docker container
        
    Returns:
        dict: Result of the operation
    """
    import subprocess
    
    try:
        # First check container status
        status = check_container_status(container_name)
        
        if not status["exists"]:
            return {
                "success": False,
                "action": "check",
                "error": f"Container '{container_name}' does not exist"
            }
        
        if status["running"]:
            return {
                "success": True,
                "action": "already_running",
                "status": status["status"]
            }
        
        # Container exists but not running, start it
        start_result = subprocess.run(
            [DOCK_ROUTE_PATH, "start", container_name],
            capture_output=True,
            text=True,
            encoding='utf-8',
            errors='replace',
            timeout=60
        )
        
        if start_result.returncode == 0:
            # Wait a moment for container to fully start
            import time
            time.sleep(3)
            
            # Check status again
            new_status = check_container_status(container_name)
            
            return {
                "success": True,
                "action": "started",
                "status": new_status["status"],
                "output": start_result.stdout
            }
        else:
            return {
                "success": False,
                "action": "start_failed",
                "error": start_result.stderr,
                "output": start_result.stdout
            }
            
    except subprocess.TimeoutExpired:
        return {
            "success": False,
            "action": "timeout",
            "error": "Container start operation timed out"
        }
    except Exception as e:
        return {
            "success": False,
            "action": "error",
            "error": str(e)
        }


def get_container_status_for_project(container_name: str) -> dict:
    """
    Get container status specifically for project management.
    
    Args:
        container_name: Name of the Docker container
        
    Returns:
        dict: Container status and management info
    """
    status = check_container_status(container_name)
    
    return {
        "container_name": container_name,
        "exists": status["exists"],
        "running": status["running"],
        "status": status["status"],
        "needs_start": status["exists"] and not status["running"],
        "error": status.get("error")
    }