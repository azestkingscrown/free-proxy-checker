#!/bin/bash

# Exit on error
set -e

# Change to project root directory
cd "$(dirname "$0")/.."

# Set static linking options
export CGO_ENABLED=0

# Build
echo "Building proxyforge..."
go build -ldflags="-w -s" -o bin/proxyforge ./cmd/

echo "Build successful! Binary is located at bin/proxyforge"
