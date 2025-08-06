#!/bin/bash

echo "Testing single agent loop..."

# Create a test input with single request and long wait
{
    echo "Create a simple React todo app with add, delete, and toggle functionality"
    sleep 1000  # Wait 2 minutes for processing
    echo "quit"
} | ./multiagent -mode=cli -llm=bedrock > output.txt 2>&1


echo "Single loop test completed!"
