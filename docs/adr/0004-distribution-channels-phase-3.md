# ADR 0004: Distribution Channels for adk-code (Phase 3)

## Status
Proposed

## Context

Phase 3 of the CI/CD implementation (as defined in ADR 0003) focuses on **distribution channels**—how users can easily discover, install, and update the `adk-code` binary. While Phase 1 (Foundation) and Phase 2 (Release Automation) establish the CI/CD pipeline and GitHub releases, Phase 3 must expand distribution to multiple package managers to reduce friction for different user groups:

- **macOS users**: Via Homebrew (primary distribution mechanism)
- **Linux users (Debian/Ubuntu)**: Via APT repositories
- **Linux users (RHEL/CentOS/Fedora)**: Via YUM/DNF repositories
- **Windows users**: Via GitHub Releases and direct downloads (no package manager priority)

Currently, users must manually download binaries from GitHub releases. This approach lacks:
- Version management and updates via package managers
- Discoverability (not listed in package manager search)
- Automatic dependency resolution
- Standardized installation paths and configuration
- Update notifications and automatic upgrading

This ADR establishes the architecture, tools, and processes for publishing `adk-code` to multiple package managers.

## Decision

### 1. Distribution Strategy Overview

**Primary Goals:**
1. Reduce installation friction for end users
2. Support automatic updates and version management
3. Integrate with standard package manager ecosystems
4. Maintain binary consistency across all channels
5. Implement GPG/cryptographic signing for security

**Distribution Channels (Priority Order):**
1. **Homebrew (Primary for macOS)** - ~40% of macOS developer market
2. **APT Repository (Primary for Debian/Ubuntu)** - Most popular Linux distribution
3. **YUM/DNF Repository (Secondary for RHEL/CentOS)** - Enterprise Linux
4. **GitHub Releases (Universal fallback)** - All platforms, no dependencies
5. **Scoop (Optional for Windows)** - Windows package manager

### 2. Homebrew Distribution (macOS Primary)

#### 2.1 Architecture

**Homebrew Concepts:**
- **Formula**: Ruby package definition for building from source
- **Cask**: Pre-built binary package (preferred for binaries)
- **Tap**: Custom formula repository (alternative to core)

