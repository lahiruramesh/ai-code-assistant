#!/bin/bash

echo "Testing Agent Loop Manager..."
echo ""

# Create a test input file with multiple requests
cat << 'EOF' > /tmp/test_requests.txt
Create a simple React calculator app with basic arithmetic operations
Create a React weather app that displays current weather
quit
EOF

echo "Sending multiple requests to test concurrent agent loops:"
echo "1. React calculator app"
echo "2. React weather app"
echo ""

# Run the multiagent system with the test requests
./multiagent -mode=cli -llm=bedrock < /tmp/test_requests.txt

# Clean up
rm -f /tmp/test_requests.txt

echo "Test completed!"
