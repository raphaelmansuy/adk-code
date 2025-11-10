# Makefile Usage Guide

This Makefile provides convenient commands for building, testing, and managing the code_agent project.

## Quick Start

```bash
# Show all available commands
make help

# Build the application
make build

# Run tests
make test

# Clean and rebuild everything
make all
```

## Common Commands

### Building

```bash
make build          # Build the application
make build-debug    # Build with debug symbols
make release        # Build optimized release version
make clean          # Remove build artifacts
```

### Testing

```bash
make test           # Run all tests
make test-short     # Run short tests only
make coverage       # Generate coverage report (opens in browser)
make bench          # Run benchmarks
```

### Code Quality

```bash
make fmt            # Format code
make vet            # Run go vet
make lint           # Run golangci-lint (requires installation)
make check          # Run all checks (fmt, vet, lint, test)
```

### Dependencies

```bash
make deps           # Download dependencies
make deps-tidy      # Tidy go.mod and go.sum
make deps-update    # Update all dependencies
make deps-verify    # Verify dependencies
```

### Installation

```bash
make install        # Install to GOPATH/bin
make uninstall      # Remove from GOPATH/bin
make run            # Build and run
```

### Utilities

```bash
make info           # Show build information
make size           # Show binary size
make todo           # List TODO/FIXME items in code
make watch          # Auto-rebuild on changes (requires entr)
```

## Examples

### Development Workflow

```bash
# 1. Clean start
make clean

# 2. Download dependencies
make deps

# 3. Format code
make fmt

# 4. Run checks
make check

# 5. Build
make build

# 6. Run
./code-agent
```

### Pre-commit Workflow

```bash
# Run all quality checks before committing
make check
```

### Release Workflow

```bash
# Create optimized release build
make release

# Check size
make size

# Test
make test

# Install
make install
```

### Testing Workflow

```bash
# Run all tests with coverage
make coverage

# Run benchmarks
make bench
```

## Requirements

### Optional Tools

Some commands require additional tools:

- **golangci-lint** (for `make lint`)
  ```bash
  # macOS
  brew install golangci-lint
  
  # or with Go
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```

- **entr** (for `make watch`)
  ```bash
  # macOS
  brew install entr
  
  # Linux
  apt-get install entr
  ```

## Makefile Customization

You can customize the Makefile by modifying variables at the top:

```makefile
BINARY_NAME=code-agent    # Output binary name
VERSION=1.0.0             # Version number
BUILD_DIR=.               # Build output directory
```

## Tips

1. **Use tab completion**: Most shells support tab completion for make targets
2. **Chain commands**: `make clean build test`
3. **Parallel execution**: `make -j4` (run 4 jobs in parallel)
4. **Dry run**: `make -n build` (show commands without executing)
5. **Silent mode**: `make -s build` (suppress make output)

## Troubleshooting

### "Command not found"

If you see "command not found" errors:

```bash
# Check if make is installed
which make

# Install make if needed (macOS)
xcode-select --install

# Install make if needed (Linux)
sudo apt-get install build-essential
```

### "No such file or directory"

Make sure you're in the correct directory:

```bash
cd code_agent
pwd  # Should show: .../code_agent
```

### Dependencies not downloading

```bash
# Try cleaning first
make clean

# Then download
make deps

# Or update
make deps-update
```

## Color Output

The Makefile uses color coding:
- 🟢 **Green**: Success messages
- 🟡 **Yellow**: Warning messages
- 🔴 **Red**: Error messages

If colors don't show, check that your terminal supports ANSI colors.

## Integration with IDE

### VS Code

Add to `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "make build",
      "type": "shell",
      "command": "make build",
      "group": {
        "kind": "build",
        "isDefault": true
      }
    },
    {
      "label": "make test",
      "type": "shell",
      "command": "make test",
      "group": {
        "kind": "test",
        "isDefault": true
      }
    }
  ]
}
```

### GoLand/IntelliJ

1. Run → Edit Configurations
2. Add New Configuration → Makefile
3. Add targets: build, test, clean, etc.

## See Also

- [Go Documentation](https://golang.org/doc/)
- [Makefile Tutorial](https://makefiletutorial.com/)
- [Code Agent Documentation](../doc/README.md)
