#!/usr/bin/env bash
# make.sh - convenience wrapper to invoke `make` in ./adk-code
# Usage:
#   ./make.sh           -> runs `make` in ./adk-code with no args
#   ./make.sh test      -> runs `make test` in ./adk-code
#   ./make.sh -k build  -> forwards arguments to make

set -euo pipefail

# Resolve the script directory reliably (works if sourced or executed)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

# Change to the adk-code directory inside the repository
cd "${SCRIPT_DIR}/adk-code"

# Print what we're running for transparency
echo "==> running: make $* (in ${PWD})"

# Forward all arguments to `make`
exec make "$@"
