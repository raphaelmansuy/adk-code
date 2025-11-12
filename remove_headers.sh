#!/bin/bash

# Script to remove Apache 2.0 license headers from all Go files in code_agent/
# 
# This script safely removes the 14-line Google LLC copyright header from Go files.
# It validates the header before removal and provides detailed feedback.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_DIR="$SCRIPT_DIR/code_agent"
BACKUP_DIR="$SCRIPT_DIR/.backup_headers_$(date +%s)"

# The exact 14-line header to remove
read -r -d '' EXPECTED_HEADER << 'EOF' || true
// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
EOF

echo "Removing Apache 2.0 license headers from Go files in: $TARGET_DIR"
echo "Creating backup directory: $BACKUP_DIR"
echo ""

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Counter for processed files
PROCESSED=0
MODIFIED=0
SKIPPED=0
FAILED=0

# Find all .go files in code_agent directory
while IFS= read -r file; do
    PROCESSED=$((PROCESSED + 1))
    
    # Extract the first 14 lines from the file
    file_header=$(head -n 14 "$file")
    
    # Check if the header matches the expected header
    if [[ "$file_header" == "$EXPECTED_HEADER" ]]; then
        echo "Processing: $file"
        
        # Create a backup
        backup_file="$BACKUP_DIR/$(basename "$file")"
        if ! cp "$file" "$backup_file"; then
            echo "  ✗ Failed to create backup"
            FAILED=$((FAILED + 1))
            continue
        fi
        
        # Create a temporary file with lines 15 onwards
        temp_file="${file}.tmp.$$"
        if ! tail -n +15 "$file" > "$temp_file"; then
            echo "  ✗ Failed to extract content after header"
            rm -f "$temp_file"
            FAILED=$((FAILED + 1))
            continue
        fi
        
        # Move temp file back to original (atomic operation)
        if ! mv "$temp_file" "$file"; then
            echo "  ✗ Failed to update file"
            rm -f "$temp_file"
            FAILED=$((FAILED + 1))
            continue
        fi
        
        # Verify the removal was successful
        if head -1 "$file" | grep -q "Copyright 2025 Google LLC"; then
            echo "  ✗ Verification failed - header still present"
            # Restore from backup
            cp "$backup_file" "$file"
            FAILED=$((FAILED + 1))
            continue
        fi
        
        MODIFIED=$((MODIFIED + 1))
        echo "  ✓ Header removed successfully"
    else
        echo "Skipping: $file (header not found or does not match exactly)"
        SKIPPED=$((SKIPPED + 1))
    fi
    
done < <(find "$TARGET_DIR" -name "*.go" -type f)

echo ""
echo "Summary:"
echo "  Total files processed: $PROCESSED"
echo "  Files modified: $MODIFIED"
echo "  Files skipped: $SKIPPED"
echo "  Files failed: $FAILED"
echo ""

if [[ $FAILED -eq 0 && $MODIFIED -gt 0 ]]; then
    echo "✓ All operations completed successfully!"
    echo "  Backups saved in: $BACKUP_DIR"
elif [[ $MODIFIED -eq 0 ]]; then
    echo "⚠ No files were modified."
    echo "  Removing empty backup directory..."
    rmdir "$BACKUP_DIR" 2>/dev/null || true
else
    echo "✗ Some operations failed. Review the output above."
    echo "  Backups preserved in: $BACKUP_DIR"
    exit 1
fi

echo "Done!"
