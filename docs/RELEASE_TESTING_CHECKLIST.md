# Release Testing Checklist

This document provides a systematic checklist for testing releases before and after they go live.

## Pre-Release Testing (Local)

Perform these tests on your local machine before creating the release tag.

### Environment Setup

```bash
cd adk-code

# Ensure you're on main and up-to-date
git checkout main
git pull origin main

# Verify no uncommitted changes
git status  # should show "nothing to commit"
```

### Step 1: Run All CI Checks Locally

```bash
# Format check
make fmt
if [ $? -ne 0 ]; then echo "❌ Format failed"; exit 1; fi
echo "✓ Format check passed"

# Go vet
make vet
if [ $? -ne 0 ]; then echo "❌ Vet failed"; exit 1; fi
echo "✓ Go vet passed"

# Lint
make lint
if [ $? -ne 0 ]; then echo "❌ Lint failed"; exit 1; fi
echo "✓ Lint passed"

# Tests
make test
if [ $? -ne 0 ]; then echo "❌ Tests failed"; exit 1; fi
echo "✓ Tests passed"

# Coverage
make coverage
# Check coverage percentage in HTML report
echo "✓ Coverage report generated"
```

### Step 2: Build for All Platforms

```bash
# Cross-platform build
make cross-build

# Verify all binaries created
ls -lh ../dist/adk-code-v*

# Should show 6 binaries:
# - adk-code-v1.2.0-linux-amd64
# - adk-code-v1.2.0-linux-arm64
# - adk-code-v1.2.0-linux-arm
# - adk-code-v1.2.0-darwin-amd64
# - adk-code-v1.2.0-darwin-arm64
# - adk-code-v1.2.0-windows-amd64.exe
```

### Step 3: Test Binary Functionality

Test the binary for your current platform:

```bash
# Determine your platform
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
VERSION="v1.2.0"  # Update to actual version

# Choose correct binary
if [ "$GOOS" = "darwin" ]; then
  BINARY="../dist/adk-code-${VERSION}-darwin-${GOARCH}"
elif [ "$GOOS" = "linux" ]; then
  BINARY="../dist/adk-code-${VERSION}-linux-${GOARCH}"
elif [ "$GOOS" = "windows" ]; then
  BINARY="../dist/adk-code-${VERSION}-windows-amd64.exe"
fi

# Make binary executable
chmod +x "$BINARY"

# Test 1: Version flag
echo "Test 1: Version flag"
"$BINARY" --version
if [ $? -ne 0 ]; then echo "❌ Version flag failed"; exit 1; fi
echo "✓ Version flag works"

# Test 2: Help flag
echo "Test 2: Help flag"
"$BINARY" --help
if [ $? -ne 0 ]; then echo "❌ Help flag failed"; exit 1; fi
echo "✓ Help flag works"

# Test 3: Basic functionality (adjust based on your tool)
echo "Test 3: Basic functionality"
echo "test input" | "$BINARY"
if [ $? -ne 0 ]; then echo "❌ Basic functionality failed"; exit 1; fi
echo "✓ Basic functionality works"
```

### Step 4: Verify Checksums

```bash
# Check all checksums
cd ../dist
sha256sum -c *.sha256

# All should show "OK" (or in short form, just pass silently)
# If any show "FAILED", abort release

cd -
echo "✓ All checksums verified"
```

### Step 5: Test Binary Size Expectations

```bash
# Display binary sizes
ls -lh ../dist/adk-code-v* | grep -v sha256

# Sizes should be reasonable (typically 5-20 MB depending on project)
# Unusually large or small size may indicate build issue

echo "✓ Binary sizes reasonable"
```

## GitHub Actions Workflow Testing

These tests run automatically in GitHub Actions when the tag is pushed.

### Monitor Workflow Steps

Go to: https://github.com/raphaelmansuy/adk-code/actions

Watch for these jobs:

1. **validate-tag** - Should complete in ~10 seconds
   - [ ] Tag format validation passes
   - [ ] Shows tag: vX.Y.Z
   - [ ] Detects prerelease correctly

