#!/bin/bash
set -euo pipefail

VERSION="${1:-${VERSION:-}}"

# FNet version is maintained in the manifest
if [ -z "${VERSION}" ]; then
  VERSION=$(grep "^version" apps/fnet/fnos/manifest | awk -F'=' '{print $2}' | tr -d ' ')
fi

[ -z "${VERSION}" ] && { echo "Failed to resolve version from manifest" >&2; exit 1; }

echo "VERSION=$VERSION"
if [ -n "${GITHUB_OUTPUT:-}" ]; then
  echo "version=$VERSION" >> "$GITHUB_OUTPUT"
fi
