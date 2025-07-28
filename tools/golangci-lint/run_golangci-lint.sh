#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${BUILD_WORKSPACE_DIRECTORY:-}" ]]; then
    echo "Error: BUILD_WORKSPACE_DIRECTORY not set" >&2
    exit 1
fi

cd "${BUILD_WORKSPACE_DIRECTORY}"

# Run directly in your shell's current directory
golangci-lint run ./...
