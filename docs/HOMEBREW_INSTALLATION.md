# Homebrew Installation Guide for adk-code

## Overview

`adk-code` can be installed on macOS using Homebrew, a popular package manager for macOS. This guide explains how to install and use adk-code via Homebrew.

## Prerequisites

### Install Homebrew (if not already installed)

If you don't have Homebrew installed, install it first:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

**Requirements:**

- macOS Sonoma (14) or higher (macOS 10.15+ may work but unsupported)
- Apple Silicon (M1/M2/M3) or Intel CPU
- Command Line Tools for Xcode
- Bash shell

### Verify Homebrew Installation

```bash
brew --version
# Example output: Homebrew 4.0.0
```

## Installation

### Step 1: Add the adk-code Tap

A "tap" is a custom Homebrew repository. Add the adk-code tap:

```bash
brew tap raphaelmansuy/adk-code
```

This command clones the `homebrew-adk-code` repository and makes its formulas/casks available.

**What happens:**

- Downloads the tap repository from [https://github.com/raphaelmansuy/homebrew-adk-code](https://github.com/raphaelmansuy/homebrew-adk-code)
- Stores it in `$(brew --repository)/Library/Taps/raphaelmansuy/homebrew-adk-code`
- adk-code becomes available for installation

### Step 2: Install adk-code

```bash
brew install adk-code
```

**What happens:**

- Downloads the latest precompiled binary for your architecture
- Verifies SHA256 checksum for security
- Installs to `/opt/homebrew/bin/adk-code` (Apple Silicon) or `/usr/local/bin/adk-code` (Intel)
- Makes `adk-code` command available in your PATH

### Step 3: Verify Installation

```bash
adk-code --version
# Example output: adk-code version 0.1.0
```

## Usage

Once installed, use adk-code normally:

```bash
# Start interactive REPL
adk-code

# Run a specific command
adk-code "write a hello world script in Python"

# Set model
adk-code /use gemini-2.5-flash
```

See the [Quick Reference](./QUICK_REFERENCE.md) for more commands.

## Updating adk-code

### Check for Updates

```bash
brew outdated
# Shows which packages have newer versions available
```

### Update adk-code

```bash
brew upgrade adk-code
```

Or update all Homebrew packages:

```bash
brew upgrade
```

### Automatic Updates

Homebrew does NOT automatically update packages. You must run `brew upgrade` manually or set up a cron job.

## Uninstalling

### Remove adk-code

```bash
brew uninstall adk-code
```

### Remove the Tap (optional)

If you no longer want the tap:

```bash
brew untap raphaelmansuy/adk-code
```

## Troubleshooting

### Command Not Found After Installation

**Problem:** `adk-code: command not found`

**Solution 1:** Ensure Homebrew's bin is in your PATH

```bash
# Check if Homebrew bin is in PATH
echo $PATH | grep -q "/opt/homebrew/bin" && echo "OK" || echo "NOT FOUND"

# For Apple Silicon, add to your shell config (~/.zshrc or ~/.bash_profile):
eval "$(/opt/homebrew/bin/brew shellenv)"

# For Intel Mac, add:
export PATH="/usr/local/bin:$PATH"

# Then reload your shell:
source ~/.zshrc  # or source ~/.bash_profile
```

**Solution 2:** Check installation location

```bash
which adk-code
# Should output: /opt/homebrew/bin/adk-code (Apple Silicon) or /usr/local/bin/adk-code (Intel)
```

### Checksum Verification Failed

**Problem:** `checksum does not match` error

**Solution:**

```bash
# Clear Homebrew cache
brew cleanup

# Try installing again
brew install adk-code
```

### Tap Not Found

**Problem:** `Error: No taps found for raphaelmansuy/adk-code`

**Solution:**

```bash
# Make sure you've tapped the repository
brew tap raphaelmansuy/adk-code

# Verify tap was added
brew tap
# Should show: raphaelmansuy/adk-code
```

### Slow Installation

**Problem:** Installation takes a long time

**Reasons:**

- First installation may be slow due to downloading the tap
- Network speed affects download time
- Homebrew is running dependency checks

**What to do:**

- This is normal. Be patient.
- Check your internet connection
- Run `brew update` to update formula caches

### Version Mismatch

**Problem:** Installed version doesn't match latest release

**Solution:**

```bash
# Update Homebrew metadata
brew update

# Check available versions
brew info adk-code

# Reinstall with specific version
brew reinstall adk-code
```

## Configuration

After installation, configure adk-code:

### Set API Key (Google Gemini)

```bash
export GOOGLE_API_KEY="your-api-key-here"
```

Add to your shell config (~/.zshrc or ~/.bash_profile) to persist:

```bash
echo 'export GOOGLE_API_KEY="your-api-key-here"' >> ~/.zshrc
source ~/.zshrc
```

### Configuration Files

adk-code stores config in:

- macOS: `~/.adk-code/config.yaml`

See [Configuration Guide](./CONFIGURATION.md) for details.

## Architecture Support

The Homebrew cask supports multiple architectures:

| Architecture | macOS Version | Status |
|---|---|---|
| Apple Silicon (M1/M2/M3) | 11+ | ✅ Supported |
| Intel (x86_64) | 10.15+ | ✅ Supported |
| PowerPC (PPC) | Any | ❌ Not supported |

Homebrew automatically detects your architecture and installs the correct binary.

## Keeping Up to Date

### Manual Updates

```bash
# Check for updates
brew outdated adk-code

# Update if available
brew upgrade adk-code
```

### Subscribe to Releases

To be notified of new releases, watch the GitHub repository:

- Visit [https://github.com/raphaelmansuy/adk-code](https://github.com/raphaelmansuy/adk-code)
- Click "Watch" → "Releases only"

## Advanced Usage

### Install Specific Version

```bash
# List available versions
brew info adk-code

# Install specific version (if available)
brew install adk-code@0.1.0
```

### Build from Source (Advanced)

If a precompiled binary isn't available for your system:

```bash
# Install from source formula (if available)
brew install adk-code --build-from-source
```

Note: This requires Go toolchain and build dependencies.

### Link/Unlink

If you have multiple versions:

```bash
# Link a specific version
brew link adk-code

# Unlink
brew unlink adk-code
```

## Support

For issues or questions:

1. Check this guide's troubleshooting section
2. Visit [https://github.com/raphaelmansuy/adk-code/issues](https://github.com/raphaelmansuy/adk-code/issues)
3. Review [FAQ](./FAQ.md)

## References

- [Homebrew Documentation](https://docs.brew.sh/)
- [How to Create a Homebrew Cask](https://docs.brew.sh/Cask-Cookbook)
- [adk-code Documentation](./README.md)
