# ADR 0003: CI/CD and Build Process for adk-code

## Status
Proposed

## Context

The `adk-code` project is a multi-model AI coding assistant CLI built with Go 1.24+ and uses the ADK framework. Currently, the project has:

- A functional `Makefile` with local build targets (build, test, install, etc.)
- A version management script (`scripts/version.sh`) for semantic versioning
- No automated CI/CD pipeline for testing, building, or releasing
- No cross-platform binary distribution mechanism
- No automated releases to GitHub or other package repositories

As the project matures, we need:
- **Automated Quality Assurance**: Run tests, linting, and code quality checks on every commit
- **Multi-Platform Builds**: Support binaries for Linux (arm64, amd64, armv7), macOS (arm64, amd64), Windows (amd64), and potentially more
- **Reproducible Builds**: Ensure consistent builds with proper version stamping and build metadata
- **Release Automation**: Automatically create GitHub releases with platform-specific binaries
- **Integration Testing**: Validate functionality across platforms before release
- **Package Distribution**: Support installation via package managers and direct binary downloads

This ADR establishes the architecture, tooling, and best practices for a professional Go CI/CD system that scales with the project.

## Decision

### 1. CI/CD Platform: GitHub Actions

**Rationale:**
- Native integration with GitHub repository
- Free tier suitable for open-source projects
- Strong Go ecosystem support
- Extensive marketplace for pre-built actions
- Matrix builds for multi-platform compilation
- Native artifact and release management

**Implementation:**
- Workflows stored in `.github/workflows/` directory
- Clear naming convention: `ci.yml`, `release.yml`, `security.yml`
- Reusable workflow components for DRY principle

### 2. Build Process Architecture

#### 2.1 Local Development Builds
**Leverage existing Makefile:**
- `make build`: Development build with version stamping
- `make build-debug`: Debug build with symbols for `dlv` debugger
- `make release`: Optimized release build (stripped, no debug info)
- Version auto-bump using `scripts/version.sh`

**Platform-specific builds:**
```bash
# Linux amd64
GOOS=linux GOARCH=amd64 make build

# macOS arm64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 make build

# Windows amd64
GOOS=windows GOARCH=amd64 make build
```

#### 2.2 Cross-Compilation Strategy
Use Go's native cross-compilation capabilities with environment variables:
- `GOOS`: Target operating system (linux, darwin, windows, freebsd)
- `GOARCH`: Target architecture (amd64, arm64, arm, 386, ppc64le)
- Architecture-specific features via `GOARM`, `GOAMD64` for enhanced ISA support

**Supported Platforms (MVP):**
| OS | Architectures | Binary Name |
|---|---|---|
| Linux | amd64, arm64, armv7 | adk-code-linux-{arch} |
| macOS | amd64, arm64 | adk-code-darwin-{arch} |
| Windows | amd64 | adk-code-windows-amd64.exe |

**Optional/Future:**
- FreeBSD (amd64, arm64)
- Linux (ppc64le, s390x) - for enterprise
- WASM - for web integration

#### 2.3 Build Script Strategy
Create `scripts/build-release.sh` for reproducible cross-platform builds:
```bash
#!/bin/bash
# Builds adk-code for all supported platforms
# Usage: ./scripts/build-release.sh <version>
# Output: ./dist/adk-code-v{version}-{os}-{arch}[.exe]

set -e

VERSION="${1:-$(./scripts/version.sh get)}"
DIST_DIR="./dist"
BINARY_NAME="adk-code"

# Define build matrix
PLATFORMS=(
  "linux:amd64"
  "linux:arm64"
  "linux:arm:7"
  "darwin:amd64"
  "darwin:arm64"
  "windows:amd64"
)

for platform in "${PLATFORMS[@]}"; do
  IFS=':' read -r GOOS GOARCH GOARM <<< "$platform"
  
  OUTPUT="${DIST_DIR}/${BINARY_NAME}-v${VERSION}-${GOOS}-${GOARCH}"
  [[ "$GOOS" == "windows" ]] && OUTPUT="${OUTPUT}.exe"
  
  echo "Building ${GOOS}/${GOARCH}..."
  GOOS="$GOOS" GOARCH="$GOARCH" GOARM="$GOARM" \
    go build -ldflags="-s -w -X adk-code/internal/app.AppVersion=v${VERSION}" \
    -o "$OUTPUT" .
  
  # Print size info
  ls -lh "$OUTPUT" | awk '{print "  Size: " $5}'
done
```

