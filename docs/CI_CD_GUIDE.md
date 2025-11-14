# CI/CD Implementation Guide

This guide explains how to set up and use the CI/CD pipeline for adk-code.

## Overview

The adk-code project uses GitHub Actions for continuous integration and continuous deployment. The pipeline includes:

- **Code Quality Checks**: Formatting, linting, and static analysis
- **Security Scanning**: Vulnerability scanning and security policy checks
- **Automated Testing**: Unit tests with race detection and coverage reporting
- **Multi-Platform Builds**: Automated binary compilation for 6 platform/architecture combinations
- **Release Automation**: Automatic GitHub release creation with checksums

## Files Reference

| File | Purpose |
|------|---------|
| `.github/workflows/ci.yml` | Primary CI pipeline (format, lint, test, build) |
| `.github/workflows/release.yml` | Release automation workflow |
| `scripts/build-release.sh` | Cross-platform build script |
| `scripts/version.sh` | Version management utility |
| `Makefile` | Local development targets |

## Quick Start

### Running CI Locally

To run the same checks that CI runs:

```bash
cd adk-code/
make ci-check
```

This runs:
1. `make fmt` - Code formatting
2. `make vet` - Go vet analysis
3. `make lint` - Linter checks
4. `make test` - Unit tests

### Building for All Platforms

```bash
cd adk-code/
./scripts/build-release.sh
```

This creates binaries in `../dist/` for all supported platforms.

### Building for Specific Platform

```bash
cd adk-code/

# Linux amd64
GOOS=linux GOARCH=amd64 make release

# macOS arm64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 make release

# Windows amd64
GOOS=windows GOARCH=amd64 make release
```

## Workflow Details

### CI Workflow (ci.yml)

Runs on every push and pull request to `main` or `develop` branches.

#### Jobs:

1. **format** - Checks code formatting with `gofmt`
   - Fails if files don't match Go format style
   - Fix with: `go fmt ./...`

2. **vet** - Runs `go vet` for suspicious code patterns
   - Checks for common mistakes like incorrect printf format strings
   - Fix issues based on error messages

3. **lint** - Runs `golangci-lint` for comprehensive linting
   - Checks code style, potential bugs, performance issues
   - Configure in `.golangci.yml` if needed

4. **security** - Runs security scanners
   - `gosec`: Detects security issues (SQL injection, hardcoded secrets, etc.)
   - `govulncheck`: Scans for known vulnerabilities in dependencies
   - Results uploaded to GitHub Security tab

5. **test** - Runs unit tests with race detection
   - Executes all `*_test.go` files
   - Detects data races with `-race` flag
   - Collects coverage and fails if below 70%
   - Uploads coverage to Codecov

6. **build** - Builds binaries for 6 platform/architecture combinations
   - Runs in parallel after all checks pass
   - Uses Go cross-compilation
   - Artifacts retained for 5 days

7. **ci-complete** - Final check that all jobs passed

#### Supported Platforms:

| OS | Architecture | Notes |
|----|--------------|-------|
| Linux | amd64 | Primary server platform |
| Linux | arm64 | Raspberry Pi 5, AWS Graviton |
| Linux | armv7 | Older ARM boards |
| macOS | amd64 | Intel Macs |
| macOS | arm64 | Apple Silicon (M1/M2/M3) |
| Windows | amd64 | x86-64 Windows |

### Release Workflow (release.yml)

Runs automatically when a version tag is pushed (e.g., `git tag v1.2.3`).

#### Jobs:

1. **validate-tag** - Validates tag format
   - Must match: `v1.2.3` or `v1.2.3-rc1`
   - Detects prerelease versions (containing -alpha, -beta, -rc)

2. **build-release** - Builds release binaries
   - Same 6 platforms as CI
   - Generates SHA256 checksums
   - Uploads artifacts

3. **create-release** - Creates GitHub Release
   - Generates release notes with checksums
   - Includes changelog (commits since last release)
   - Attaches all binaries
   - Marks as prerelease if tag contains `-alpha`, `-beta`, `-rc`

