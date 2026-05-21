#!/bin/bash
set -euo pipefail

VERSION="${VERSION:-dev}"
ARCH="${ARCH:-x86}"

echo "==> Building FNet ${VERSION} (${ARCH})"

SRC_DIR="apps/fnet/src"
OUT_DIR="app_root"

mkdir -p "${OUT_DIR}/bin" "${OUT_DIR}/ui/images"

# Compile Go binary
case "${ARCH}" in
  x86) GOARCH=amd64 ;;
  arm) GOARCH=arm64 ;;
  *) echo "Unsupported arch: ${ARCH}" >&2; exit 1 ;;
esac
GOOS=linux GOARCH=${GOARCH} go build -ldflags="-s -w" -o "${OUT_DIR}/fnet" "${SRC_DIR}/main.go"

# Copy static files
cp -a apps/fnet/fnos/bin/fnet-server "${OUT_DIR}/bin/"
chmod +x "${OUT_DIR}/bin/fnet-server"
cp -a apps/fnet/fnos/ui/* "${OUT_DIR}/ui/" 2>/dev/null || true

cd "${OUT_DIR}"
tar -czf ../app.tgz .
