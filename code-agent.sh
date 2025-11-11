#!/bin/bash

# code-agent.sh
# Convenience script to start the code agent CLI

set -e

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Path to the binary
BINARY="$SCRIPT_DIR/bin/code-agent"

# Check if binary exists
if [ ! -f "$BINARY" ]; then
    echo "Error: code-agent binary not found at $BINARY"
    echo ""
    echo "Building the code agent..."
    cd "$SCRIPT_DIR/code_agent"
    make build
    cd "$SCRIPT_DIR"
fi

# Run the code agent
"$BINARY" "$@"
