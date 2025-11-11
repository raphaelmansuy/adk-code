# 2025-11-11 Working Directory Flag Implementation

## Summary

Added support for specifying a working directory via `--working-directory` flag. The agent can now be invoked with an explicit working directory, enabling it to work in different locations without changing the current shell directory.

## Changes Made

### 1. Enhanced GetProjectRoot() in `coding_agent.go`
**Previous**: Only searched upward from the starting directory for `go.mod`
**Updated**: Now searches:
1. Current path for `go.mod`
2. Immediate subdirectories (e.g., `code_agent/`) for `go.mod`
3. Parent directories upward

**Benefit**: Allows the agent to work when invoked from parent directories of the project

### 2. Added WorkingDirectory CLI Flag in `cli.go`
**New Field**: `CLIConfig.WorkingDirectory`
**New Flag**: `--working-directory` (optional, defaults to current directory)

```bash
# Usage examples:
./code-agent --working-directory /path/to/project
./code-agent --working-directory ~
./code-agent --working-directory ~/my-project
```

### 3. Updated main.go to Handle Working Directory
**Changes**:
- Check for `--working-directory` flag first
- Fall back to `os.Getwd()` if flag not provided
- Expand `~` to home directory using `os.UserHomeDir()`
- Pass resolved working directory to agent

**Code Logic**:
```go
workingDir := cliConfig.WorkingDirectory
if workingDir == "" {
    workingDir, err = os.Getwd()  // Default to current directory
}

// Expand ~ in the path
if strings.HasPrefix(workingDir, "~") {
    homeDir, err := os.UserHomeDir()
    workingDir = filepath.Join(homeDir, workingDir[1:])
}
```

## Test Results

✅ Build successful with new flag
✅ `--help` shows new `--working-directory` flag
✅ Default behavior (no flag) uses current directory
✅ `--working-directory /absolute/path` works correctly
✅ `--working-directory ~` expands to home directory and works
✅ Agent initializes successfully with specified working directory

## Usage Examples

### Default (current directory)
```bash
cd /Users/raphaelmansuy/Github/03-working/adk_training_go
./bin/code-agent
# Working directory: /Users/raphaelmansuy/Github/03-working/adk_training_go
```

### Specify absolute path
```bash
./bin/code-agent --working-directory /Users/raphaelmansuy/Github/03-working/adk_training_go
```

### Use home directory
```bash
./bin/code-agent --working-directory ~
# Working directory: /Users/raphaelmansuy
```

### Use relative path with tilde
```bash
./bin/code-agent --working-directory ~/Github/03-working/adk_training_go
```

### From any directory
```bash
cd /tmp
/Users/raphaelmansuy/Github/03-working/adk_training_go/bin/code-agent --working-directory /Users/raphaelmansuy/Github/03-working/adk_training_go
```

## Files Modified

1. **code_agent/cli.go**
   - Added `WorkingDirectory` field to `CLIConfig`
   - Added `--working-directory` flag to `ParseCLIFlags()`

2. **code_agent/main.go**
   - Updated working directory logic to use flag value or default to `os.Getwd()`
   - Added path expansion for `~` to home directory

3. **code_agent/agent/coding_agent.go**
   - Enhanced `GetProjectRoot()` to search subdirectories and parent directories

## Benefits

1. **Flexibility**: Users can specify working directory explicitly
2. **Consistency**: Can work in the same directory from anywhere
3. **Path Expansion**: Supports `~` for home directory
4. **Backward Compatible**: Default behavior unchanged (uses current directory)
5. **Better Project Detection**: Searches subdirectories for `go.mod`

## Technical Notes

- `os.UserHomeDir()` is used for secure home directory detection
- Path expansion handles edge cases (just `~` with no trailing slash)
- Working directory is used to find project root via `GetProjectRoot()`
- Agent operates relative to the found project root, not the working directory

## Exit Codes

- 0: Normal execution
- 1: Initialization errors (e.g., invalid working directory)
- 130: SIGINT (Ctrl+C)

## Future Enhancements

1. Add `--home` shortcut flag that expands to `~/`
2. Add config file support for default working directory
3. Add environment variable support: `CODE_AGENT_WORKING_DIR`
4. Project-specific configurations in `.code_agent.yml`
