#!/bin/bash
# Quick visual test for display improvements

cd "$(dirname "$0")"

echo "=== Testing Display Enhancements ==="
echo ""
echo "This will demonstrate:"
echo "  1. Left border on agent responses"
echo "  2. Smart path truncation for long file paths"
echo "  3. Enhanced completion message"
echo "  4. Warning/Info message styling"
echo ""
echo "Running a simple query..."
echo ""

# Test with a simple query that will show our improvements
echo "List the files in the current directory" | ./code-agent

echo ""
echo "=== Display Test Complete ==="
