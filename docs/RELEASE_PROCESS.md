# Release Process Guide

This document describes the complete release process for adk-code, from preparation through post-release tasks.

## Table of Contents

1. [Pre-Release Preparation](#pre-release-preparation)
2. [Release Execution](#release-execution)
3. [Post-Release Verification](#post-release-verification)
4. [Troubleshooting](#troubleshooting)
5. [Emergency Rollback](#emergency-rollback)

## Pre-Release Preparation

### Step 1: Feature Completion and Merge

Ensure all features for the release are complete and merged to `main`:

```bash
# Create a feature branch
git checkout -b feature/new-feature
git commit -m "feat: implement new feature"

# Push and create PR
git push origin feature/new-feature

# Get review approval, then merge to main
```

### Step 2: Run Pre-Release Checks

Before creating a release, verify everything works:

```bash
cd adk-code

# Run all CI checks locally
make ci-check

# Verify tests pass
make test

# Check code coverage
make coverage

# Build for all platforms locally
make cross-build

# Verify all binaries were created
ls -lh ../dist/adk-code-*
```

### Step 3: Verify CI/CD Pipeline

Push any changes to `main` and verify GitHub Actions passes:

```bash
git push origin main

# Go to GitHub Actions tab and wait for all checks to pass
# All of these must pass:
# - Format
# - Vet
# - Lint
# - Tests
# - Build (6 platforms)
```

### Step 4: Update Documentation

Update relevant documentation with new features/changes:

- Update `README.md` with new features or significant changes
- Update `CHANGELOG.md` with release notes
- Update tool documentation if tools were added/modified
- Update `docs/QUICK_REFERENCE.md` if CLI changes

Example CHANGELOG entry:

```markdown
## [1.2.0] - 2025-11-14

### Added
- New feature: improved code search
- Support for additional model providers

### Fixed
- Bug in terminal pagination
- Issue with workspace path resolution

### Changed
- Updated dependencies to latest versions
```

### Step 5: Update Version File

Update the version file to match the planned release:

```bash
cd adk-code

# Current version
./scripts/version.sh get
# Output: 1.1.5.12

# Set to release version (without build number)
./scripts/version.sh set 1.2.0

# Verify
./scripts/version.sh get
# Output: 1.2.0
```

### Step 6: Create Release Branch (Optional)

For major releases, optionally create a release branch:

```bash
git checkout -b release/v1.2.0
```

This branch can be used for last-minute fixes before release.

### Step 7: Commit Version Update

Commit the version change:

```bash
git add .version
git commit -m "chore: bump version to 1.2.0 for release"
git push origin main
# Or push to release branch if created
```

## Release Execution

### Step 1: Create Release Tag

Once all checks pass and documentation is updated, create the release tag:

```bash
# Ensure you're on main branch and up-to-date
git checkout main
git pull origin main

# Verify version
./scripts/version.sh get
# Should output: 1.2.0

# Create annotated tag with release notes
git tag -a v1.2.0 -m "Release v1.2.0

Major Features:
- Feature 1 description
- Feature 2 description

Bug Fixes:
- Fix 1 description

Contributors:
- @contributor1
- @contributor2"

# Verify tag was created
git tag -l | grep v1.2.0

# Show tag details
git show v1.2.0
```

### Step 2: Push Tag to GitHub

Push the tag to trigger the Release workflow:

```bash
git push origin v1.2.0
```

### Step 3: Monitor Release Workflow

Watch the automated release process:

```bash
# Option 1: GitHub Web UI
# Go to: https://github.com/raphaelmansuy/adk-code/actions
# Click on "Release" workflow
# Watch until all jobs complete (usually 3-5 minutes)

# Option 2: GitHub CLI
gh run list --workflow release.yml | head -10
gh run watch <run-id>
```

The Release workflow will:
1. **validate-tag** - Verify tag format is correct (v1.2.0 or v1.2.0-rc1)
2. **build-release** - Build binaries for 6 platforms in parallel
3. **create-release** - Create GitHub Release with:
   - All platform-specific binaries
   - SHA256 checksums
   - Release notes with changelog
   - Marked as prerelease if tag contains `-alpha`, `-beta`, `-rc`
4. **post-release** - Log completion

### Step 4: Verify Release Created

Once the workflow completes, verify the release:

```bash
# View release on GitHub
# https://github.com/raphaelmansuy/adk-code/releases/tag/v1.2.0

# Or use GitHub CLI
gh release view v1.2.0

# Should show:
# - Version tag
# - Release notes
# - 6 binary downloads
# - 6 SHA256 checksum files
```

## Post-Release Verification

### Step 1: Download and Test Binaries

Test binaries for your platform:

```bash
# Download latest release
gh release download v1.2.0 -p '*linux-amd64'

# Verify checksum
sha256sum -c adk-code-v1.2.0-linux-amd64.sha256

# Test the binary
./adk-code-v1.2.0-linux-amd64 --version
# Should output: v1.2.0

# Test basic functionality
echo "test" | ./adk-code-v1.2.0-linux-amd64 --help
```

### Step 2: Test Multiple Platforms (if possible)

If you have access to multiple platforms, test at least one additional platform:

```bash
# macOS
gh release download v1.2.0 -p '*darwin-arm64'
./adk-code-v1.2.0-darwin-arm64 --version

# Windows (if using Windows)
gh release download v1.2.0 -p '*windows-amd64.exe'
.\adk-code-v1.2.0-windows-amd64.exe --version
```

### Step 3: Announce Release

Share the release with the community:

1. **Update GitHub Discussions** - Post in Announcements
2. **Social Media** - Tweet/share the release
3. **Email/Newsletter** - If applicable
4. **Documentation Site** - Update release page if exists

Example announcement:

```
🎉 v1.2.0 is here!

Download for your platform:
- macOS: Intel / Apple Silicon
- Linux: x86-64 / ARM64 / ARMv7
- Windows: x86-64

New features:
- [Feature 1]
- [Feature 2]

Download: https://github.com/raphaelmansuy/adk-code/releases/v1.2.0
```

### Step 4: Create GitHub Discussion Post

Post a discussion for users to provide feedback:

```bash
# Or create manually in GitHub UI
# Go to Discussions → New discussion
# Tag: Release Announcement
# Title: v1.2.0 Released
# Body: List features, link to release
```

## Troubleshooting

### Release Workflow Fails at validate-tag

**Error:** "Invalid tag format: 1.2.0" or similar

**Cause:** Tag doesn't match required format

**Solution:**
```bash
# Delete the invalid tag
git tag -d 1.2.0
git push origin :refs/tags/1.2.0

# Create with correct format (must start with 'v')
git tag v1.2.0
git push origin v1.2.0
```

### Release Workflow Fails at build-release

**Error:** "Build failed for platform X"

**Cause:** Go build error for specific OS/architecture

**Solution:**
1. Check the workflow logs in GitHub Actions
2. Try building locally:
   ```bash
   GOOS=linux GOARCH=arm GOARM=7 go build .
   ```
3. Common causes:
   - CGO dependencies not compatible with platform
   - Platform-specific code without build constraints
   - Missing dependencies

### Release Workflow Fails at create-release

**Error:** "Failed to create release" or "Failed to upload artifacts"

**Cause:** GitHub API error or permission issue

**Solution:**
1. Verify `GITHUB_TOKEN` has sufficient permissions (should have `contents: write`)
2. Check GitHub status page
3. Retry by manually running workflow or creating release manually:
   ```bash
   # Manual release creation
   gh release create v1.2.0 dist/* --title "v1.2.0" --notes "See CHANGELOG.md"
   ```

### Release Created but Binaries Missing

**Error:** Release exists but binaries not attached

**Cause:** Artifact upload failed silently

**Solution:**
1. Rerun the workflow:
   ```bash
   # Go to Actions → Release workflow → Run workflow → Select tag
   # Or delete and recreate tag
   ```
2. Or manually upload binaries:
   ```bash
   # Build locally
   ./scripts/build-release.sh 1.2.0
   
   # Upload to existing release
   gh release upload v1.2.0 dist/*
   ```

### Checksum Verification Fails

**Error:** "sha256sum: FAILED" when verifying

**Cause:** Binary was corrupted during download or checksum file is wrong

**Solution:**
```bash
# Download fresh from GitHub
gh release download v1.2.0 -p '*linux-amd64*'

# Verify checksum again
sha256sum -c adk-code-v1.2.0-linux-amd64.sha256

# If still fails, re-download and try again
```

## Emergency Rollback

### If a Release Has Critical Issues

If a release has a critical bug that breaks functionality:

#### Option 1: Remove Release (Quick Fix)

```bash
# Delete the release (keeps the tag)
gh release delete v1.2.0 --yes

# Delete the tag
git tag -d v1.2.0
git push origin :refs/tags/v1.2.0
```

#### Option 2: Release Patch Version

If users have already downloaded, release a patch immediately:

```bash
# Fix the issue
git commit -m "fix: critical issue in v1.2.0"

# Bump to patch version
./scripts/version.sh set 1.2.1

# Create new release
git tag v1.2.1
git push origin v1.2.1

# Immediately announce patch
# "If you downloaded v1.2.0, please upgrade to v1.2.1"
```

#### Option 3: Mark as Prerelease

If you want to revoke a release without deleting:

```bash
# Edit release on GitHub UI
# Mark as "Prerelease"
# Add warning: "⚠️ This release has issues, please use vX.X.X instead"
```

### Document the Incident

Always document what went wrong:

```bash
# Create incident log
cat > logs/incident-v1.2.0.md << 'EOF'
# Incident Report: v1.2.0 Release

## Timeline
- 14:30 UTC: v1.2.0 released
- 14:45 UTC: User reported issue with X
- 15:00 UTC: Root cause identified
- 15:15 UTC: v1.2.1 patch released

## Root Cause
[Description of what went wrong]

## Fix Applied
[Description of fix]

## Prevention
[How to prevent this in future releases]

## Lessons Learned
[What we learned]
EOF

git add logs/incident-v1.2.0.md
git commit -m "docs: document v1.2.0 incident"
git push origin main
```

## Pre-Release Checklist

Use this checklist before creating each release:

```markdown
## v1.2.0 Release Checklist

### Code Quality
- [ ] All tests pass locally: `make test`
- [ ] Coverage check: `make coverage`
- [ ] Code formatting: `make fmt`
- [ ] Go vet: `make vet`
- [ ] Linting: `make lint`
- [ ] GitHub CI passes on main branch

### Documentation
- [ ] CHANGELOG.md updated
- [ ] README.md updated with new features
- [ ] API documentation updated
- [ ] Tool documentation updated
- [ ] Quick reference updated

### Version
- [ ] Version number updated in .version
- [ ] Version format: MAJOR.MINOR.PATCH
- [ ] No build numbers in release version

### Release Tag
- [ ] Tag format: vX.Y.Z
- [ ] Tag is annotated (not lightweight)
- [ ] Tag includes release notes

### Testing
- [ ] Built for all 6 platforms locally
- [ ] At least one binary tested on actual platform
- [ ] Help text verified: `./binary --help`
- [ ] Version flag works: `./binary --version`

### Post-Release
- [ ] GitHub Release created with all binaries
- [ ] Checksums verify correctly
- [ ] Announcement posted
- [ ] Issue closed/milestone marked complete
```

## Related Documents

- [CI/CD_GUIDE.md](CI_CD_GUIDE.md) - General CI/CD overview
- [QUICK_REFERENCE.md](QUICK_REFERENCE.md) - CLI usage
- [0003-cicd-and-build-process.md](adr/0003-cicd-and-build-process.md) - Architecture Decision Record

## Version Formats

### Stable Release
Format: `vX.Y.Z` (e.g., `v1.2.3`)
- X = Major version (breaking changes)
- Y = Minor version (new features)
- Z = Patch version (bug fixes)

### Pre-Release
Format: `vX.Y.Z-rc1`, `vX.Y.Z-beta1`, `vX.Y.Z-alpha1`
- Automatically marked as "prerelease" on GitHub
- Not recommended for production use
- Used for community testing

### Development Version
Format: `vX.Y.Z.BUILD` (e.g., `v1.2.0.5`)
- Locally generated by `make build`
- Not released officially
- Only for testing during development