4. **post-release** - Post-release tasks (future)
   - Currently just logs completion
   - Future: update docs, publish to package managers

## How to Create a Release

> **Note:** For a comprehensive step-by-step release process, see [RELEASE_PROCESS.md](RELEASE_PROCESS.md).
> For detailed testing procedures, see [RELEASE_TESTING_CHECKLIST.md](RELEASE_TESTING_CHECKLIST.md).

The release process has four main phases:

### Phase 1: Pre-Release Preparation

1. Complete and merge all features for the release to `main`
2. Run pre-release checks locally: `make ci-check`
3. Verify GitHub Actions pipeline passes on `main` branch
4. Update documentation (README, CHANGELOG)
5. Update version file: `./scripts/version.sh set 1.2.0`

See [RELEASE_PROCESS.md - Pre-Release Preparation](RELEASE_PROCESS.md#pre-release-preparation) for detailed steps.

### Phase 2: Create Release Tag

Once all checks pass, create and push the release tag:

```bash
cd adk-code

# Verify version
./scripts/version.sh get
# Should output: 1.2.0

# Create annotated tag (format: vX.Y.Z)
git tag -a v1.2.0 -m "Release v1.2.0

New features:
- Feature 1
- Feature 2

Bug fixes:
- Fix 1"

# Push tag to GitHub (triggers Release workflow)
git push origin v1.2.0
```

### Phase 3: Monitor Automated Release

Go to GitHub Actions → Release workflow and watch completion:

The workflow automatically:

1. **validate-tag** - Verifies tag format (v1.2.0 or v1.2.0-rc1)
2. **build-release** - Builds all 6 platform binaries in parallel
3. **create-release** - Creates GitHub Release with all assets and checksums
4. **post-release** - Logs completion

Typical time: 3-5 minutes

### Phase 4: Post-Release Verification

After workflow completes:

1. **Verify release on GitHub:**

```bash
gh release view v1.2.0
# Should show 12 assets (6 binaries + 6 checksums)
```

2. **Download and test binaries:**

```bash
gh release download v1.2.0 -p '*linux-amd64*'
sha256sum -c adk-code-v1.2.0-linux-amd64.sha256
./adk-code-v1.2.0-linux-amd64 --version
```

3. **Test alternative platforms** if possible (see [RELEASE_TESTING_CHECKLIST.md](RELEASE_TESTING_CHECKLIST.md))

4. **Announce release** to community (GitHub Discussions, social media, etc.)

See [RELEASE_PROCESS.md - Release Execution](RELEASE_PROCESS.md#release-execution) for comprehensive instructions.

### Future: Publish to Package Managers

Planned distributions (Phase 3):

- Homebrew (macOS)
- APT/YUM (Linux)
- Scoop (Windows)
- And more

## Troubleshooting

### CI fails on lint

**Problem:** `golangci-lint` fails

**Solution:**
```bash
# See detailed lint errors
cd adk-code
make lint

# Fix issues (many auto-fixable)
# Then commit and push
```

### CI fails on format

**Problem:** `gofmt` fails

**Solution:**
```bash
cd adk-code
go fmt ./...
git add .
git commit -m "chore: format code"
git push
```

### Test coverage below 70%

**Problem:** Coverage check fails

**Solution:**
```bash
cd adk-code
make coverage

# Opens coverage.html - shows which lines lack coverage
# Add tests for uncovered code
```

### Build fails for specific platform

**Problem:** Build fails for `linux/arm`

**Solution:**
1. Check build logs in GitHub Actions
2. Try building locally: `GOOS=linux GOARCH=arm GOARM=7 go build .`
3. Common issues:
   - Cgo dependencies (avoid for maximum portability)
   - Platform-specific syscalls (use build constraints)
   - Check error message for specifics

### Release tag is wrong

**Problem:** Tagged wrong version like `1.2.3` instead of `v1.2.3`

**Solution:**
```bash
# Delete wrong tag
git tag -d 1.2.3
git push origin :refs/tags/1.2.3

# Delete GitHub release if created
# Create correct tag
git tag v1.2.3
git push origin v1.2.3
```

## Best Practices

### 1. Commit Messages

Follow conventional commits for automatic changelog generation:
- `feat: Add new feature` - New feature
- `fix: Fix bug in X` - Bug fix
- `docs: Update README` - Documentation
- `chore: Update deps` - Chores/maintenance
- `test: Add tests for X` - Test additions

### 2. Pull Requests

- Keep PRs focused on one feature/fix
- Write clear description
- Ensure CI passes before requesting review
- Require approvals before merging (set in GitHub settings)

### 3. Versioning

Use semantic versioning: `MAJOR.MINOR.PATCH`

- `1.0.0` - Initial release
- `1.1.0` - New feature (backward compatible)
- `1.1.1` - Bug fix
- `2.0.0` - Breaking change

### 4. Releases

- Create release only from `main` branch
- Tag format: `v1.2.3` (with leading `v`)
- Prerelease format: `v1.2.3-rc1` (release candidates)
- One release per version

### 5. Testing

- Add tests for new features
- Maintain >70% code coverage
- Use `t.Run()` for subtests
- Test both success and error cases
- Use table-driven tests for multiple cases

## Monitoring

### GitHub Actions Dashboard

View CI/CD history and performance:
1. Go to your repository
2. Click "Actions" tab
3. See workflow runs and logs

### Coverage Reports

Coverage is uploaded to Codecov:
- View detailed coverage reports
- Track coverage over time
- Set coverage decrease alerts

### Security Alerts

GitHub automatically monitors:
- Dependency vulnerabilities (Dependabot)
- Code scanning results (CodeQL)
- Security policies (in GitHub Security tab)

## Advanced Configuration

### Customizing Platforms

To add a new platform (e.g., FreeBSD):

1. Update `.github/workflows/ci.yml`:
   ```yaml
   - { goos: freebsd, goarch: amd64, name: "FreeBSD x86-64" }
   ```

2. Update `.github/workflows/release.yml` similarly

3. Test locally first:
   ```bash
   GOOS=freebsd GOARCH=amd64 go build .
   ```

### Skipping CI

To skip CI for a specific commit:

```bash
git commit --message "docs: update README [skip ci]"
git push
```

### Running Workflows Manually

From GitHub Actions tab:
1. Click workflow name
2. Click "Run workflow" button
3. Select branch
4. Click green "Run workflow" button

## Security Considerations

### Secrets Management

**Never commit secrets!** Use GitHub Secrets for:
- API keys
- Tokens
- Credentials

Configure in: Settings → Secrets and variables → Actions

### Building Dependencies

All dependencies are checksummed in `go.sum`. Verify:

```bash
go mod verify
```

### Release Signing (Future)

Plan to add:
- GPG signing of releases
- Cosign for container images
- SBOM (Software Bill of Materials)

## Performance

### Caching

Go modules are cached automatically:
- Build cache: `~/.cache/go-build` (local)
- Module cache: `~/go/pkg/mod` (local)
- GitHub Actions uses built-in caching

### Build Time

Typical CI run times:
- Format/Vet/Lint: ~30 seconds
- Tests: ~45 seconds
- Builds (6 platforms, parallel): ~60 seconds
- **Total: ~2-3 minutes**

### Optimization Tips

1. Use `-v` flag in build for detailed output
2. Check `GOMODCACHE` environment variable
3. Run `go mod download` to pre-fetch dependencies
4. Use `-trimpath` for reproducible builds

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Build Constraints](https://pkg.go.dev/cmd/go#hdr-Build_constraints)
- [Go Cross-Compilation](https://golang.org/doc/install/source#environment)
- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)

## Support

For issues or questions:
1. Check this guide
2. Review GitHub Actions logs for error messages
3. Open an issue on GitHub with CI/CD tag
4. Ask in project discussions