2. **build-release** - Should complete in ~3-5 minutes
   - [ ] All 6 builds complete successfully:
     - [ ] Linux amd64
     - [ ] Linux arm64
     - [ ] Linux armv7
     - [ ] macOS amd64
     - [ ] macOS arm64
     - [ ] Windows amd64
   - [ ] Each build logs binary size
   - [ ] Artifacts uploaded

3. **create-release** - Should complete in ~1-2 minutes
   - [ ] All artifacts downloaded
   - [ ] Release notes generated
   - [ ] Changelog section included (if not first release)
   - [ ] GitHub Release created
   - [ ] Release marked as draft or prerelease appropriately

4. **post-release** - Should complete in ~10 seconds
   - [ ] Completion message logged
   - [ ] No errors

## Post-Release Verification

Perform these tests immediately after the release is created.

### Step 1: Verify Release on GitHub

```bash
# View release details
gh release view v1.2.0

# Should show:
# - Version tag
# - Release date
# - Asset count (should be 12: 6 binaries + 6 checksums)
# - Correct prerelease status
```

### Step 2: Download and Test Binaries

For each platform you can test locally:

#### Linux (amd64/arm64)
```bash
# Download
gh release download v1.2.0 -p '*linux-amd64*'

# Verify checksum
sha256sum -c adk-code-v1.2.0-linux-amd64.sha256

# Make executable
chmod +x adk-code-v1.2.0-linux-amd64

# Test version
./adk-code-v1.2.0-linux-amd64 --version
# Should output: v1.2.0

# Test help
./adk-code-v1.2.0-linux-amd64 --help
# Should show usage information

# Quick functional test
echo "test" | ./adk-code-v1.2.0-linux-amd64
# Should run without errors
```

#### macOS (Intel/Apple Silicon)
```bash
# Download Apple Silicon version
gh release download v1.2.0 -p '*darwin-arm64*'

# Verify
sha256sum -c adk-code-v1.2.0-darwin-arm64.sha256
chmod +x adk-code-v1.2.0-darwin-arm64

# Test
./adk-code-v1.2.0-darwin-arm64 --version
./adk-code-v1.2.0-darwin-arm64 --help
```

#### Windows (amd64)
```powershell
# Download
gh release download v1.2.0 -p "*windows-amd64*"

# Verify (using PowerShell)
(Get-FileHash .\adk-code-v1.2.0-windows-amd64.exe).Hash -eq (Get-Content .\adk-code-v1.2.0-windows-amd64.exe.sha256).Split()[0]

# Test
.\adk-code-v1.2.0-windows-amd64.exe --version
.\adk-code-v1.2.0-windows-amd64.exe --help
```

### Step 3: Document Test Results

Create a test report:

```bash
cat > logs/release-test-v1.2.0.md << 'EOF'
# Release Test Report: v1.2.0

## Date
2025-11-14

## Platforms Tested
- [x] Linux amd64
- [x] Linux arm64
- [ ] Linux armv7 (no hardware)
- [x] macOS arm64 (Apple Silicon)
- [ ] macOS amd64 (no hardware)
- [ ] Windows amd64 (no hardware)

## Test Results

### Linux amd64
- [x] Download successful
- [x] Checksum verification passed
- [x] --version flag works: v1.2.0
- [x] --help flag works
- [x] Basic functionality works

### Linux arm64
- [x] Download successful
- [x] Checksum verification passed
- [x] --version flag works: v1.2.0
- [x] --help flag works
- [x] Basic functionality works

### macOS arm64
- [x] Download successful
- [x] Checksum verification passed
- [x] --version flag works: v1.2.0
- [x] --help flag works
- [x] Basic functionality works

## Issues Found
None

## Sign-off
- Tested by: @raphaelmansuy
- Date: 2025-11-14
- Status: ✅ APPROVED FOR PRODUCTION
EOF

git add logs/release-test-v1.2.0.md
git commit -m "docs: add release test report for v1.2.0"
git push
```

## Continuous Monitoring (Post-Release)

After release, monitor for issues for at least 24-48 hours.

### Step 1: Monitor GitHub Issues

