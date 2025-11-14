# CI/CD Implementation Guide for adk-code

Based on ADR 0003: CI/CD and Build Process for adk-code

## Overview

The `adk-code` project uses GitHub Actions for continuous integration, automated testing, security scanning, and release management. This guide explains how the CI/CD system works and how to use it for development and releases.

## Quick Start

### For Developers

```bash
# Before committing: run local CI checks
make ci-check    # Runs: fmt, vet, lint, test

# This ensures your changes will pass GitHub Actions CI
```

### For Releases

```bash
# 1. Ensure all changes are on main/develop and committed
# 2. Create and push a version tag
git tag v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3

# GitHub Actions automatically:
# - Validates the tag format
# - Builds for 6 platforms
# - Creates a GitHub Release with all binaries
# - Generates SHA256 checksums
```

## Architecture

### Component Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    GitHub Actions CI/CD                      │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────┐   │
│  │  CI Pipeline (ci.yml)                                │   │
│  │  ├─ Format Check (gofmt)                             │   │
│  │  ├─ Static Analysis (go vet, golangci-lint)          │   │
│  │  ├─ Security Scanning (gosec, govulncheck)           │   │
│  │  ├─ Testing (race detection, coverage)               │   │
│  │  └─ Build Matrix (6 platforms)                       │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Release Pipeline (release.yml)                       │   │
│  │  ├─ Tag Validation (semantic versioning)             │   │
│  │  ├─ Multi-Platform Build                            │   │
│  │  ├─ Checksum Generation                             │   │
│  │  └─ GitHub Release Creation                         │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Build Tools                                         │   │
│  │  ├─ Makefile (local development)                    │   │
│  │  ├─ scripts/version.sh (version management)         │   │
│  │  └─ scripts/build-release.sh (cross-platform build) │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## CI Workflow (`ci.yml`)

**Trigger Events:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop`

### Jobs (Run in Parallel)

#### 1. Format Check
- **Command:** `gofmt -l`
- **Purpose:** Ensures consistent code formatting
- **Fail Condition:** Any Go file is improperly formatted
- **Local Fix:** `make fmt`

#### 2. Go Vet
- **Command:** `go vet ./...`
- **Purpose:** Detects suspicious code patterns
- **Fail Condition:** Vet reports issues
- **Local Fix:** `go vet ./...` and fix reported issues

#### 3. Lint (golangci-lint)
- **Command:** `golangci-lint run ./... --timeout=5m`
- **Purpose:** Comprehensive code quality analysis
- **Fail Condition:** Lint errors found
- **Local Fix:** `make lint`

#### 4. Security Scan
- **Tools:**
  - `gosec`: Go security issues (SQL injection, hardcoded secrets, etc.)
  - `govulncheck`: Known vulnerabilities in dependencies
- **Purpose:** Detect security vulnerabilities
- **Upload:** Results to GitHub Code Scanning (SARIF format)
- **Local Fix:** 
  ```bash
  go install golang.org/x/vuln/cmd/govulncheck@latest
  govulncheck ./...
  ```

#### 5. Test
- **Command:** `go test -v -race -coverprofile=coverage.out ./...`
- **Features:**
  - Race condition detection (`-race`)
  - Verbose output (`-v`)
  - Coverage measurement
  - Coverage threshold check (minimum 70%)
- **Upload:** Coverage to Codecov
- **Local Fix:** `make test` and write tests for uncovered code

#### 6. Build (Matrix)
- **Platforms:**
  - Linux: amd64, arm64, armv7
  - macOS: amd64 (Intel), arm64 (Apple Silicon)
  - Windows: amd64
- **Features:**
  - Optimized build flags (`-s -w`)
  - Version stamping via ldflags
  - Artifact upload (5-day retention)
- **Local Build:** `make cross-build`

#### 7. CI Complete
- **Purpose:** Final check that all jobs passed
- **Effect:** Prevents merge if any job failed

### Performance

**Typical total time:** 5-10 minutes (with parallel jobs)

| Job | Duration |
|---|---|
| Format | ~5s |
| Vet | ~10s |
| Lint | ~30-60s |
| Security | ~15-30s |
| Test | ~30-120s |
| Build (6 platforms) | ~3-5 min |

## Release Workflow (`release.yml`)

**Trigger Event:** Pushing a version tag matching `v*` (e.g., `v1.2.3`, `v1.2.3-rc1`)

### Validation Job

- **Validates tag format:** `v{MAJOR}.{MINOR}.{PATCH}[-prerelease]`
- **Detects prerelease:** Tags with `-alpha`, `-beta`, `-rc`, etc.
- **Example valid tags:**
  - `v1.0.0` (stable release)
  - `v1.0.1` (patch release)
  - `v2.0.0-alpha1` (prerelease)
  - `v1.2.3-rc2` (release candidate)

### Build Release Job

- **Builds all 6 platforms** (same matrix as CI)
- **Generates checksums:** SHA256 for each binary
- **Artifacts:** Uploaded with 1-day retention (auto-cleaned after release)

### Create Release Job

- **Downloads all artifacts**
- **Generates release notes with:**
  - Platform table with download links and file sizes
  - SHA256 checksums for integrity verification
  - Changelog (commits since previous release)
- **Creates GitHub Release with:**
  - All binaries as downloadable assets
  - Full release notes as description
  - Prerelease flag set correctly
- **Available at:** `https://github.com/raphaelmansuy/adk-code/releases/tag/{version}`