### 3. Continuous Integration Workflow (`ci.yml`)

**Trigger:** On push to main/feature branches and pull requests

**Stages:**

```yaml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  quality:
    runs-on: ubuntu-latest
    name: Code Quality
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
          cache: true
      
      # Code formatting
      - name: Check formatting
        run: |
          if [ -n "$(gofmt -l .)" ]; then
            gofmt -l .
            exit 1
          fi
      
      # Static analysis
      - name: Run go vet
        run: go vet ./...
      
      # Linting
      - name: Install golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout=5m
      
      # Security scanning
      - name: Run gosec
        uses: securego/gosec@master
        with:
          args: '-no-fail -fmt sarif -out gosec-results.sarif ./...'
      
      - name: Upload gosec results
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: gosec-results.sarif

  test:
    runs-on: ubuntu-latest
    name: Tests
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
          cache: true
      
      - name: Run unit tests
        run: go test -v -race -coverprofile=coverage.out ./...
      
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          flags: unittests

  build:
    runs-on: ubuntu-latest
    needs: [quality, test]
    name: Build Binaries
    strategy:
      matrix:
        include:
          - { goos: linux, goarch: amd64 }
          - { goos: linux, goarch: arm64 }
          - { goos: linux, goarch: arm, goarm: 7 }
          - { goos: darwin, goarch: amd64 }
          - { goos: darwin, goarch: arm64 }
          - { goos: windows, goarch: amd64 }
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
          cache: true
      
      - name: Build ${{ matrix.goos }}-${{ matrix.goarch }}
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
          GOARM: ${{ matrix.goarm || '0' }}
        run: |
          VERSION=$(./scripts/version.sh get)
          BINARY="adk-code-${{ matrix.goos }}-${{ matrix.goarch }}"
          [[ "${{ matrix.goos }}" == "windows" ]] && BINARY="${BINARY}.exe"
          
          go build -ldflags="-s -w -X adk-code/internal/app.AppVersion=${VERSION}" \
            -o "${BINARY}" .
          
          ls -lh "${BINARY}"
      
      - name: Upload binary
        uses: actions/upload-artifact@v3
        with:
          name: binaries
          path: adk-code-*
          retention-days: 5
```

### 4. Release Workflow (`release.yml`)

**Trigger:** On version tag creation (e.g., `v1.2.3`)

