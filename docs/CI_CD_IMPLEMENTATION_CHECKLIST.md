# CI/CD Implementation Checklist

Use this checklist to implement the CI/CD pipeline described in ADR 0003.

## Phase 1: Foundation & Setup (Weeks 1-2)

### 1.1 Prepare Repository
- [ ] Ensure `.github/workflows/` directory exists
- [ ] Verify `scripts/` directory exists with `version.sh`
- [ ] Check that `Makefile` has all required targets
- [ ] Verify `go.mod` and `go.sum` are present and valid

### 1.2 Create CI Workflow
- [ ] Create `.github/workflows/ci.yml` with:
  - [ ] Format job (gofmt)
  - [ ] Vet job (go vet)
  - [ ] Lint job (golangci-lint)
  - [ ] Security job (gosec, govulncheck)
  - [ ] Test job (unit tests, race detection, coverage)
  - [ ] Build job (matrix for 6 platforms)
  - [ ] CI-complete job (final check)
- [ ] Test CI workflow triggers on push
- [ ] Test CI workflow triggers on pull request
- [ ] Verify workflow can be run manually

### 1.3 Create Build Script
- [ ] Create `scripts/build-release.sh`
- [ ] Make script executable: `chmod +x scripts/build-release.sh`
- [ ] Test script locally:
  ```bash
  cd adk-code/
  ./scripts/build-release.sh
  ```
- [ ] Verify all 6 binaries are created
- [ ] Check binary sizes are reasonable (~10-20 MB)
- [ ] Verify checksums are generated

### 1.4 Extend Makefile
- [ ] Add `cross-build` target
- [ ] Add `dist-clean` target
- [ ] Add `ci-check` target
- [ ] Test each new target:
  ```bash
  make ci-check
  make cross-build
  make dist-clean
  ```

### 1.5 Local Testing
- [ ] Install `act` tool for local GitHub Actions testing:
  ```bash
  # macOS
  brew install act
  ```
- [ ] Test CI workflow locally:
  ```bash
  act -j format
  act -j vet
  act -j lint
  act -j test
  act -j build
  ```
- [ ] Fix any failures
- [ ] Push changes to feature branch

### 1.6 Code Quality Setup
- [ ] Ensure `.golangci.yml` exists and is configured
- [ ] Verify `golangci-lint` is installed locally:
  ```bash
  golangci-lint --version
  ```
- [ ] Run local lint check:
  ```bash
  make lint
  ```
- [ ] Fix any linting issues
- [ ] Update `.gitignore` to include:
  ```
  dist/
  coverage.out
  coverage.html
  ```

## Phase 2: Release Automation (Weeks 3-4)

### 2.1 Create Release Workflow
- [ ] Create `.github/workflows/release.yml` with:
  - [ ] Tag validation job
  - [ ] Build release job (matrix for 6 platforms)
  - [ ] Create release job (GitHub Release)
  - [ ] Post-release job (optional)
- [ ] Update workflow to reference correct version script
- [ ] Test workflow syntax (GitHub validates automatically)

### 2.2 Test Release Workflow
- [ ] Create test tag on feature branch:
  ```bash
  git tag v0.0.0-test
  git push origin v0.0.0-test
  ```
- [ ] Monitor Actions → Release workflow
- [ ] Verify all steps complete successfully
- [ ] Check GitHub Release was created at `/releases`
- [ ] Verify all 6 binaries are attached
- [ ] Verify checksums are present
- [ ] Download a binary and test it:
  ```bash
  ./adk-code-v0.0.0-test-linux-amd64 --version
  ```
- [ ] Delete test tag and release:
  ```bash
  git tag -d v0.0.0-test
  git push origin :refs/tags/v0.0.0-test
  ```
- [ ] Delete test release from GitHub UI

### 2.3 Version Strategy Documentation
- [ ] Document versioning scheme in README
- [ ] Create CONTRIBUTING.md with release instructions
- [ ] Add version history to CHANGELOG.md
- [ ] Document semver rules

### 2.4 Release Process Documentation
- [ ] Document release checklist
- [ ] Document how to create releases
- [ ] Document how to verify released binaries
- [ ] Add troubleshooting guide

## Phase 3: Distribution Channels (Weeks 5-6)

### 3.1 GitHub Releases
- [ ] Verify releases are properly created ✓ (from Phase 2)
- [ ] Add release notes templates
- [ ] Set up automatic changelog generation in workflow

