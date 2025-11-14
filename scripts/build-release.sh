#!/bin/bash
# Builds adk-code for all supported platforms
# Usage: ./scripts/build-release.sh [version]
# Output: ./dist/adk-code-v{version}-{os}-{arch}[.exe]

set -e

VERSION="${1:-$(./scripts/version.sh get)}"
DIST_DIR="../dist"
BINARY_NAME="adk-code"

# Ensure dist directory exists
mkdir -p "$DIST_DIR"

# Define build matrix: GOOS:GOARCH:GOARM
PLATFORMS=(
  "linux:amd64:0"
  "linux:arm64:0"
  "linux:arm:7"
  "darwin:amd64:0"
  "darwin:arm64:0"
  "windows:amd64:0"
)

echo "Building adk-code v${VERSION} for all platforms..."
echo "=================================================="

for platform in "${PLATFORMS[@]}"; do
  IFS=':' read -r GOOS GOARCH GOARM <<< "$platform"
  
  OUTPUT="${DIST_DIR}/${BINARY_NAME}-v${VERSION}-${GOOS}-${GOARCH}"
  [[ "$GOOS" == "windows" ]] && OUTPUT="${OUTPUT}.exe"
  
  # Skip GOARM export if 0 (not applicable for non-ARM platforms)
  if [[ "$GOARM" == "0" ]]; then
    echo "Building ${GOOS}/${GOARCH}..."
    GOOS="$GOOS" GOARCH="$GOARCH" go build \
      -ldflags="-s -w -X adk-code/internal/app.AppVersion=v${VERSION}" \
      -trimpath \
      -o "$OUTPUT" .
  else
    echo "Building ${GOOS}/${GOARCH} (GOARM=${GOARM})..."
    GOOS="$GOOS" GOARCH="$GOARCH" GOARM="$GOARM" go build \
      -ldflags="-s -w -X adk-code/internal/app.AppVersion=v${VERSION}" \
      -trimpath \
      -o "$OUTPUT" .
  fi
  
  # Generate SHA256 checksum
  shasum -a 256 "$OUTPUT" > "${OUTPUT}.sha256"
  
  # Print size
  ls -lh "$OUTPUT" | awk '{print "  Size: " $5}'
done

echo "=================================================="
echo "✓ Build complete! Binaries in $DIST_DIR"
ls -lh "$DIST_DIR"/adk-code-* | grep -v sha256 | awk '{print "  " $9 " (" $5 ")"}'
