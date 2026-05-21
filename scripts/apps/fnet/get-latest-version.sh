#!/bin/bash
set -euo pipefail

VERSION="${1:-${VERSION:-}}"

# FNet is a Go app, version is just a timestamp-based tag
# For CI builds, we use the commit short hash
if [ -z "${VERSION}" ]; then
  VERSION=$(git rev-parse --short HEAD 2>/dev/null || echo "dev")
fi

echo "VERSION=$VERSION"
if [ -n "${GITHUB_OUTPUT:-}" ]; then
  echo "version=$VERSION" >> "$GITHUB_OUTPUT"
fi
