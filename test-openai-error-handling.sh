#!/bin/bash

# Test script to verify OpenAI error handling
# This will test that tool errors are properly conveyed to the model

cd "$(dirname "$0")"

echo "Testing OpenAI error handling..."
echo ""

# Run the code agent with a command that should trigger a tool call
echo "List the files in the current directory" | ./bin/code-agent --model openai/gpt-4o-mini --no-persist

echo ""
echo "Test complete!"
