#!/bin/bash

# Test script for OpenAI tool calling

cd "$(dirname "$0")"

echo "Testing OpenAI tool calling..."
echo ""

# Test with a simple directory listing
echo "list the files" | ./code-agent.sh --model openai/gpt-4o-mini
