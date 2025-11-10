#!/bin/bash

# Test script for code agent

echo "Testing Code Agent..."
echo ""

# Check if GOOGLE_API_KEY is set
if [ -z "$GOOGLE_API_KEY" ]; then
    echo "❌ Error: GOOGLE_API_KEY environment variable is not set"
    echo "Please set it with: export GOOGLE_API_KEY='your-api-key'"
    exit 1
fi

echo "✓ GOOGLE_API_KEY is set"
echo ""

# Check if code-agent binary exists
if [ ! -f "./code-agent" ]; then
    echo "Building code-agent..."
    go build -o code-agent
    if [ $? -ne 0 ]; then
        echo "❌ Build failed"
        exit 1
    fi
    echo "✓ Build successful"
fi

echo "✓ code-agent binary found"
echo ""

# Create a test directory
TEST_DIR="test_workspace"
rm -rf $TEST_DIR
mkdir -p $TEST_DIR

echo "✓ Created test workspace: $TEST_DIR"
echo ""

echo "===========================================" 
echo "Code Agent is ready!"
echo "==========================================="
echo ""
echo "You can now run the agent with:"
echo "  ./code-agent"
echo ""
echo "Or test it interactively by running it and asking:"
echo "  - 'Create a simple hello.go file with a main function'"
echo "  - 'List the files in the current directory'"
echo "  - 'Read the README.md file'"
echo ""
echo "Type 'exit' or 'quit' to stop the agent."
echo ""
