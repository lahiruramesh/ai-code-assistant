# This file deploy function template and return the project path, container name, and port
import os
import shutil

def deploy_app(template_name: str,project_name: str, container_name: str, port: int) -> dict:
    """Deploy the application and return deployment details."""
    try:
        # Define the project path
        template_path = os.path.join(os.getenv("PROJECTS_TEMPLATE_DIR", "/tmp/projects/templates"), template_name)
        project_path = os.path.join(os.getenv("PROJECTS_DIR", "/tmp/projects"), project_name)
        
        # Copy template files to the project directory
        shutil.copytree(template_path, project_path)
        
        # Define the command and its arguments as a list
        command_as_list = [
            os.getenv("DOCK_ROUTE_PATH"),
            "deploy",
            "nextjs",
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