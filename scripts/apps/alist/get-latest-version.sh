#!/bin/bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
source "$REPO_ROOT/scripts/lib/gh-api.sh" 2>/dev/null || true

INPUT_VERSION="${1:-}"

if type gh_curl &>/dev/null; then
  TAG=$(gh_curl "https://api.github.com/repos/AlistGo/alist/releases/latest" | jq -r '.tag_name')
else
  TAG=$(curl -sL "https://api.github.com/repos/AlistGo/alist/releases/latest" | jq -r '.tag_name')
fi

if [ -n "$INPUT_VERSION" ]; then
  VERSION="$INPUT_VERSION"
else
  VERSION=$(echo "$TAG" | sed 's/^v//')
fi

[ -z "$VERSION" ] || [ "$VERSION" = "null" ] && { echo "Failed to resolve version for alist" >&2; exit 1; }

echo "VERSION=$VERSION"

if [ -n "${GITHUB_OUTPUT:-}" ]; then
  echo "version=$VERSION" >> "$GITHUB_OUTPUT"
fi