### Post-Release Job

- **Runs only:** For non-prerelease versions after successful release
- **Actions:** Logs success, suggests next steps

## Local Development Workflow

### Setup

```bash
# Install golangci-lint (if not present)
brew install golangci-lint  # macOS
# or: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest
```

### Making Changes

```bash
# 1. Create feature branch
git checkout -b feature/my-feature

# 2. Make your changes
# ... edit files ...

# 3. Run local CI checks (mimics GitHub Actions)
make ci-check    # Runs: fmt, vet, lint, test

# 4. Fix any issues
make fmt         # Auto-format code
make lint        # Review and fix lint issues
make test        # Write tests for coverage

# 5. Commit and push
git add .
git commit -m "feat: description"
git push origin feature/my-feature

# 6. Create PR on GitHub
# CI runs automatically on the PR
```

### Testing Locally

```bash
# Run tests with coverage
make test
make coverage    # Opens HTML report

# Run only short tests (faster)
make test-short

# Run tests with race detection
go test -race ./...

# Check specific coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Release Workflow

### Create a Release

```bash
# 1. Ensure changes are committed to main/develop
git status

# 2. Verify CI passes
# Run locally: make ci-check
# Or check: https://github.com/raphaelmansuy/adk-code/actions

# 3. Create release tag
git tag v1.2.3 -m "Release v1.2.3: Major features and bug fixes"
git push origin v1.2.3

# 4. GitHub Actions automatically:
#    - Validates tag
#    - Builds for all platforms
#    - Creates GitHub Release
#    - Uploads binaries and checksums
#    - Generates release notes

# 5. Monitor progress
# Visit: https://github.com/raphaelmansuy/adk-code/actions

# 6. Download released binaries
# Visit: https://github.com/raphaelmansuy/adk-code/releases/tag/v1.2.3
```

### Version Management

```bash
# Get current version
./scripts/version.sh get        # e.g., 1.0.0.42

# Bump build number (development)
./scripts/version.sh bump       # e.g., 1.0.0.43

# Set specific version
./scripts/version.sh set 1.0.0.5
```

### Build for Release Locally

```bash
# Build for all platforms
make cross-build

# Build with specific version
./scripts/build-release.sh v1.2.3

# Output: ./dist/adk-code-v1.2.3-{os}-{arch}[.exe]
```

## Platform & Architecture Support

### Supported Platforms (MVP)

| OS | Architectures | Status |
|---|---|---|
| **Linux** | amd64, arm64, armv7 | ✅ Fully Supported |
| **macOS** | amd64, arm64 | ✅ Fully Supported |
| **Windows** | amd64 | ✅ Fully Supported |

### Future Platforms

- FreeBSD (amd64, arm64)
- Linux (ppc64le, s390x) for enterprise
- WebAssembly (WASM) for web integration

## Build Optimization

### Release Build Flags

```bash
-s              # Strip symbol table (smaller binary)
-w              # Omit DWARF debug info (smaller binary)
-trimpath       # Remove filesystem paths (reproducible)
```

### Typical Binary Sizes

- Linux amd64: ~15-20 MB
- Linux arm64: ~15-20 MB
- Linux armv7: ~14-18 MB
- macOS amd64: ~15-20 MB
- macOS arm64: ~15-20 MB
- Windows amd64: ~16-21 MB

## Troubleshooting

### Code Formatting Issues

**Problem:** CI fails with "Code formatting issues found"

```bash
# Fix: Format your code
make fmt

# Commit the changes
git add . && git commit -m "style: format code"
```

### Vet Failures

**Problem:** CI fails with vet errors

```bash
# Check locally
go vet ./...

# Fix issues in your code
# Common issues: unused variables, suspicious code
```

### Low Test Coverage

**Problem:** CI fails with "Coverage 65% is below threshold of 70%"

```bash
# Check coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# View coverage gaps
go tool cover -html=coverage.out