**Stages:**

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build-release:
    runs-on: ubuntu-latest
    name: Build Release Binaries
    strategy:
      matrix:
        include:
          - { goos: linux, goarch: amd64, suffix: '' }
          - { goos: linux, goarch: arm64, suffix: '' }
          - { goos: linux, goarch: arm, goarm: 7, suffix: '' }
          - { goos: darwin, goarch: amd64, suffix: '' }
          - { goos: darwin, goarch: arm64, suffix: '' }
          - { goos: windows, goarch: amd64, suffix: '.exe' }
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
          cache: true
      
      - name: Build release binary
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
          GOARM: ${{ matrix.goarm || '0' }}
        run: |
          VERSION=${GITHUB_REF#refs/tags/}
          BINARY="adk-code-${VERSION}-${{ matrix.goos }}-${{ matrix.goarch }}${{ matrix.suffix }}"
          
          go build \
            -ldflags="-s -w -X adk-code/internal/app.AppVersion=${VERSION}" \
            -o "${BINARY}" .
          
          sha256sum "${BINARY}" > "${BINARY}.sha256"
          ls -lh "${BINARY}"*

      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: release-${{ matrix.goos }}-${{ matrix.goarch }}
          path: adk-code-*
          retention-days: 1

  create-release:
    runs-on: ubuntu-latest
    needs: build-release
    name: Create GitHub Release
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
      
      - name: Download all artifacts
        uses: actions/download-artifact@v3
        with:
          path: release-artifacts
      
      - name: Prepare release assets
        run: |
          mkdir -p release
          find release-artifacts -type f -name 'adk-code-*' -exec cp {} release/ \;
          cd release
          ls -lh
      
      - name: Generate release notes
        id: notes
        run: |
          VERSION=${GITHUB_REF#refs/tags/}
          echo "version=${VERSION}" >> $GITHUB_OUTPUT
          echo "## Release ${VERSION}" > RELEASE_NOTES.md
          echo "" >> RELEASE_NOTES.md
          echo "### Changes" >> RELEASE_NOTES.md
          git log $(git describe --tags --abbrev=0 2>/dev/null || git rev-list --all | tail -1)..HEAD --oneline >> RELEASE_NOTES.md 2>/dev/null || echo "Initial release" >> RELEASE_NOTES.md
      
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: release/*
          body_path: RELEASE_NOTES.md
          draft: false
          prerelease: false
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### 5. Makefile Integration

Extend existing Makefile with CI/CD-aware targets:

```makefile
## cross-build: Build for all supported platforms
cross-build:
	@echo "$(GREEN)🔨 Building for all platforms...$(NC)"
	@mkdir -p dist
	@bash scripts/build-release.sh
	@echo "$(GREEN)✓ Cross-platform build complete$(NC)"
	@ls -lh dist/

## dist-clean: Remove dist directory
dist-clean:
	@rm -rf dist/
	@echo "$(GREEN)✓ Cleaned dist directory$(NC)"

## ci-check: Run all CI checks locally
ci-check: fmt vet lint test
	@echo "$(GREEN)✓ All CI checks passed$(NC)"
```

### 6. Version Management Strategy

**Semantic Versioning:** `MAJOR.MINOR.PATCH.BUILD`

**Rules:**
- `MAJOR`: Breaking API changes
- `MINOR`: New features (backward compatible)
- `PATCH`: Bug fixes
- `BUILD`: Auto-incremented on local builds (for development)

**Tags for releases:** Only `v{MAJOR}.{MINOR}.{PATCH}` (no BUILD component)

**Implementation:**
- `scripts/version.sh` manages `.version` file
- CI/CD uses exact tag version for releases
- Local builds auto-increment BUILD number

### 7. Dependency Management

**Best Practices:**
- `go mod tidy`: Keep dependencies clean
- `go mod verify`: Verify integrity (run in CI)
- GitHub Dependabot: Automated security updates
- SCA (Software Composition Analysis): Optional via GitHub Advanced Security

**Tools:**
- `go list -m all`: List all dependencies
- `go mod graph`: Visualize dependency graph
- `govulncheck`: Scan for known vulnerabilities (run in CI)

### 8. Build Optimization

**Release Build Flags:**
```bash
-ldflags="-s -w"  # Strip symbols and debug info for smaller binary
-trimpath          # Remove filesystem paths for reproducibility
```

**Optional advanced optimizations:**
- Profile-guided optimization (PGO) - Go 1.20+
- Link-time optimization - Requires LTO-capable compiler
- Static linking - Use `-tags netgo` for pure Go DNS

### 9. Testing Strategy

**Test Coverage Requirements:**
- Minimum 70% code coverage for CI pass
- Race condition detection (`-race` flag) on all tests
- Benchmark tests for performance-critical code
- Integration tests in `internal/` package tree

**Test environments:**
- Linux (primary): ubuntu-latest runner
- Windows: windows-latest runner (optional for smoke tests)
- macOS: macos-latest runner (optional for ARM testing)

### 10. Security Scanning

**Built into CI:**
- `gosec`: Check for common Go security issues
- `govulncheck`: Scan for known vulnerabilities in dependencies
- SARIF upload to GitHub Code Scanning
- Optional: Sonarqube for advanced SAST

### 11. Artifact Management

**Build Artifacts (CI):**
- Retained for 5 days
- Used for testing and verification
- Deleted automatically after retention period

**Release Artifacts (Releases):**
- SHA256 checksums for each binary
- Digital signatures (future: GPG or cosign)
- SBOM (Software Bill of Materials) generation (future)

### 12. Distribution Channels (Future)

**Planned:**
1. GitHub Releases (primary)
2. Homebrew Formula (macOS)
3. APT/YUM repositories (Linux)
4. Scoop manifest (Windows)
5. Direct binary downloads from website

## Consequences

### Positive

1. **Quality Assurance**: Automated testing ensures code quality before merge
2. **Multi-Platform Support**: Users can install binaries for their OS/architecture
3. **Release Automation**: Reduces manual work and human error
4. **Reproducibility**: Every build is traceable to specific commit
5. **Security**: Automated vulnerability scanning and dependency management
6. **Professional Operations**: Scales with project growth
7. **Community Contributions**: Clear CI/CD process enables confident contributions

### Negative/Challenges

1. **Maintenance Overhead**: CI/CD pipeline requires monitoring and updates
2. **GitHub Actions Quota**: Free tier has limits (may require paid plan if heavy usage)
3. **Build Time**: Cross-platform compilation adds ~5-10 minutes per CI run
4. **Complexity**: Multiple workflows and configurations to maintain
5. **Debugging CI Failures**: Remote environment requires different debugging approach

### Mitigation

- **Monitoring**: Set up GitHub Actions usage alerts
- **Caching**: Use Go module caching to speed up builds
- **Parallelization**: Matrix builds run simultaneously
- **Documentation**: Maintain workflow documentation in ADR and wiki
- **Local Testing**: Test CI steps locally with `act` tool before pushing

## Implementation Plan

### Phase 1 (Weeks 1-2): Foundation
- [ ] Create `.github/workflows/ci.yml` with quality + test stages
- [ ] Update Makefile with `ci-check` target
- [ ] Create `scripts/build-release.sh`
- [ ] Test locally with `act` tool

### Phase 2 (Weeks 3-4): Release Automation
- [ ] Create `.github/workflows/release.yml`
- [ ] Add release tag validation
- [ ] Test manual tag push to staging
- [ ] Document release process

### Phase 3 (Weeks 5-6): Distribution
- [ ] Create Homebrew formula template
- [ ] Create APT/YUM repository setup (or use separate service)
- [ ] Set up package manager CI integrations
- [ ] Document installation methods in README

### Phase 4 (Ongoing): Monitoring & Refinement
- [ ] Monitor CI/CD performance metrics
- [ ] Collect user feedback on installation experience
- [ ] Add new platforms as requested
- [ ] Implement security scanning enhancements

## Related Documents

- See `ADR 0004: Distribution Channels (Phase 3)` for package manager distribution
- See `TOOL_DEVELOPMENT.md` for tool development and testing patterns
- See `ARCHITECTURE.md` for system design and component interactions
- See `QUICK_REFERENCE.md` for CLI usage reference

## References

- [Go Cross-Compilation](https://golang.org/doc/install/source#environment)
- [GitHub Actions Workflow Syntax](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)
- [Go Release Checklists](https://golang.org/doc/devel/release)
- [Semantic Versioning](https://semver.org/)
- [SBOM Best Practices](https://cyclonedx.org/specification/overview/)

## Open Questions

1. Should we support additional architectures like `ppc64le`, `s390x`?
2. Should we implement GPG signing for released binaries?
3. Should we auto-publish to package managers or require manual approval?
4. What is the minimum Go version we should support (1.24 required currently)?

## Decision Record

- **Date**: November 14, 2025
- **Proposer**: @raphaelmansuy
- **Status**: Proposed for review
- **Next Steps**: Team review and approval, then Phase 1 implementation