### 3.2 Homebrew (macOS)
- [ ] Create `adk-code.rb` homebrew formula:
  ```ruby
  class AdkCode < Formula
    desc "AI coding assistant CLI"
    homepage "https://github.com/raphaelmansuy/adk-code"
    url "https://github.com/raphaelmansuy/adk-code/releases/download/vVERSION/adk-code-vVERSION-darwin-arm64"
    sha256 "CHECKSUM_HERE"
    
    def install
      bin.install "adk-code-vVERSION-darwin-arm64" => "adk-code"
    end
  end
  ```
- [ ] Test locally:
  ```bash
  brew install ./adk-code.rb
  adk-code --version
  ```
- [ ] Create tap repository (optional)
- [ ] Document Homebrew installation in README

### 3.3 Documentation
- [ ] Update README with installation instructions for all platforms
- [ ] Add platform-specific installation sections
- [ ] Document checksum verification
- [ ] Link to GitHub Releases
- [ ] Document system requirements (Go 1.24+, OS support)

## Phase 4: Ongoing Maintenance (Continuous)

### 4.1 Monitoring
- [ ] Set up GitHub Actions usage monitoring
- [ ] Set up workflow failure notifications
- [ ] Monitor build times and optimize if needed
- [ ] Check for deprecated Actions (update quarterly)

### 4.2 Dependency Management
- [ ] Enable Dependabot for GitHub Actions (`dependabot.yml`)
- [ ] Enable Dependabot for Go modules
- [ ] Review and merge dependency updates regularly
- [ ] Pin major versions of Actions for stability

### 4.3 Regular Testing
- [ ] Test CI locally before pushing changes
- [ ] Test release process quarterly
- [ ] Test on actual platforms (Windows, macOS, Linux)
- [ ] Collect user feedback on installation

### 4.4 Documentation Maintenance
- [ ] Update ADR 0003 if processes change
- [ ] Keep CI_CD_GUIDE.md current
- [ ] Update troubleshooting section with real issues
- [ ] Review and improve error messages in scripts

### 4.5 Performance Optimization
- [ ] Monitor CI run times
- [ ] Optimize build matrices if needed
- [ ] Cache analysis and improvements
- [ ] Parallel job optimization

## Validation Checklist

After each phase, verify:

### Phase 1 Complete
- [ ] `make ci-check` passes locally
- [ ] GitHub Actions CI passes on pull request
- [ ] All 6 binaries build successfully
- [ ] Binaries are executable and work
- [ ] Coverage report is generated and meets threshold

### Phase 2 Complete
- [ ] Creating a git tag triggers release workflow
- [ ] Release workflow completes successfully
- [ ] GitHub Release is created with binaries
- [ ] Checksums are correct
- [ ] Release notes are generated

### Phase 3 Complete
- [ ] README has installation instructions for all platforms
- [ ] Homebrew formula works (if implemented)
- [ ] Package manager distributions work (if implemented)
- [ ] Documentation is complete and accurate

### Phase 4 Ongoing
- [ ] No failing workflows in past 30 days
- [ ] Build times are < 5 minutes average
- [ ] Dependencies are up to date
- [ ] Security scans pass
- [ ] Coverage remains > 70%

## Success Metrics

Target these metrics for successful CI/CD:

| Metric | Target | Current |
|--------|--------|---------|
| CI Pass Rate | >95% | - |
| Build Time | <5 min | - |
| Test Coverage | >70% | - |
| Security Issues | 0 known | - |
| Release Time | <10 min | - |
| Platform Support | 6+ | - |

## Troubleshooting During Implementation

### GitHub Actions quota exceeded
- Check usage in Settings → Billing → Actions
- Optimize build matrix (reduce platforms temporarily)
- Consider paid plan if needed

### Build fails for specific platform
- Test locally with that GOOS/GOARCH
- Check for platform-specific code/dependencies
- Use build constraints if needed

### Release workflow doesn't trigger
- Verify tag format matches `v*` pattern
- Check branch is set to main/master
- Verify push permissions on tags

### Binaries don't work on target platform
- Download and test before releasing
- Check for glibc version compatibility (Linux)
- Verify architecture match (amd64 vs arm64)

## Resources

### Documentation
- [ADR 0003: CI/CD and Build Process](./adr/0003-cicd-and-build-process.md)
- [CI/CD Implementation Guide](./CI_CD_GUIDE.md)
- [GitHub Actions Docs](https://docs.github.com/en/actions)

### Tools
- [act - Local GitHub Actions testing](https://github.com/nektos/act)
- [GitHub CLI](https://cli.github.com/)
- [golangci-lint](https://golangci-lint.run/)

### References
- [Go Build Documentation](https://golang.org/doc/cmd#hdr-Build_packages_and_dependencies)
- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)