# Write tests to reach 70% threshold
```

### Build Fails for Specific Platform

**Problem:** "Build failed for linux/arm64"

```bash
# Test locally if possible
GOOS=linux GOARCH=arm64 go build -o test-binary ./adk-code

# Check build logs for platform-specific errors
# May need to use CGO_ENABLED=0 for pure Go builds
```

### Security Vulnerabilities Detected

**Problem:** govulncheck finds vulnerabilities

```bash
# List vulnerabilities
govulncheck ./...

# Update vulnerable packages
go get -u vulnerable-package

# Or skip if false positive (document in code)
```

## GitHub Actions Debugging

### View CI Logs

1. Go to: https://github.com/raphaelmansuy/adk-code/actions
2. Click on the failed workflow run
3. Expand the failed job to see detailed logs

### Run CI Locally (Optional)

Install `act` tool:
```bash
brew install act  # macOS
```

Run locally:
```bash
# Simulate CI pipeline
act -j format    # Just format check
act -j test      # Just tests
act push         # Simulate push event
```

### Common CI Log Patterns

- **"gofmt -l" output:** Code formatting needed → Run `make fmt`
- **"no such file or directory":** Missing file in git → Run `git add .`
- **"undefined" in build:** Missing dependency → Run `go mod tidy`
- **Race detector output:** Concurrent access bug → Fix in code

## Advanced Configuration

### Customizing Build Matrix

Edit `.github/workflows/ci.yml` to add/remove platforms:

```yaml
strategy:
  matrix:
    include:
      - { goos: linux, goarch: amd64, name: "Linux x86-64" }
      # Add more platforms here
```

### Adjusting Coverage Threshold

Edit `.github/workflows/ci.yml`:

```bash
THRESHOLD=70  # Change this value (currently 70%)
```

### Custom Build Tags

Edit build step in workflow:

```yaml
go build -tags "custom_tag" \
  -ldflags="-s -w -X adk-code/internal/app.AppVersion=${VERSION}" \
  -o "${BINARY}" .
```

## Performance Optimization

### Speed Up Builds

1. **Use Go module caching** (already enabled)
2. **Parallelize jobs** (already done)
3. **Fail fast** - Early format/vet checks catch issues quickly
4. **Use `-short` flag** for quick test runs in development

### Monitor Usage

GitHub Actions free tier includes:
- 2,000 minutes/month for private repos
- Unlimited for public repos (like adk-code)

Current approximate usage:
- ~10 min per CI run × 10 runs/week = ~100 min/week = ~400 min/month
- Well within free tier limits

## Integration with IDEs

### VS Code

Install extensions:
- **Go** (golang.go) - Official Go support
- **GitHub Actions** (github.github-actions) - View/run workflows

### GoLand/IntelliJ IDEA

Built-in support:
- Run tests directly from editor
- View coverage in-editor
- Git integration for commits/pushes

## Security & Privacy

### No Secrets Required

Public CI uses default `GITHUB_TOKEN` automatically.

For future private integrations:
1. Store secrets in: Settings → Secrets and variables → Actions
2. Reference: `${{ secrets.SECRET_NAME }}`
3. Never commit secrets to repository

### Dependency Security

- **govulncheck**: Scans every build for vulnerabilities
- **Dependabot**: (Optional) Can be configured for automatic updates
- **gosec**: Scans for security patterns in code

## References

- **ADR 0003:** `/docs/adr/0003-cicd-and-build-process.md`
- **Go Cross-Compilation:** https://golang.org/doc/install/source#environment
- **GitHub Actions:** https://docs.github.com/en/actions
- **Semantic Versioning:** https://semver.org/
- **Makefile:** Run `make help` for all targets
- **Build Script:** `./adk-code/scripts/build-release.sh`

## Summary of Key Commands

```bash
# Local development
make ci-check                    # Run all CI checks before committing
make fmt                         # Format code
make test                        # Run tests
make coverage                    # View coverage report
make cross-build                 # Build for all platforms

# Release
git tag v1.2.3                   # Create release tag
git push origin v1.2.3           # Trigger release workflow

# Version management
./scripts/version.sh get         # Get current version
./scripts/version.sh bump        # Increment build number
./scripts/version.sh set 1.0.0   # Set specific version
```

## Getting Help

1. **Check workflow logs:** https://github.com/raphaelmansuy/adk-code/actions
2. **Run locally:** Most CI steps can be run with `make ci-check`
3. **Review ADR:** `/docs/adr/0003-cicd-and-build-process.md`
4. **Check Makefile:** `make help`

---

**Last Updated:** November 14, 2025  
**Status:** Implementation Complete (Phase 1 & 2)  
**Related:** ADR 0003, Makefile, .github/workflows/
