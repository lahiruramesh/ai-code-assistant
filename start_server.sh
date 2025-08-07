#!/bin/bash

echo "Starting Multi-Agent Server for Frontend Integration..."
echo "======================================================"

# Create required directories
mkdir -p /tmp/aiagent /tmp/projects

# Copy necessary files
cp -r /Users/lahiruramesh/myspace/code-editing-agent/api/.env /tmp/aiagent/
cp -r /Users/lahiruramesh/myspace/code-editing-agent/api/prompts /tmp/aiagent/

# Change to working directory
cd /tmp/aiagent

echo "Starting server on port 8084..."
echo "Frontend can connect to: http://localhost:8084"
echo "Use Ctrl+C to stop the server"
echo ""

# Start the multiagent system in server mode
/Users/lahiruramesh/myspace/code-editing-agent/api/bin/multiagent -mode=server -llm=openrouter -port=8084