**Decision**: Use **Homebrew Cask** (not formula) for `adk-code`:
- Cask is designed for pre-compiled binaries
- Simpler distribution (no compilation on user's machine)
- Faster installation (~5-10 seconds vs ~2 minutes for formula)
- Supports SHA256 verification automatically

#### 2.2 Cask File Structure

Create `homebrew-adk-code` repository as custom Homebrew tap:

```
homebrew-adk-code/
├── Casks/
│   └── adk-code.rb          # Main cask definition
├── Formula/
│   └── adk-code.rb          # Optional formula for building from source
├── README.md
├── LICENSE
└── .github/
    └── workflows/
        └── release.yml      # Auto-publish on new releases
```

#### 2.3 Cask File Template

```ruby
# homebrew-adk-code/Casks/adk-code.rb

cask "adk-code" do
  version "1.0.0"
  sha256 "abc123..."  # Auto-generated during release
  
  url "https://github.com/raphaelmansuy/adk-code/releases/download/v#{version}/adk-code-v#{version}-darwin-arm64.zip"
  homepage "https://github.com/raphaelmansuy/adk-code"
  license "MIT"
  
  # Supported architectures
  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/raphaelmansuy/adk-code/releases/download/v#{version}/adk-code-v#{version}-darwin-arm64"
      sha256 "sha256_arm64_value"
    elsif Hardware::CPU.intel?
      url "https://github.com/raphaelmansuy/adk-code/releases/download/v#{version}/adk-code-v#{version}-darwin-amd64"
      sha256 "sha256_amd64_value"
    end
  end
  
  binary "adk-code"
  
  # Post-install script (optional)
  postinstall do
    chmod 0755, staged_path/"adk-code"
  end
  
  # Uninstall if needed
  uninstall_postflight do
    # Clean up any caches
  end
  
  zap trash: [
    "~/.adk-code",
  ]
  
  # Testing
  test do
    system "#{staged_path}/adk-code", "--version"
  end
end
```

#### 2.4 Installation Process

**For Users:**
```bash
# Add custom tap
brew tap raphaelmansuy/adk-code

# Install
brew install adk-code

# Update
brew upgrade adk-code

# Uninstall
brew uninstall adk-code
```

**For Maintainers:**
```bash
# Create new cask
homebrew-adk-code/Casks/adk-code.rb

# Audit (before publishing)
brew audit --cask adk-code

# Publish (push to GitHub)
git push origin main
```

### 3. APT Repository Distribution (Debian/Ubuntu)

#### 3.1 Architecture

**Debian Repository Structure:**
```
deb.example.com/debian/
├── dists/
│   ├── stable/
│   │   ├── main/
│   │   │   ├── binary-amd64/
│   │   │   │   └── Packages.gz
│   │   │   └── binary-arm64/
│   │   │       └── Packages.gz
│   │   └── Release       # signed metadata
│   └── testing/
│       └── main/
│           └── ...
└── pool/
    └── main/
        └── a/
            └── adk-code/
                ├── adk-code_1.0.0_amd64.deb
                ├── adk-code_1.0.0_arm64.deb
                └── adk-code_1.0.0_armhf.deb
```

**Decision**: Use **GitHub Pages** + **reprepro** for hosting:
- GitHub Pages provides free hosting
- reprepro handles repository index generation
- Simpler than dedicated package server
- Automatic HTTPS
- Integrated with repository workflows

#### 3.2 Debian Package Creation

**Tool**: Use **nfpm** (multiplatform package builder) to create .deb files:

```yaml
# nfpm.yaml (in adk-code root)

name: adk-code
arch: amd64
platform: linux
version: 1.0.0
release: 1
license: MIT
homepage: https://github.com/raphaelmansuy/adk-code
description: Multi-model AI coding assistant CLI
maintainer: Raphael Mansuy <raphael@example.com>

files:
  "adk-code-linux-amd64": "/usr/local/bin/adk-code"

dirs:
  "/etc/adk-code": "0755"

scripts:
  postinstall: scripts/postinstall.sh
  preremove: scripts/preremove.sh

contents:
  - src: LICENSE
    dst: /usr/share/doc/adk-code/LICENSE
  - src: README.md
    dst: /usr/share/doc/adk-code/README.md

version_schema: semver
```

#### 3.3 APT Repository Setup (GitHub Pages)

**Structure:**
```
raphaelmansuy/adk-code-apt (separate GitHub repo)
├── dists/
│   ├── stable/
│   │   ├── main/
│   │   │   ├── binary-amd64/
│   │   │   │   └── Packages.gz
│   │   │   ├── binary-arm64/
│   │   │   │   └── Packages.gz
│   │   │   └── binary-armv7/
│   │   │       └── Packages.gz
│   │   ├── InRelease     # signed Release file
│   │   └── Release.gpg   # detached signature
│   └── testing/
│       └── ...
└── pool/
    └── main/
        └── adk-code/
            ├── adk-code_1.0.0_amd64.deb
            ├── adk-code_1.0.0_arm64.deb
            └── adk-code_1.0.0_armv7.deb
```

**Installation for Users:**
```bash
# Add repository key
curl -fsSL https://adk-code.example.com/apt/key.gpg | sudo apt-key add -

# Add repository
echo "deb [signed-by=/usr/share/keyrings/adk-code-archive-keyring.gpg] https://adk-code.example.com/apt stable main" | sudo tee /etc/apt/sources.list.d/adk-code.list

# Update and install
sudo apt update
sudo apt install adk-code

# Auto-update via system
# (apt will automatically check for updates)
```

#### 3.4 Repository Signing with GPG

**Process:**
1. Generate GPG key for repository (CI/CD secret)
2. Export public key to repository
3. Sign Release file with GPG
4. Generate InRelease file (clearsigned Release)

```bash
# Generate key (CI/CD only, one-time)
gpg --gen-key  # Non-interactive: gpg --batch --generate-key

# Export public key
gpg --export -a "Package Manager" > apt/key.gpg

# Create armored key for repository
gpg --export-secret-key > ~/.gnupg/secret-key.asc

# Sign Release file
gpg --clearsign -u "Package Manager" Release > InRelease
gpg -abs -u "Package Manager" Release > Release.gpg
```

### 4. YUM/DNF Repository Distribution (RHEL/CentOS)

#### 4.1 Architecture

**RPM Repository Structure:**
```
rpm.example.com/yum/
├── stable/
│   ├── x86_64/
│   │   ├── adk-code-1.0.0-1.x86_64.rpm
│   │   └── repodata/
│   │       ├── repomd.xml
│   │       ├── repomd.xml.asc
│   │       └── primary.xml.gz
│   └── aarch64/
│       └── ...
└── testing/
    └── ...
```

**Decision**: Use **createrepo** to generate repository metadata:
- Standard tool for RPM repositories
- Creates repodata index files
- Supports GPG signing
- Compatible with yum/dnf

#### 4.2 RPM Package Creation

**Tool**: Use **nfpm** to create .rpm files:

```yaml
# nfpm.yaml (same config, outputs multiple formats)

name: adk-code
version: 1.0.0
release: 1
arch: x86_64
platform: linux
license: MIT
homepage: https://github.com/raphaelmansuy/adk-code
description: Multi-model AI coding assistant CLI
maintainer: Raphael Mansuy <raphael@example.com>

files:
  "adk-code-linux-x86_64": "/usr/local/bin/adk-code"

scripts:
  postinstall: scripts/postinstall.sh
  preremove: scripts/preremove.sh

# Generate .rpm, .deb, .apk, .tar.gz from same config
```

#### 4.3 YUM Repository Setup (GitHub Pages)

**Installation for Users:**
```bash
# Create repo file
sudo tee /etc/yum.repos.d/adk-code.repo <<EOF
[adk-code]
name=ADK Code Repository
baseurl=https://rpm.example.com/yum/stable/\$basearch/
enabled=1
gpgcheck=1
gpgkey=https://rpm.example.com/yum/RPM-GPG-KEY-adk-code
EOF

# Install
sudo yum install adk-code

# Update
sudo yum update adk-code
```

#### 4.4 Repository Signing

Similar to APT, sign RPM packages with GPG:

```bash
# Create .rpmmacros for signing
echo '%_gpg_name ADK Code Maintainer' > ~/.rpmmacros

# Sign packages
rpm --addsign adk-code-1.0.0-1.x86_64.rpm

# Create repository metadata
createrepo --sign --gpg-key "ADK Code Maintainer" /path/to/repo
```

### 5. Scoop Distribution (Windows - Optional)

#### 5.1 Architecture

**Scoop Manifest Structure:**
```json
{
  "version": "1.0.0",
  "description": "Multi-model AI coding assistant CLI",
  "homepage": "https://github.com/raphaelmansuy/adk-code",
  "license": "MIT",
  "architecture": {
    "64bit": {
      "url": "https://github.com/raphaelmansuy/adk-code/releases/download/v1.0.0/adk-code-v1.0.0-windows-amd64.exe",
      "hash": "sha256:abc123..."
    }
  },
  "bin": "adk-code.exe",
  "checkver": "github",
  "autoupdate": {
    "architecture": {
      "64bit": {
        "url": "https://github.com/raphaelmansuy/adk-code/releases/download/v$version/adk-code-v$version-windows-amd64.exe",
        "hash": {
          "url": "https://github.com/raphaelmansuy/adk-code/releases/download/v$version/adk-code-v$version-windows-amd64.exe.sha256"
        }
      }
    }
  }
}
```

**Installation for Users:**
```bash
# Add bucket
scoop bucket add adk-code https://github.com/raphaelmansuy/scoop-adk-code

# Install
scoop install adk-code

# Update
scoop update adk-code
```

### 6. Release Automation Workflow

#### 6.1 CI/CD Pipeline Integration

**New Workflow: `.github/workflows/distribute.yml`**

```yaml
name: Distribute to Package Managers

on:
  release:
    types: [published]

jobs:
  publish-homebrew:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Update Homebrew Cask
        env:
          HOMEBREW_REPO_TOKEN: ${{ secrets.HOMEBREW_REPO_TOKEN }}
        run: |
          git clone https://raphaelmansuy:$HOMEBREW_REPO_TOKEN@github.com/raphaelmansuy/homebrew-adk-code.git
          cd homebrew-adk-code
          
          # Update cask with new version and checksums
          ./scripts/update-cask.sh "${{ github.ref_name }}"
          
          git config user.name "ADK Release Bot"
          git config user.email "release@adk-code.dev"
          git add Casks/adk-code.rb
          git commit -m "chore: update adk-code to ${{ github.ref_name }}"
          git push

  publish-apt:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Install nfpm
        run: go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
      
      - name: Create Debian packages
        run: |
          nfpm package -f nfpm.yaml -p deb \
            --config-from-env
        env:
          NFPM_AMDCK_PACKAGE_ARCH: amd64
      
      - name: Publish to APT repository
        env:
          APT_REPO_TOKEN: ${{ secrets.APT_REPO_TOKEN }}
        run: |
          git clone https://raphaelmansuy:$APT_REPO_TOKEN@github.com/raphaelmansuy/adk-code-apt.git
          cd adk-code-apt
          
          # Add new .deb files to pool
          cp ../adk-code_*.deb pool/main/adk-code/
          
          # Regenerate repository metadata
          reprepro includedeb stable pool/main/adk-code/*.deb
          
          # Sign Release file
          gpg --clearsign -u $GPG_KEY_ID -o dists/stable/InRelease dists/stable/Release
          
          git add .
          git commit -m "chore: add adk-code ${{ github.ref_name }}"
          git push
        env:
          GPG_KEY_ID: ${{ secrets.GPG_KEY_ID }}
          GPG_PASSPHRASE: ${{ secrets.GPG_PASSPHRASE }}

  publish-yum:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Install nfpm
        run: go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
      
      - name: Create RPM packages
        run: nfpm package -f nfpm.yaml -p rpm
      
      - name: Publish to YUM repository
        env:
          RPM_REPO_TOKEN: ${{ secrets.RPM_REPO_TOKEN }}
        run: |
          git clone https://raphaelmansuy:$RPM_REPO_TOKEN@github.com/raphaelmansuy/adk-code-yum.git
          cd adk-code-yum
          
          # Add new .rpm files
          cp ../adk-code-*.rpm stable/x86_64/
          
          # Regenerate repository metadata
          createrepo --update --sign --gpg-key "$GPG_KEY_ID" stable/x86_64/
          
          git add .
          git commit -m "chore: add adk-code ${{ github.ref_name }}"
          git push
        env:
          GPG_KEY_ID: ${{ secrets.GPG_KEY_ID }}
          GPG_PASSPHRASE: ${{ secrets.GPG_PASSPHRASE }}

  publish-scoop:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Update Scoop Manifest
        env:
          SCOOP_REPO_TOKEN: ${{ secrets.SCOOP_REPO_TOKEN }}
        run: |
          git clone https://raphaelmansuy:$SCOOP_REPO_TOKEN@github.com/raphaelmansuy/scoop-adk-code.git
          cd scoop-adk-code
          
          ./scripts/update-manifest.sh "${{ github.ref_name }}"
          
          git config user.name "ADK Release Bot"
          git config user.email "release@adk-code.dev"
          git add bucket/adk-code.json
          git commit -m "chore: update adk-code to ${{ github.ref_name }}"
          git push
```

### 7. Repository Infrastructure

#### 7.1 Separate Repository Strategy

Create three separate GitHub repositories for package managers:

**1. `homebrew-adk-code`**
- Contains Homebrew casks and formulas
- Tap repository (users: `brew tap raphaelmansuy/adk-code`)
- Light automation, mostly generated files

**2. `adk-code-apt`**
- Contains APT repository structure
- GitHub Pages enabled for serving .deb packages
- Automated Debian package hosting

**3. `adk-code-yum`**
- Contains YUM repository structure
- GitHub Pages enabled for serving .rpm packages
- Automated RPM package hosting

**Benefits:**
- Separation of concerns
- Independent update cycles
- Simpler CI/CD workflows (no conflicts)
- Easier to maintain metadata separately

#### 7.2 Secrets Management

**GitHub Secrets Required:**
```
HOMEBREW_REPO_TOKEN        # Personal access token for homebrew-adk-code
APT_REPO_TOKEN             # Personal access token for adk-code-apt
RPM_REPO_TOKEN             # Personal access token for adk-code-yum
SCOOP_REPO_TOKEN           # Personal access token for scoop-adk-code
GPG_KEY_ID                 # GPG key ID for signing
GPG_PASSPHRASE             # GPG private key passphrase
```

**Setup:**
1. Generate GPG key: `gpg --gen-key`
2. Export secret key: `gpg --export-secret-key --armor`
3. Add to GitHub secrets with base64 encoding
4. Create repo-specific tokens with limited permissions

### 8. Documentation & User Communication

#### 8.1 Installation Documentation

**Update README.md with:**
```markdown
## Installation

### macOS
```bash
brew tap raphaelmansuy/adk-code
brew install adk-code
```

### Linux (Debian/Ubuntu)
```bash
curl -fsSL https://adk-code.example.com/apt/key.gpg | sudo apt-key add -
echo "deb https://adk-code.example.com/apt stable main" | sudo tee /etc/apt/sources.list.d/adk-code.list
sudo apt update
sudo apt install adk-code
```

### Linux (RHEL/CentOS/Fedora)
```bash
sudo yum-config-manager --add-repo https://rpm.example.com/yum/stable
sudo yum install adk-code
```

### Windows (Scoop)
```bash
scoop bucket add adk-code https://github.com/raphaelmansuy/scoop-adk-code
scoop install adk-code
```

### Direct Download
Visit [Releases](https://github.com/raphaelmansuy/adk-code/releases)
```
```

#### 8.2 Configuration File Locations

**Standard Locations (per OS):**
- **macOS**: `~/.config/adk-code/` or `~/Library/Application Support/adk-code/`
- **Linux**: `~/.config/adk-code/` (XDG standard)
- **Windows**: `%APPDATA%\adk-code\` or `C:\Users\<user>\AppData\Local\adk-code\`

### 9. Quality Assurance

#### 9.1 Testing Distribution

**For Each Channel:**
1. **Functional Testing**: Install, run `--version`, verify binary works
2. **Upgrade Testing**: Install v1, upgrade to v2, verify seamless transition
3. **Dependency Testing**: Verify no missing runtime dependencies
4. **Integration Testing**: Test in clean VMs for each OS/arch

**Test Matrix:**
| Channel | OS | Arch | Test |
|---------|----|----|------|
| Homebrew | macOS | arm64 | brew install + upgrade |
| Homebrew | macOS | amd64 | brew install + upgrade |
| APT | Ubuntu 20.04 | amd64 | apt install + upgrade |
| APT | Ubuntu 22.04 | arm64 | apt install + upgrade |
| YUM | CentOS 7 | x86_64 | yum install + upgrade |
| YUM | CentOS 8 | aarch64 | yum install + upgrade |
| Scoop | Windows | amd64 | scoop install + upgrade |

#### 9.2 Pre-Release Validation

Before publishing to package managers:
1. Test each distribution channel locally
2. Verify cryptographic signatures
3. Check repository metadata integrity
4. Validate checksums match binaries
5. Confirm version numbers are correct

### 10. Monitoring & Maintenance

#### 10.1 Repository Health Checks

**Automated Checks:**
- Weekly repository metadata validation
- Monthly checksum verification
- Quarterly test installations from each channel
- Monitor download statistics

**Manual Reviews:**
- Quarterly repository cleanup (remove old versions)
- Annual security audit of signing keys
- Review of package manager policy changes
- User feedback collection

#### 10.2 Dependency Management

**Keep Updated:**
- Monitor Homebrew formula deprecations
- Track Debian/Ubuntu LTS releases
- Monitor RHEL/CentOS EOL dates
- Review YUM/DNF best practices

### 11. Security Considerations

#### 11.1 Cryptographic Signing

**Requirements:**
1. All packages must be cryptographically signed
2. All repository metadata must be signed with GPG
3. Public keys must be published and verified
4. Private keys stored as GitHub Secrets (encrypted)
5. Regular key rotation policy

**Implementation:**
- GPG signing in CI/CD workflows
- Automated signature verification during installation
- Key pinning documentation for users
- Emergency key rotation procedures

#### 11.2 Package Integrity

**Protections:**
1. SHA256 checksums for all binaries
2. GPG signatures for releases
3. Repository metadata signed with GPG
4. HTTPS-only distribution URLs
5. Automated vulnerability scanning

### 12. Implementation Timeline

#### Phase 3A: Weeks 5-6 (Homebrew Primary)
- [ ] Create `homebrew-adk-code` tap repository
- [ ] Develop cask template and update scripts
- [ ] Create Homebrew CI/CD workflow
- [ ] Test locally with `brew audit`
- [ ] Publish first release to Homebrew
- [ ] Document installation instructions

#### Phase 3B: Weeks 7-8 (APT Secondary)
- [ ] Create `adk-code-apt` repository with GitHub Pages
- [ ] Configure nfpm for .deb generation
- [ ] Set up repository signing with GPG
- [ ] Create APT CI/CD workflow
- [ ] Test with `apt install` in Ubuntu VM
- [ ] Document APT installation instructions

#### Phase 3C: Weeks 9-10 (YUM Tertiary)
- [ ] Create `adk-code-yum` repository with GitHub Pages
- [ ] Configure nfpm for .rpm generation
- [ ] Set up repository signing with GPG
- [ ] Create YUM CI/CD workflow
- [ ] Test with `yum install` in CentOS VM
- [ ] Document YUM installation instructions

#### Phase 3D: Weeks 11-12 (Polish & Scoop Optional)
- [ ] Create Scoop bucket (optional)
- [ ] Comprehensive integration testing
- [ ] Documentation finalization
- [ ] User communication and blog post
- [ ] Monitoring and health checks setup

### 13. Rollback & Contingency

**If Distribution Fails:**
1. GitHub Releases remain primary fallback
2. Previous version remains available in all channels
3. Users can pin to specific versions
4. Rollback procedure documented

**Version Pinning Examples:**
```bash
# Homebrew
brew install adk-code@1.0.0

# APT
apt install adk-code=1.0.0-1

# YUM
yum install adk-code-1.0.0-1

# Scoop
scoop install adk-code@1.0.0
```

## Consequences

### Positive

1. **User Experience**: Installation via familiar package managers (one command)
2. **Automatic Updates**: Users can enable automatic updates
3. **Discoverability**: Found via `brew search`, `apt search`, `yum search`
4. **Standardized Paths**: Binaries installed to standard locations (`/usr/local/bin`)
5. **Reduced Support Burden**: Users benefit from package manager support
6. **Professional Distribution**: Signals maturity and project quality
7. **Enterprise Adoption**: Easier integration with IT deployment systems
8. **Security**: Cryptographic signing provides integrity verification

### Negative/Challenges

1. **Maintenance Complexity**: Multiple package formats and repositories to maintain
2. **Repository Hosting**: Requires static hosting (GitHub Pages or similar)
3. **GPG Key Management**: Secret key storage and rotation procedures
4. **Testing Overhead**: Must test on multiple OS/distro combinations
5. **Policy Compliance**: Must comply with each package manager's guidelines
6. **Deprecation Risk**: Package managers may change requirements
7. **Rollback Complexity**: Removing old versions from multiple repositories

### Mitigation Strategies

- **Automation**: All distribution via CI/CD (minimal manual steps)
- **Templating**: Use scripts to generate casks/manifests from templates
- **Testing**: Comprehensive test matrix prevents regression
- **Documentation**: Clear runbooks for troubleshooting and rollback
- **Monitoring**: Automated health checks alert to issues early
- **Gradual Rollout**: Publish to Homebrew first, then expand
- **Community**: Engage package manager communities for guidance

## Related Documents

- `ADR 0003: CI/CD and Build Process` - Phase 1 and 2 implementation
- `ARCHITECTURE.md` - System design and component interactions
- `TOOL_DEVELOPMENT.md` - Tool development patterns

## References

### Homebrew
- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Homebrew Cask Cookbook](https://docs.brew.sh/Cask-Cookbook)
- [Homebrew Tap Documentation](https://docs.brew.sh/Taps)

### Debian/APT
- [Debian Repository Format](https://wiki.debian.org/DebianRepository/Format)
- [How to Package for Debian](https://wiki.debian.org/HowToPackageForDebian)
- [Debian Policy Manual](https://www.debian.org/doc/debian-policy/)

### RPM/YUM
- [How to Sign RPMs with GPG](https://access.redhat.com/articles/3359321)
- [Creating RPM Spec Files](https://fedoraproject.org/wiki/PackageMaintainers/CreatingPackageSpec)
- [RPM Repository Setup](https://access.redhat.com/articles/3359321)

### Tools
- [nfpm: Multiplatform Package Builder](https://nfpm.goreleaser.com/)
- [reprepro: Debian Repository Management](https://mirrorer.alioth.debian.org/)
- [createrepo: RPM Repository Creation](https://linux.die.net/man/8/createrepo)
- [Scoop: Windows Package Manager](https://scoop.sh/)

## Open Questions

1. Should we implement differential updates (delta packages)?
2. Should we provide statically-linked binaries (reduce dependencies)?
3. Should we publish to Snapcraft (Ubuntu snap store)?
4. Should we implement automatic security updates?
5. Should we establish a security advisory process?
6. Should we implement download analytics tracking?
7. Should we provide container images (Docker)?

## Decision Record

- **Date**: November 14, 2025
- **Proposer**: @raphaelmansuy
- **Status**: Proposed for implementation
- **Next Steps**: 
  1. Create `homebrew-adk-code` repository
  2. Develop Homebrew cask and update scripts
  3. Implement distribution CI/CD workflow
  4. Plan Phase 3A launch and user communication

---

## Appendix: Script Templates

### A. Homebrew Cask Update Script

```bash
#!/bin/bash
# scripts/update-cask.sh - Update Homebrew cask with new version

set -euo pipefail

VERSION="${1:?Version required (e.g., v1.0.0)}"
VERSION_NUM="${VERSION#v}"  # Remove 'v' prefix

# Download binaries and compute checksums
ARM64_URL="https://github.com/raphaelmansuy/adk-code/releases/download/${VERSION}/adk-code-${VERSION}-darwin-arm64"
AMD64_URL="https://github.com/raphaelmansuy/adk-code/releases/download/${VERSION}/adk-code-${VERSION}-darwin-amd64"

ARM64_SHA=$(curl -sL "$ARM64_URL" | shasum -a 256 | awk '{print $1}')
AMD64_SHA=$(curl -sL "$AMD64_URL" | shasum -a 256 | awk '{print $1}')

# Update cask file
cat > Casks/adk-code.rb <<EOF
cask "adk-code" do
  version "${VERSION_NUM}"
  
  on_macos do
    if Hardware::CPU.arm?
      url "${ARM64_URL}"
      sha256 "${ARM64_SHA}"
    elsif Hardware::CPU.intel?
      url "${AMD64_URL}"
      sha256 "${AMD64_SHA}"
    end
  end
  
  homepage "https://github.com/raphaelmansuy/adk-code"
  license "MIT"
  
  binary "adk-code"
  
  test do
    system "\#{staged_path}/adk-code", "--version"
  end
end
EOF

echo "Updated Casks/adk-code.rb to version ${VERSION_NUM}"
```

### B. Debian Repository Setup Script

```bash
#!/bin/bash
# scripts/setup-apt-repo.sh - Initialize APT repository structure

set -euo pipefail

REPO_DIR="${1:-.}"
DIST="${2:-stable}"

mkdir -p "$REPO_DIR/dists/$DIST/main/binary-"{amd64,arm64,armhf}
mkdir -p "$REPO_DIR/dists/$DIST/main/source"
mkdir -p "$REPO_DIR/pool/main/adk-code"

# Create initial Release file
cat > "$REPO_DIR/dists/$DIST/Release" <<EOF
Origin: ADK Code
Label: ADK Code Repository
Suite: $DIST
Codename: $DIST
Version: 1.0
Date: $(date -R)
Architectures: amd64 arm64 armhf
Components: main
Description: Multi-model AI coding assistant CLI
EOF

echo "APT repository structure created in $REPO_DIR"
```

### C. RPM Repository Setup Script

```bash
#!/bin/bash
# scripts/setup-yum-repo.sh - Initialize YUM repository structure

set -euo pipefail

REPO_DIR="${1:-.}"

for arch in x86_64 aarch64; do
  mkdir -p "$REPO_DIR/stable/$arch/repodata"
done

# Initialize empty repositories
for arch in x86_64 aarch64; do
  createrepo "$REPO_DIR/stable/$arch/"
done

echo "YUM repository structure created in $REPO_DIR"
```

