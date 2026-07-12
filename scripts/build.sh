#!/bin/bash

# Exit on error
set -e

# Set static linking options
export CGO_ENABLED=0

# Build
echo "Building proxyforge..."
go build -ldflags="-w -s" -o bin/proxyforge cmd/main.go

echo "Build successful! Binary is located at bin/proxyforge"
