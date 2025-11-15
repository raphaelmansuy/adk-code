# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Google Search tool integration via ADK's `geminitool.GoogleSearch`
  - Enables web search capabilities for the agent
  - Works with Gemini 2.0+ models
  - Auto-registered in Search & Discovery category with highest priority
  - See ADR 0005 for implementation details
- New `websearch` package in `tools/` for web search functionality
- Comprehensive unit tests for Google Search tool
- Documentation in TOOL_DEVELOPMENT.md for using ADK built-in tools

## [Unreleased]

## [0.2.1] - 2025-11-14

### Fixed

- Enable CGO for SQLite support in all build targets to ensure proper database functionality across all platforms

## [0.2.0] - 2025-11-14

### Added

- Phase 3 Distribution Channels documentation
  - ADR-0004: Comprehensive distribution strategy covering Homebrew, APT, YUM, and Scoop
  - PHASE3_DISTRIBUTION_GUIDE.md: Step-by-step implementation guide for multi-platform distribution
  - Complete scripts and templates for package manager automation
  - Security best practices for cryptographic signing (GPG, Cosign)
  - Testing matrix for all platforms and architectures
- Homebrew installation guide with platform-specific instructions
- Improved workspace project root detection for better path resolution

### Changed

- Enhanced project root detection to handle both go.mod and .git files
- Improved workspace switching and multi-workspace support
- Updated dependencies with stability improvements

### Fixed

- Workspace path resolution in various edge cases
- Project root detection in CI/CD environments
- Enhanced coding agent prompts for better guidance

## [0.1.1] - 2025-11-14

## [0.0.1] - 2025-11-14

### Added

- Initial project structure
- Dynamic Ollama model discovery
- Support for multiple LLM providers:
  - Google Gemini (Vertex AI)
  - OpenAI models
  - Ollama local models
- Interactive REPL with command history
- File and workspace operations
- Code search and analysis tools
- Terminal execution capabilities
- Model context protocol (MCP) support
- Configuration management
- Session persistence

### Features

- **Display System**: Terminal UI with ANSI colors and markdown rendering
- **Model System**: LLM provider abstraction with capability tracking
- **Agent Loop**: ADK-based agentic framework for autonomous operations
- **Session Management**: Conversation history and token tracking
- **Workspace Tools**: Multi-root path resolution with VCS awareness
- **Tool Integration**: 30+ autonomous tools across 8 categories

## Unreleased Features

### Planned

- GPG signing of releases
- Cosign for container image signing
- Software Bill of Materials (SBOM) generation
- Homebrew distribution support (Phase 3A)
- APT/YUM package manager support (Phase 3B-3C)
- Scoop (Windows) package manager support (Phase 3D)
- Enhanced security scanning
- Container image builds and distribution
- Documentation site deployment
- Performance benchmarking suite

## How to Release

### Version Format

Follow [Semantic Versioning](https://semver.org/):

- `MAJOR.MINOR.PATCH` (e.g., `1.2.3`)
- Pre-release: `MAJOR.MINOR.PATCH-alpha.1`, `-beta.1`, `-rc.1`

### Release Steps

1. Update version in `.version` file
2. Update this CHANGELOG.md with release date and version
3. Commit: `git commit -m "chore: release v1.2.3"`
4. Create tag: `git tag v1.2.3`
5. Push: `git push origin main && git push origin v1.2.3`
6. GitHub Actions automatically creates release with binaries

### Changelog Guidelines

- Keep sections organized by type (Added, Changed, Fixed, etc.)
- Use clear, descriptive language
- Link to related issues/PRs when applicable
- Mark unreleased changes under `[Unreleased]`

## Semantic Versioning

This project follows [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backward-compatible functionality additions
- **PATCH** version for backward-compatible bug fixes

### Pre-release Versions

- Alpha: `v1.0.0-alpha.1` - Early development, unstable
- Beta: `v1.0.0-beta.1` - Feature complete, testing phase
- Release Candidate: `v1.0.0-rc.1` - Final pre-release testing

## Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that don't affect code meaning (formatting, etc.)
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Code change that improves performance
- `test`: Adding or updating tests
- `chore`: Changes to build process, dependencies, etc.
- `ci`: Changes to CI/CD configuration

### Examples

- `feat: add multi-platform binary support`
- `fix: resolve version script path in CI`
- `docs: update CI/CD guide`
- `chore: update dependencies`

## Links

- [Repository](https://github.com/raphaelmansuy/adk-code)
- [Issues](https://github.com/raphaelmansuy/adk-code/issues)
- [Releases](https://github.com/raphaelmansuy/adk-code/releases)
- [CI/CD Guide](docs/CI_CD_GUIDE.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Contributing](CONTRIBUTING.md)

## Versioning History

### Branch Strategy

- `main` - Stable releases only
- `develop` - Integration branch for features
- `feature/*` - Feature branches
- `hotfix/*` - Critical bug fixes

### Release Cadence

- Releases are created as needed (not on a fixed schedule)
- Security fixes are released as patch versions immediately
- Feature releases are coordinated and tested before release
