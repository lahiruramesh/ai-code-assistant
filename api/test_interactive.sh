#!/bin/bash

echo "Starting interactive Agent Loop Manager test..."
echo "This will start the multiagent system and wait for your input."
echo ""
echo "Try these commands:"
echo "1. Type: Create a React todo app"
echo "2. Type: Create a React weather app"
echo "3. Wait and see both loops processing concurrently"
echo "4. Press Ctrl+C when ready to exit"
echo ""

# Run the multiagent system interactively
./multiagent -mode=cli -llm=bedrock
