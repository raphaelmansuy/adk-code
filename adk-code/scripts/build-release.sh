#!/bin/bash

# build-release.sh - Cross-platform release builder for adk-code
# Usage: ./scripts/build-release.sh [version]
# 
# Builds adk-code for all supported platforms and architectures.
# Output: ./dist/adk-code-v{version}-{os}-{arch}[.exe]
#
# Environment variables:
#   PLATFORMS  - Override default platform matrix (space-separated, format: "os:arch:arm")
#   DIST_DIR   - Override output directory (default: ./dist)
#   VERBOSE    - Set to 1 for detailed output

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="adk-code"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
DIST_DIR="${DIST_DIR:-${PROJECT_ROOT}/../dist}"
VERSION="${1:-}"
VERBOSE="${VERBOSE:-0}"

# Default platform matrix
# Format: os:arch:arm (arm is optional, only used for linux/arm)
DEFAULT_PLATFORMS=(
  "linux:amd64"
  "linux:arm64"
  "linux:arm:7"
  "darwin:amd64"
  "darwin:arm64"
  "windows:amd64"
)

PLATFORMS=("${PLATFORMS:-${DEFAULT_PLATFORMS[@]}}")

# Functions
print_header() {
  echo -e "\n${GREEN}╔═══════════════════════════════════════════════════╗${NC}"
  echo -e "${GREEN}║ adk-code Cross-Platform Release Builder           ║${NC}"
  echo -e "${GREEN}╚═══════════════════════════════════════════════════╝${NC}\n"
}

print_info() {
  echo -e "${YELLOW}ℹ${NC}  $1"
}

print_success() {
  echo -e "${GREEN}✓${NC}  $1"
}

print_error() {
  echo -e "${RED}✗${NC}  $1" >&2
}

print_step() {
  echo -e "\n${GREEN}→${NC}  $1"
}

validate_environment() {
  # Check Go is installed
  if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
  fi
  
  # Check we're in the right directory
  if [[ ! -f "${PROJECT_ROOT}/go.mod" ]]; then
    print_error "go.mod not found in ${PROJECT_ROOT}"
    print_error "Please run this script from the project root or adk-code directory"
    exit 1
  fi
  
  # Check version script exists
  if [[ ! -f "${SCRIPT_DIR}/version.sh" ]]; then
    print_error "version.sh not found in ${SCRIPT_DIR}"
    exit 1
  fi
  
  print_success "Environment validated"
}

get_version() {
  local version="${1:-}"
  
  if [[ -z "$version" ]]; then
    # Try to get from git tag
    if git rev-parse --git-dir > /dev/null 2>&1; then
      version=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
      if [[ -z "$version" ]]; then
        version=$(bash "${SCRIPT_DIR}/version.sh" get)
      fi
    else
      version=$(bash "${SCRIPT_DIR}/version.sh" get)
    fi
  fi
  
  # Ensure version starts with 'v'
  if [[ ! "$version" =~ ^v ]]; then
    version="v${version}"
  fi
  
  echo "$version"
}

build_binary() {
  local goos="$1"
  local goarch="$2"
  local goarm="${3:-0}"
  local version="$4"
  
  local binary_base="${BINARY_NAME}-${version}-${goos}-${goarch}"
  local output="${DIST_DIR}/${binary_base}"
  
  # Add .exe extension for Windows
  if [[ "$goos" == "windows" ]]; then
    output="${output}.exe"
  fi
  
  # Prepare environment
  export GOOS="$goos"
  export GOARCH="$goarch"
  [[ "$goarm" != "0" ]] && export GOARM="$goarm"
  
  print_step "Building ${goos}/${goarch}${goarm:+v${goarm}}..."
  
  # Run build with ldflags
  if [[ "$VERBOSE" == "1" ]]; then
    go build -v \
      -ldflags="-s -w -X adk-code/internal/app.AppVersion=${version}" \
      -o "$output" \
      "$PROJECT_ROOT"
  else
    go build \
      -ldflags="-s -w -X adk-code/internal/app.AppVersion=${version}" \
      -o "$output" \
      "$PROJECT_ROOT" 2>&1 | grep -E "(^go:|error:|warning:)" || true
  fi
  
  if [[ ! -f "$output" ]]; then
    print_error "Failed to build ${goos}/${goarch}"
    return 1
  fi
  
  # Get file size
  local size
  size=$(ls -lh "$output" | awk '{print $5}')
  print_success "Built ${binary_base} (${size})"
  
  # Generate checksum
  local checksum_file="${output}.sha256"
  sha256sum "$output" > "$checksum_file"
  
  return 0
}

build_all() {
  local version="$1"
  local total=${#PLATFORMS[@]}
  local current=0
  local failed=0
  
  print_step "Building ${total} platform(s)..."
  echo ""
  
  for platform in "${PLATFORMS[@]}"; do
    ((current++))
    
    # Parse platform string (os:arch:arm)
    IFS=':' read -r goos goarch goarm <<< "$platform"
    goarm="${goarm:-0}"
    
    echo "[${current}/${total}] Building for ${goos}/${goarch}${goarm:+v${goarm}}..."
    
    if ! build_binary "$goos" "$goarch" "$goarm" "$version"; then
      ((failed++))
    fi
  done
  
  echo ""
  
  if [[ $failed -eq 0 ]]; then
    print_success "All ${total} platform(s) built successfully!"
    return 0
  else
    print_error "${failed} out of ${total} build(s) failed!"
    return 1
  fi
}

print_summary() {
  local version="$1"
  
  echo ""
  echo -e "${GREEN}╔═══════════════════════════════════════════════════╗${NC}"
  echo -e "${GREEN}║ Build Summary                                     ║${NC}"
  echo -e "${GREEN}╚═══════════════════════════════════════════════════╝${NC}"
  echo ""
  
  echo "Version:    ${version}"
  echo "Output Dir: ${DIST_DIR}"
  echo "Binaries:   $(ls -1 "${DIST_DIR}/adk-code-${version}-"* 2>/dev/null | wc -l)"
  echo ""
  
  if [[ -d "$DIST_DIR" ]]; then
    echo "Artifacts:"
    ls -lh "${DIST_DIR}/adk-code-${version}-"* 2>/dev/null | awk '{printf "  %-40s %5s\n", $9, $5}'
    echo ""
  fi
  
  echo "Next steps:"
  echo "  1. Test the binaries on their target platforms"
  echo "  2. Create a Git tag: git tag ${version}"
  echo "  3. Push the tag: git push origin ${version}"
  echo "  4. GitHub Actions will create a release automatically"
  echo ""
}

main() {
  cd "$PROJECT_ROOT"
  
  print_header
  
  # Validate environment
  print_step "Validating environment"
  validate_environment
  
  # Get version
  VERSION=$(get_version "$VERSION")
  print_info "Version: ${VERSION}"
  
  # Create dist directory
  mkdir -p "$DIST_DIR"
  print_info "Output directory: ${DIST_DIR}"
  
  # Build all platforms
  if build_all "$VERSION"; then
    print_summary "$VERSION"
    exit 0
  else
    print_error "Build process completed with errors"
    exit 1
  fi
}

# Run main
main "$@"
