#!/usr/bin/env bash
set -euo pipefail

cd "${BUILD_WORKSPACE_DIRECTORY}"

# Run directly in your shell's current directory
golangci-lint run ./...
