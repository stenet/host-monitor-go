#!/bin/bash
# build.sh - Cross-platform build script

set -e

APP_NAME="host-monitor"
VERSION=${VERSION:-"1.0.0"}
BUILD_DIR="./build"

echo "Building ${APP_NAME} v${VERSION}..."

# Create build directory
mkdir -p ${BUILD_DIR}

# Build for Windows (amd64) - with service support (no CGO needed)
echo "Building for Windows (amd64)..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
    -ldflags "-X main.version=${VERSION} -w -s -H windowsgui" \
    -o ${BUILD_DIR}/${APP_NAME}-windows-amd64.exe .

# Build for current OS (macOS) - with CGO support
echo "Building for current OS..."
CGO_ENABLED=1 go build \
    -ldflags "-X main.version=${VERSION} -w -s" \
    -o ${BUILD_DIR}/${APP_NAME}-$(go env GOOS)-$(go env GOARCH) .

# Note: Cross-compilation with CGO for Linux would require additional setup
echo "Note: Linux builds with accurate CPU monitoring require native compilation on Linux hosts"

echo "Done! All binaries built successfully."