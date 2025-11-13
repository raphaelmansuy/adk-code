#!/bin/bash

# Script to manage build version numbers
# Usage:
#   version.sh get     - Get current version
#   version.sh bump    - Increment build number and return new version
#   version.sh set V   - Set version to V

VERSION_FILE=".version"

# Ensure version file exists
if [ ! -f "$VERSION_FILE" ]; then
    echo "1.0.0.1" > "$VERSION_FILE"
fi

get_version() {
    cat "$VERSION_FILE" | tr -d '\n'
}

bump_version() {
    local current=$(get_version)
    
    # Split version into parts
    IFS='.' read -r major minor patch build <<< "$current"
    
    # Increment build number
    ((build++))
    
    # Reconstruct version
    local new_version="${major}.${minor}.${patch}.${build}"
    
    # Write back to file
    echo "$new_version" > "$VERSION_FILE"
    
    # Return new version
    echo "$new_version"
}

set_version() {
    local new_version="$1"
    if [ -z "$new_version" ]; then
        echo "Error: Version argument required" >&2
        exit 1
    fi
    echo "$new_version" > "$VERSION_FILE"
    echo "$new_version"
}

case "${1:-get}" in
    get)
        get_version
        ;;
    bump)
        bump_version
        ;;
    set)
        set_version "$2"
        ;;
    *)
        echo "Usage: $0 {get|bump|set VERSION}" >&2
        exit 1
        ;;
esac
