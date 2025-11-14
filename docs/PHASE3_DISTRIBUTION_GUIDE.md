# Phase 3: Distribution Channels Implementation Guide

## Overview

This document provides a detailed implementation guide for Phase 3 of the `adk-code` CI/CD and distribution strategy. Phase 3 focuses on expanding distribution beyond GitHub Releases to multiple package managers, reducing installation friction and improving user experience across different platforms.

**Branch**: `feature/phase3-distribution-channels`

**Related ADR**: [ADR 0004: Distribution Channels (Phase 3)](./adr/0004-distribution-channels-phase-3.md)

## Quick Summary

| Channel | OS | Priority | Status | Timeline |
|---------|----|----|--------|----------|
| **Homebrew** | macOS | Primary | Phase 3A | Weeks 5-6 |
| **APT** | Debian/Ubuntu | Secondary | Phase 3B | Weeks 7-8 |
| **YUM/DNF** | RHEL/CentOS | Tertiary | Phase 3C | Weeks 9-10 |
| **Scoop** | Windows | Optional | Phase 3D | Weeks 11-12 |
| **GitHub Releases** | All | Universal | Continuous | All phases |

## Phase 3A: Homebrew (Weeks 5-6)

### Objectives
- Create custom Homebrew tap repository
- Develop automated cask generation
- Publish first release via Homebrew
- Validate installation and upgrades

### Key Tasks

1. **Create `homebrew-adk-code` Repository**
   ```bash
   # Create repository structure
   mkdir homebrew-adk-code
   cd homebrew-adk-code
   git init
   
   # Create directory structure
   mkdir -p Casks Formula scripts
   ```

2. **Develop Cask Template** (`Casks/adk-code.rb`)
   ```ruby
   # See ADR 0004, Section 2.3 for complete template
   cask "adk-code" do
     version "{{ VERSION }}"
     sha256 "{{ ARM64_SHA }}"  # auto-generated
     
     url "https://github.com/.../adk-code-v#{version}-darwin-#{arch}"
     homepage "https://github.com/raphaelmansuy/adk-code"
     license "MIT"
     
     binary "adk-code"
     
     test do
       system "#{staged_path}/adk-code", "--version"
     end
   end
   ```

3. **Create Update Script** (`scripts/update-cask.sh`)
   - Accepts version as argument
   - Downloads binaries
   - Computes SHA256 checksums
   - Updates cask file template
   - Ready for commit/push

4. **Set Up CI/CD Integration** (`.github/workflows/distribute.yml`)
   - Trigger on release publication
   - Call `update-cask.sh` with version
   - Commit and push to `homebrew-adk-code`
   - Use GitHub Secrets for authentication

5. **Test Locally**
   ```bash
   # Simulate installation
   brew audit --cask adk-code
   brew install --build-from-source ./Casks/adk-code.rb
   
   # Verify functionality
   adk-code --version
   
   # Test upgrade
   brew upgrade adk-code
   ```

### Success Criteria
- [ ] `homebrew-adk-code` repository created and published
- [ ] Cask file passes `brew audit`
- [ ] Installation works on macOS (both arm64 and amd64)
- [ ] Upgrade from v1.0 to v1.1 works seamlessly
- [ ] CI/CD workflow automatically updates cask on release
- [ ] Installation instructions documented in README

### Deliverables
- `homebrew-adk-code` GitHub repository
- Automated cask update workflow
- Documentation for Homebrew installation
- Example: `brew tap raphaelmansuy/adk-code && brew install adk-code`

---

## Phase 3B: APT Repository (Weeks 7-8)

### Objectives
- Create APT repository infrastructure
- Set up Debian package generation
- Implement GPG signing
- Publish to GitHub Pages

### Key Tasks

1. **Create `adk-code-apt` Repository**
   ```bash
   mkdir adk-code-apt
   cd adk-code-apt
   git init
   
   # Create directory structure
   mkdir -p dists/stable/main/binary-{amd64,arm64,armhf}
   mkdir -p dists/stable/main/source
   mkdir -p pool/main/adk-code
   ```

2. **Enable GitHub Pages**
   - Settings → Pages → Source: main branch
   - Custom domain (optional): `apt.adk-code.dev`
   - Repository becomes: `https://<username>.github.io/adk-code-apt/`

