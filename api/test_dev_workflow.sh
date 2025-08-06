#!/bin/bash

echo "Testing development workflow with Docker hot reloading..."

# Create a test input focusing on development environment
{
    echo "Create a React todo app with Docker development environment. I want hot reloading so when I make changes to components, they reflect immediately in the browser. Set up volume mounts and enable package installation without rebuilding the container."
    sleep 1800  # Wait 3 minutes for complete setup
    echo "quit"
} | ./multiagent -mode=cli -llm=bedrock > dev_output.txt 2>&1 &

# Store the process ID
PROCESS_PID=$!

echo "Development workflow test started with PID: $PROCESS_PID"
echo "Check dev_output.txt for progress"
echo "Process will run for 3 minutes to complete full setup"

# Monitor the process
while kill -0 $PROCESS_PID 2>/dev/null; do
    echo "$(date): Development setup in progress..."
    sleep 30
done

echo "Development workflow test completed!"
echo "Check dev_output.txt for full results"
