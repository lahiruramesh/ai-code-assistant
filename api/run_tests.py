#!/usr/bin/env python3
"""
Test runner script for the API with comprehensive coverage reporting.
"""
import subprocess
import sys
import os
from pathlib import Path


def run_tests():
    """Run all tests with coverage reporting."""
    
    # Ensure we're in the API directory
    api_dir = Path(__file__).parent
    os.chdir(api_dir)
    
    print("🧪 Running API Tests with Coverage")
    print("=" * 50)
    
    # Test commands to run
    commands = [
        # Run unit tests
        [
            "uv", "run", "pytest", 
            "tests/", 
            "-v", 
            "--tb=short",
            "--color=yes",
            "--durations=10"
        ],
        
        # Run tests with coverage
        [
            "uv", "run", "pytest", 
            "tests/", 
            "--cov=app",
            "--cov-report=html",
            "--cov-report=term-missing",
            "--cov-fail-under=80",
            "-v"
        ]
    ]
    
    for i, cmd in enumerate(commands, 1):
        print(f"\n📋 Running command {i}/{len(commands)}: {' '.join(cmd)}")
        print("-" * 40)
        
        try:
            result = subprocess.run(cmd, check=True, capture_output=False)
            print(f"✅ Command {i} completed successfully")
        except subprocess.CalledProcessError as e:
            print(f"❌ Command {i} failed with exit code {e.returncode}")
            if i == 1:  # If basic tests fail, don't run coverage
                print("Skipping coverage report due to test failures")
                return e.returncode
        except FileNotFoundError:
            print(f"❌ Command not found: {cmd[0]}")
            print("Make sure uv and pytest are installed")
            return 1
    
    print("\n🎉 All tests completed!")
    print("\n📊 Coverage report generated in htmlcov/index.html")
    return 0


def run_specific_tests():
    """Run specific test categories."""
    
    test_categories = {
        "unit": "tests/test_*.py -m 'not integration'",
        "integration": "tests/test_integration.py",
        "database": "tests/test_database_service.py",
        "api": "tests/test_projects.py tests/test_streaming.py tests/test_models_tokens.py",
        "auth": "tests/test_auth.py",
        "main": "tests/test_main.py"
    }
    
    if len(sys.argv) > 1:
        category = sys.argv[1]
        if category in test_categories:
            cmd = f"uv run pytest {test_categories[category]} -v"
            print(f"🧪 Running {category} tests: {cmd}")
            os.system(cmd)
        else:
            print(f"❌ Unknown test category: {category}")
            print(f"Available categories: {', '.join(test_categories.keys())}")
            return 1
    else:
        return run_tests()


if __name__ == "__main__":
    exit_code = run_specific_tests()
    sys.exit(exit_code)