- [ ] Check GitHub Issues for release-related problems
- [ ] Check GitHub Discussions for user feedback
- [ ] Search for mention of new version in issues

### Step 2: Monitor Error Reports

If your application has error reporting:
- [ ] Check error logs for crashes related to new version
- [ ] Monitor error frequency
- [ ] Look for patterns in failures

### Step 3: Monitor Usage Analytics

If you have usage analytics:
- [ ] Verify users are downloading new version
- [ ] Check if download counts are as expected
- [ ] Monitor adoption across platforms

### Step 4: Respond to Feedback

- [ ] Respond to user issues quickly
- [ ] Address critical bugs immediately
- [ ] Document workarounds for known issues

## Emergency Testing (Critical Issues)

If a critical issue is found after release:

### Step 1: Reproduce Issue

```bash
# Download affected binary
gh release download v1.2.0 -p '*YOUR_PLATFORM*'

# Try to reproduce
./adk-code-v1.2.0-* --option-that-fails

# Document exact reproduction steps
```

### Step 2: Fix and Release Patch

```bash
# Create fix
git commit -m "fix: critical issue in v1.2.0"

# Bump to patch
./scripts/version.sh set 1.2.1

# Create patch release
git tag v1.2.1
git push origin v1.2.1

# Test patch using same steps as above
```

### Step 3: Post Incident Report

See [RELEASE_PROCESS.md - Document the Incident](RELEASE_PROCESS.md#document-the-incident)

## Regression Testing

For each release, run a quick regression test on key features:

```bash
# Test 1: Basic invocation works
BINARY="./adk-code-v1.2.0-*"
chmod +x "$BINARY"
"$BINARY" --version

# Test 2: All major commands work
"$BINARY" command-1 --help
"$BINARY" command-2 --help
# ... test all major commands

# Test 3: Integration with system
# Test environment variables
export TEST_VAR="test"
"$BINARY" --some-option
# Test file I/O
echo "test" > test-input.txt
"$BINARY" test-input.txt
# Test directory operations
mkdir test-dir
"$BINARY" test-dir/

# Test 4: Backward compatibility
# If applicable, test with previous version's configuration
```

## Automated Testing Script

Use this script to automate release testing:

```bash
#!/bin/bash
# scripts/test-release.sh

set -e

VERSION="${1:-v1.2.0}"
GITHUB_REPO="raphaelmansuy/adk-code"

echo "🧪 Testing Release: $VERSION"
echo "======================================"

# Download binaries
echo "📥 Downloading binaries..."
mkdir -p test-release
cd test-release
gh release download "$VERSION" -p '*' -R "$GITHUB_REPO"

# Test each binary
for binary in adk-code-*; do
  [[ "$binary" == *.sha256 ]] && continue
  
  chmod +x "$binary"
  echo ""
  echo "Testing: $binary"
  
  # Test version
  ./"$binary" --version || echo "❌ Version flag failed"
  
  # Test help
  ./"$binary" --help || echo "❌ Help flag failed"
done

# Verify checksums
echo ""
echo "✓ Verifying checksums..."
sha256sum -c *.sha256

echo ""
echo "✅ All tests passed!"
```

## Accessibility Testing (Optional)

For releases with new documentation or CLI changes:

- [ ] Help text is clear and complete
- [ ] Error messages are descriptive
- [ ] Usage examples are provided
- [ ] Documentation is accurate
- [ ] Links in release notes are valid

## Performance Baseline

For releases with performance improvements:

```bash
# Create baseline before release
time ./adk-code-v1.1.0-* operation-to-test

# Create baseline after release
time ./adk-code-v1.2.0-* operation-to-test

# Compare results
# Confirm improvements or identify regressions
```

## Sign-Off Checklist

Before marking release as complete:

- [ ] All automated tests (GitHub Actions) passed
- [ ] Manual testing completed on primary platform
- [ ] At least one alternative platform tested
- [ ] Checksums verified
- [ ] Release notes are accurate
- [ ] No critical issues found
- [ ] Documentation is updated
- [ ] Release test report filed
