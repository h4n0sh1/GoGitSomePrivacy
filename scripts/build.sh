#!/bin/bash

# Build script for GoGitSomePrivacy
# Supports multiple platforms and architectures

set -e

VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

BUILD_DIR="./build/dist"
BINARY_NAME="gogitsomeprivacy"

# Create build directory
mkdir -p "${BUILD_DIR}"

# Platforms to build for
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
)

echo "Building GoGitSomePrivacy ${VERSION}"
echo "Commit: ${COMMIT}"
echo "Date: ${DATE}"
echo ""

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r -a array <<< "$platform"
    GOOS="${array[0]}"
    GOARCH="${array[1]}"
    
    output_name="${BINARY_NAME}-${GOOS}-${GOARCH}"
    
    if [ "$GOOS" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo "Building for ${GOOS}/${GOARCH}..."
    
    GOOS=$GOOS GOARCH=$GOARCH go build \
        -trimpath \
        -ldflags="${LDFLAGS}" \
        -o "${BUILD_DIR}/${output_name}" \
        ./cmd/gogitsomeprivacy
    
    echo "✓ Built ${BUILD_DIR}/${output_name}"
done

echo ""
echo "Build complete! Binaries are in ${BUILD_DIR}/"
ls -lh "${BUILD_DIR}/"
