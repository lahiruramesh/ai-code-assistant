#!/bin/bash

echo "Testing Multi-Agent System..."
echo "=============================="

# Create required directories
mkdir -p /tmp/aiagent /tmp/projects

# Copy necessary files
cp /Users/lahiruramesh/myspace/code-editing-agent/api/.env /tmp/aiagent/
cp -r /Users/lahiruramesh/myspace/code-editing-agent/api/prompts /tmp/aiagent/

# Change to working directory
cd /tmp/aiagent

echo "Starting interactive mode..."
echo "Type: Create a React todo app with add and delete functionality"
echo "Then wait for processing to complete"
echo "Use 'quit' to exit when done"
echo ""

# Start the multiagent system in interactive mode
/Users/lahiruramesh/myspace/code-editing-agent/api/bin/multiagent -mode=cli -llm=openrouter
