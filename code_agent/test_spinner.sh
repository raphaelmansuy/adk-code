#!/bin/bash

# Test script to verify spinner improvements
# This will test the code_agent with various commands

echo "Testing spinner improvements..."
echo ""
echo "Test 1: Simple file read"
echo "read main.go" | ./code-agent --output-format=rich

echo ""
echo "Test 2: List directory"
echo "list files in current directory" | ./code-agent --output-format=rich

echo ""
echo "Test 3: Multiple operations"
echo "list the files then read main.go" | ./code-agent --output-format=rich

echo ""
echo "Tests complete!"
