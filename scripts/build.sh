#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# --- Configuration ---
# Output directory for the binary
OUTPUT_DIR="./build"
# Name of the binary file
BINARY_NAME="you-go-server"
# Path to the main Go package
MAIN_PACKAGE="youGo/cmd/api"
# Target Operating System (see `go tool dist list`)
TARGET_OS="linux"
# Target Architecture (see `go tool dist list`)
TARGET_ARCH="amd64"
# (Optional) Version information to embed (can be set via CI/CD)
VERSION=${APP_VERSION:-"1.0.0"} # Default version if APP_VERSION env var is not set
COMMIT_HASH=$(git rev-parse --short HEAD || echo "dev")
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

# --- Build Flags ---
# -ldflags allow embedding variables into the binary
# Variable paths must match variables defined in your Go code (e.g., in main package)
# Example Go code:
#   var (
#       AppVersion string
#       BuildCommit string
#       BuildDate   string
#   )
#
LDFLAGS=(
  "-s -w" # Strip debugging information and symbol table (-w omits DWARF symbols)
  "-X '${MAIN_PACKAGE}.AppVersion=${VERSION}'"
  "-X '${MAIN_PACKAGE}.BuildCommit=${COMMIT_HASH}'"
  "-X '${MAIN_PACKAGE}.BuildDate=${BUILD_DATE}'"
)
# Join the array elements into a single string for the ldflags argument
LDFLAGS_STR="${LDFLAGS[*]}"


# --- Build Process ---
echo "--- Starting Production Build ---"
echo "Version: ${VERSION}"
echo "Commit: ${COMMIT_HASH}"
echo "Build Date: ${BUILD_DATE}"
echo "Target: ${TARGET_OS}/${TARGET_ARCH}"

# Clean previous build directory (optional)
# rm -rf "${OUTPUT_DIR}"

# Create output directory
mkdir -p "${OUTPUT_DIR}"

# Set target environment variables for cross-compilation
export GOOS="${TARGET_OS}"
export GOARCH="${TARGET_ARCH}"
# CGO_ENABLED=0 is often used for static binaries, especially with musl/alpine targets
# export CGO_ENABLED=0

echo "Building Go binary..."
# Build the application
go build \
  -ldflags="${LDFLAGS_STR}" \
  -o "${OUTPUT_DIR}/${BINARY_NAME}" \
  "${MAIN_PACKAGE}"

# Check if build was successful
if [ $? -ne 0 ]; then
  echo "ERROR: Go build failed!"
  exit 1
fi

echo "Build successful!"
echo "Binary created at: ${OUTPUT_DIR}/${BINARY_NAME}"

# (Optional) List details of the created binary
ls -lh "${OUTPUT_DIR}/${BINARY_NAME}"
# file "${OUTPUT_DIR}/${BINARY_NAME}" # Show file type information

echo "--- Build Complete ---"
