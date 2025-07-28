#!/usr/bin/env bash
set -euo pipefail

# Start from repo root regardless of exec location
cd "$(git rev-parse --show-toplevel)"

mapfile -t go_files < <(find . -name '*.go' \
  -not -path './bazel-out/*' \
  -not -path './bazel-bin/*' \
  -not -path './bazel-testlogs/*' \
  -not -path './vendor/*')

unformatted=$(gofmt -l "${go_files[@]}")
if [[ -n "$unformatted" ]]; then
  echo "These files are not gofmt'd:"
  echo "$unformatted"
  exit 1
fi