3. **Configure nfpm** (in main adk-code repo)
   ```yaml
   # nfpm.yaml
   name: adk-code
   arch: amd64
   platform: linux
   version: 1.0.0
   license: MIT
   
   files:
     "adk-code-linux-amd64": "/usr/local/bin/adk-code"
   
   scripts:
     postinstall: scripts/postinstall.sh
     preremove: scripts/preremove.sh
   ```

4. **Generate Packages**
   ```bash
   # Install nfpm
   go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
   
   # Generate Debian packages
   for arch in amd64 arm64; do
     nfpm package -f nfpm.yaml -p deb \
       -a $arch -o adk-code_1.0.0_${arch}.deb
   done
   ```

5. **Set Up GPG Signing**
   ```bash
   # Generate GPG key (one-time, in CI/CD secrets)
   gpg --gen-key
   
   # Export public key
   gpg --export -a "ADK Code Maintainer" > apt/key.gpg
   
   # Store private key in GitHub Secrets
   gpg --export-secret-key --armor | base64 > secret.txt
   ```

6. **Create Update Script** (`scripts/update-apt-repo.sh`)
   - Copy .deb files to pool
   - Generate package indices with reprepro
   - Sign Release file with GPG
   - Push to GitHub Pages

7. **Add CI/CD Workflow**
   - Runs after Homebrew publication
   - Builds .deb packages with nfpm
   - Signs packages and metadata
   - Pushes to `adk-code-apt` repository

### Success Criteria
- [ ] `adk-code-apt` repository created with proper structure
- [ ] .deb packages generate successfully for all architectures
- [ ] GitHub Pages hosting works (packages accessible via HTTPS)
- [ ] GPG signatures verify correctly
- [ ] Installation works: `sudo apt install adk-code`
- [ ] Upgrades work seamlessly
- [ ] CI/CD automatically publishes to APT on release

### Deliverables
- `adk-code-apt` GitHub repository with GitHub Pages
- nfpm configuration for .deb generation
- APT repository update workflow
- Documentation for Debian/Ubuntu installation

---

## Phase 3C: YUM/DNF Repository (Weeks 9-10)

### Objectives
- Create YUM repository infrastructure
- Set up RPM package generation
- Implement GPG signing for RPMs
- Publish to GitHub Pages

### Key Tasks

1. **Create `adk-code-yum` Repository**
   ```bash
   mkdir adk-code-yum
   cd adk-code-yum
   git init
   
   # Create directory structure
   for arch in x86_64 aarch64; do
     mkdir -p stable/$arch/repodata
   done
   ```

2. **Enable GitHub Pages**
   - Settings → Pages → Source: main branch
   - Repository becomes: `https://<username>.github.io/adk-code-yum/`

3. **Generate RPM Packages**
   ```bash
   # Using same nfpm.yaml as APT
   for arch in x86_64 aarch64; do
     nfpm package -f nfpm.yaml -p rpm \
       -a $arch -o adk-code-1.0.0-1.${arch}.rpm
   done
   ```

4. **Set Up GPG Signing for RPMs**
   ```bash
   # Configure ~/.rpmmacros for signing
   echo '%_gpg_name ADK Code Maintainer' > ~/.rpmmacros
   
   # Sign individual packages
   rpm --addsign adk-code-1.0.0-1.x86_64.rpm
   ```

5. **Create Repository Metadata**
   ```bash
   # Install createrepo
   # (automatically included in RHEL/CentOS/Fedora)
   
   # Generate repository metadata
   createrepo --sign --gpg-key "ADK Code Maintainer" \
     stable/x86_64/
   ```

6. **Create Update Script** (`scripts/update-yum-repo.sh`)
   - Copy .rpm files to pool
   - Generate metadata with createrepo
   - Sign repository with GPG
   - Push to GitHub Pages

7. **Add CI/CD Workflow**
   - Builds .rpm packages with nfpm
   - Signs packages individually
   - Generates repository metadata
   - Pushes to `adk-code-yum` repository

### Success Criteria
- [ ] `adk-code-yum` repository created with proper structure
- [ ] .rpm packages generate successfully for all architectures
- [ ] Repository metadata generates correctly
- [ ] GPG signatures verify on RHEL/CentOS
- [ ] Installation works: `sudo yum install adk-code`
- [ ] Upgrades work: `sudo yum upgrade adk-code`
- [ ] CI/CD automatically publishes to YUM on release

### Deliverables
- `adk-code-yum` GitHub repository with GitHub Pages
- RPM repository update workflow
- Documentation for RHEL/CentOS/Fedora installation

