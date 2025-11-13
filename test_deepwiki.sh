#!/bin/bash

# Test DeepWiki MCP integration
# This script tests the DeepWiki MCP server integration with timeouts

cd "$(dirname "$0")" || exit 1

echo "Testing DeepWiki MCP Server Integration"
echo "========================================"
echo ""

# Test 1: Check if server loads
echo "Test 1: Loading MCP Server..."
timeout 15 ./code-agent.sh --mcp-config code_agent/examples/mcp/deepwiki-mcp.json <<< "/mcp list" 2>&1 | grep -A 5 "Configured MCP"

echo ""
echo "Test 2: Checking available tools..."
timeout 15 ./code-agent.sh --mcp-config code_agent/examples/mcp/deepwiki-mcp.json <<< "/mcp tools" 2>&1 | grep -A 10 "Tools from MCP"

echo ""
echo "Test 3: Testing with simple request (5 second limit)..."
timeout 8 ./code-agent.sh --mcp-config code_agent/examples/mcp/deepwiki-mcp.json <<< "tell me about golang in one sentence" 2>&1 | tail -20 || echo "[Timeout - tool execution may be slow]"

echo ""
echo "Done!"
