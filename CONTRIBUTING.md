# Contributing to adk-code

Thank you for your interest in contributing! We welcome contributions of all kinds.

## Code of Conduct

Be respectful. We aim to maintain a welcoming, inclusive community.

## How to Contribute

### Reporting Bugs

Found a bug? [Open an issue](https://github.com/your-username/adk-code/issues) with:

- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Environment (Go version, OS, etc.)

### Suggesting Features

Have an idea? [Start a discussion](https://github.com/your-username/adk-code/discussions) or open an issue with:

- Clear use case
- How it benefits users
- Any alternative approaches

### Writing Code

**Quick Start:**

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/your-feature`
3. Make your changes
4. Run tests: `make check`
5. Commit with clear messages
6. Push and create a pull request

**Code Style:**

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `make fmt` before committing
- Run `make check` (fmt + vet + lint + test) before pushing

**Testing:**

- Write tests for new features
- Ensure `make test` passes
- Aim for >80% coverage

### Adding Tools

Creating a new tool? Follow [TOOL_DEVELOPMENT.md](docs/TOOL_DEVELOPMENT.md):

```go
// 1. Define input/output types
type MyInput struct { Path string }
type MyOutput struct { Result string }

// 2. Implement handler
func handler(ctx Context, input MyInput) MyOutput { ... }

// 3. Register with functiontool
func init() { common.Register(...) }
```

See [docs/TOOL_DEVELOPMENT.md](docs/TOOL_DEVELOPMENT.md) for complete guide.

### Documenting Changes

- Update [docs/README.md](docs/README.md) if adding features
- Update [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) if changing core
- Add comments to complex functions
- Update this file if process changes

## Development Workflow

```bash
# Setup
cd adk-code
go mod download

# Build & test
make build
make test

# Quality checks (required)
make check

# Watch mode for development
make watch
```

## Pull Request Process

1. **Before submitting**: Run `make check` and ensure all tests pass
2. **Descriptive title**: "feat: add X" or "fix: resolve issue #123"
3. **Clear description**: What changed and why
4. **Link issues**: "Fixes #123" or "Related to #456"
5. **Keep it focused**: One feature or fix per PR
6. **Be responsive**: Address feedback promptly

## Getting Help

- **Questions?** [Start a discussion](https://github.com/your-username/adk-code/discussions)
- **Stuck?** Check [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) or existing code
- **Design help?** Open an issue for discussion before coding

## Release Process

(For maintainers)

```bash
# Update version
make version-set V=1.2.3

# Tag and push
git tag v1.2.3
git push origin main --tags

# Create release
# (Automated via GitHub Actions)
```

## Project Structure

```
adk-code/
├── main.go                    # Entry point
├── go.mod                     # Dependencies
├── Makefile                   # Build targets
├── internal/
│   ├── app/                   # Application
│   ├── orchestration/         # Component wiring
│   ├── repl/                  # CLI loop
│   ├── display/               # Terminal UI
│   ├── session/               # Persistence
│   ├── config/                # Configuration
│   ├── cli/                   # Commands
│   └── ...
├── pkg/
│   ├── models/                # LLM abstraction
│   ├── errors/                # Error handling
│   └── workspace/             # Path resolution
├── tools/                     # Tool ecosystem
│   ├── file/
│   ├── edit/
│   ├── exec/
│   └── ...
└── docs/                      # Documentation
```

## Key Concepts

- **Tool Pattern**: 4 steps (types → handler → wrapper → register)
- **Components**: Display, Model, Agent, Session
- **Agent Loop**: LLM → Tool calls → Execution → Results → Repeat
- **Session**: Persistent storage of conversations + token tracking

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for details.

## Common Tasks

### Add a new tool

1. Create `tools/category/tool_name.go`
2. Follow 4-step pattern (see TOOL_DEVELOPMENT.md)
3. Test: `make test`
4. Register and export in `tools/tools.go`

### Fix a bug

1. Add test that reproduces bug
2. Fix the issue
3. Ensure test passes
4. Run `make check`

### Add documentation

1. Edit relevant `.md` file
2. Use clear examples
3. Keep it concise
4. Update nav if needed

### Update dependencies

1. `make deps-update`
2. Test thoroughly
3. `make check`
4. Commit changes

## Questions?

- 💬 [Discussions](https://github.com/your-username/adk-code/discussions)
- 🐛 [Issues](https://github.com/your-username/adk-code/issues)
- 📚 [Documentation](docs/)

---

**Thank you for contributing!** 🙏

Every contribution—from bug reports to code to documentation—helps make adk-code better.