---

## Phase 3D: Polish & Scoop (Weeks 11-12)

### Objectives
- Optional Scoop distribution for Windows
- Comprehensive integration testing
- Final documentation
- Launch communication

### Key Tasks

1. **Create `scoop-adk-code` Repository** (Optional)
   ```bash
   mkdir scoop-adk-code
   cd scoop-adk-code
   git init
   
   mkdir bucket scripts
   ```

2. **Develop Scoop Manifest** (`bucket/adk-code.json`)
   ```json
   {
     "version": "1.0.0",
     "description": "Multi-model AI coding assistant CLI",
     "homepage": "https://github.com/raphaelmansuy/adk-code",
     "license": "MIT",
     "architecture": {
       "64bit": {
         "url": "https://github.com/.../adk-code-v1.0.0-windows-amd64.exe",
         "hash": "sha256:..."
       }
     },
     "bin": "adk-code.exe",
     "checkver": "github",
     "autoupdate": { ... }
   }
   ```

3. **Integration Testing**
   - **macOS**: Test on arm64 and amd64 VMs
     ```bash
     brew tap raphaelmansuy/adk-code
     brew install adk-code
     adk-code --version
     brew upgrade adk-code
     ```
   
   - **Ubuntu**: Test on 20.04 and 22.04
     ```bash
     sudo apt update
     sudo apt install adk-code
     adk-code --version
     sudo apt upgrade adk-code
     ```
   
   - **CentOS**: Test on 7 and 8
     ```bash
     sudo yum install adk-code
     adk-code --version
     sudo yum upgrade adk-code
     ```

4. **Comprehensive Documentation**
   - Update main README.md with all installation methods
   - Create INSTALLATION.md with troubleshooting
   - Document each channel (Homebrew, APT, YUM, Scoop)
   - Add FAQ section
   - Document configuration file locations

5. **Launch Communication**
   - Blog post: "adk-code is now available via package managers"
   - Update GitHub README
   - Announce on social media
   - Notify existing users
   - Ask for feedback

6. **Set Up Monitoring**
   - Track download statistics from each channel
   - Monitor installation issues
   - Collect user feedback
   - Plan quarterly reviews

### Success Criteria
- [ ] All distribution channels tested thoroughly
- [ ] Comprehensive documentation complete
- [ ] User communication finished
- [ ] Monitoring systems in place
- [ ] Users can install via all channels
- [ ] Automatic updates work seamlessly

### Deliverables
- Complete documentation for all channels
- Integration test results
- Launch blog post
- Monitoring dashboard (optional)

---

## Testing Matrix

### Comprehensive Test Plan

| Channel | OS | Version | Arch | Test Cases |
|---------|----|----|------|-----------|
| Homebrew | macOS | 12+ | arm64 | Install, Verify, Upgrade |
| Homebrew | macOS | 12+ | amd64 | Install, Verify, Upgrade |
| APT | Ubuntu | 20.04 | amd64 | Install, Verify, Upgrade, Remove |
| APT | Ubuntu | 22.04 | amd64 | Install, Verify, Upgrade, Remove |
| APT | Ubuntu | 22.04 | arm64 | Install, Verify, Upgrade, Remove |
| YUM | CentOS | 7 | x86_64 | Install, Verify, Upgrade, Remove |
| YUM | CentOS | 8 | x86_64 | Install, Verify, Upgrade, Remove |
| YUM | Rocky | 8 | aarch64 | Install, Verify, Upgrade, Remove |
| Scoop | Windows | 10+ | amd64 | Install, Verify, Upgrade |

**Test Commands:**
```bash
# Homebrew
brew install adk-code
which adk-code
adk-code --version
brew upgrade adk-code
brew uninstall adk-code

# APT
sudo apt install adk-code
which adk-code
adk-code --version
sudo apt upgrade adk-code
sudo apt remove adk-code

# YUM
sudo yum install adk-code
which adk-code
adk-code --version
sudo yum upgrade adk-code
sudo yum remove adk-code

# Scoop
scoop install adk-code
adk-code.exe --version
scoop update adk-code
scoop uninstall adk-code
```

---

## Security Checklist

