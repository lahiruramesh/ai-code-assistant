#!/bin/bash

echo "Testing tool calling with development workflow..."

# Create a test input focusing on actual file creation
{
    echo "Create a simple React counter app with Docker development setup. Use tools to actually create the files and verify they exist on the host machine."
    sleep 120  # Wait 2 minutes for complete setup
    echo "quit"
} | ./multiagent -mode=cli -llm=bedrock > tool_output.txt 2>&1 &

# Store the process ID
PROCESS_PID=$!

echo "Tool calling test started with PID: $PROCESS_PID"
echo "Check tool_output.txt for progress"
echo "Process will run for 2 minutes to test tool execution"

# Monitor the process
while kill -0 $PROCESS_PID 2>/dev/null; do
    echo "$(date): Tool calling test in progress..."
    sleep 30
done

echo "Tool calling test completed!"
echo "Check tool_output.txt for results and created files"
echo "Listing any created project files:"
find . -name "my-react-app" -type d 2>/dev/null || echo "No project directory found yet"
