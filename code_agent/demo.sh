#!/bin/bash

# Quick demo of the code agent with a simple task

echo "🚀 Testing Code Agent with a simple task..."
echo ""

# Check if GOOGLE_API_KEY is set
if [ -z "$GOOGLE_API_KEY" ]; then
    echo "❌ Error: GOOGLE_API_KEY environment variable is not set"
    echo "Please set it with: export GOOGLE_API_KEY='your-api-key'"
    exit 1
fi

# Create a test input file
TEST_INPUT="demo_input.txt"
echo "List all Go files in the current directory" > $TEST_INPUT

echo "Input: List all Go files in the current directory"
echo ""
echo "Running agent..."
echo ""

# Run the agent with the test input (in a non-interactive way, we'll pipe the input)
# For now, just show how to run it
echo "To test the agent interactively, run:"
echo "  ./code-agent"
echo ""
echo "Then try these commands:"
echo "  - List all files in the current directory"
echo "  - Create a file called hello.go with a main function"
echo "  - Read the README.md file"
echo ""

rm -f $TEST_INPUT