- [ ] GPG keys generated with strong entropy
- [ ] Private keys stored as GitHub Secrets (encrypted)
- [ ] All packages signed with GPG
- [ ] All repository metadata signed
- [ ] SHA256 checksums verified on installation
- [ ] HTTPS-only distribution URLs
- [ ] Vulnerability scanning integrated
- [ ] Key rotation policy documented
- [ ] Emergency key rotation procedure tested
- [ ] Security advisory process established

---

## Secrets Management

### Required GitHub Secrets

```
HOMEBREW_REPO_TOKEN          # PAT for homebrew-adk-code repo
APT_REPO_TOKEN               # PAT for adk-code-apt repo
RPM_REPO_TOKEN               # PAT for adk-code-yum repo
SCOOP_REPO_TOKEN             # PAT for scoop-adk-code repo
GPG_KEY_ID                   # Key ID for signing (short form)
GPG_PASSPHRASE               # GPG private key passphrase
GPG_PRIVATE_KEY              # Exported secret key (base64)
```

### Setup Instructions

1. **Create Personal Access Tokens**
   - Go to GitHub Settings → Developer settings → Personal access tokens
   - Create token with `repo` and `contents` permissions
   - Set expiration to 90 days
   - Add to repo secrets

2. **Generate GPG Key**
   ```bash
   gpg --gen-key
   # Fill in: Real name, Email, Comment (RPM Signing Key)
   
   # Get key ID
   gpg --list-keys
   # Output: pub   rsa4096/12345678 2025-11-14
   # Key ID is: 12345678
   
   # Export for secrets
   gpg --export-secret-key --armor 12345678 | base64
   ```

3. **Add to GitHub Secrets**
   - Repository → Settings → Secrets and variables → Actions
   - Add each secret with exact names above

---

## CI/CD Workflow Reference

### Distribution Workflow Execution

```
Release published (v1.0.0)
    ↓
.github/workflows/distribute.yml triggered
    ├── publish-homebrew
    │   ├── Update cask version
    │   ├── Compute SHA256
    │   └── Push to homebrew-adk-code
    │
    ├── publish-apt
    │   ├── Generate .deb packages
    │   ├── Generate package indices
    │   ├── Sign Release file
    │   └── Push to adk-code-apt
    │
    ├── publish-yum
    │   ├── Generate .rpm packages
    │   ├── Generate repository metadata
    │   ├── Sign with GPG
    │   └── Push to adk-code-yum
    │
    └── publish-scoop
        ├── Update manifest
        ├── Compute SHA256
        └── Push to scoop-adk-code
```

---

## Rollback Procedures

### If Homebrew Fails
```bash
# Users can install previous version
brew install adk-code@1.0.0

# Or reinstall from GitHub directly
curl -sL https://github.com/.../releases/.../adk-code | tar xz
sudo mv adk-code /usr/local/bin/
```

### If APT Fails
```bash
# Users can pin to previous version
sudo apt install adk-code=0.9.9-1

# Or install from GitHub directly
```

### If YUM Fails
```bash
# Users can pin to previous version
sudo yum install adk-code-0.9.9-1

# Or install from GitHub directly
```

---

## Monitoring & Health Checks

### Quarterly Health Check Checklist
- [ ] All repositories accessible via HTTPS
- [ ] Package metadata integrity verified
- [ ] GPG signatures still valid
- [ ] Download statistics reviewed
- [ ] User issues tracked and addressed
- [ ] Package manager policy changes reviewed
- [ ] Dependency versions updated
- [ ] Security audit completed

### Metrics to Track
- Downloads per channel (GitHub, Homebrew, APT, YUM, Scoop)
- Installation success rate
- User support tickets by channel
- Average time to upgrade
- Error rates from each package manager

---

## Next Steps

1. **Immediately**: Review ADR 0004 in detail
2. **Week 1-2**: Create Homebrew infrastructure and test
3. **Week 3-4**: Create APT repository and GPG setup
4. **Week 5-6**: Create YUM repository and test
5. **Week 7-8**: Integration testing and documentation
6. **Week 9-10**: Scoop (optional) and final polish
7. **Week 11-12**: Public launch and monitoring setup

---

## Questions & Discussions

See ADR 0004, Section 13 for open questions and discussion points:
- Differential updates?
- Static linking?
- Snap/Flatpak support?
- Container images (Docker)?
- Automatic security updates?

---

## Document History

| Date | Author | Status | Notes |
|------|--------|--------|-------|
| 2025-11-14 | @raphaelmansuy | Created | Initial Phase 3 implementation guide |

