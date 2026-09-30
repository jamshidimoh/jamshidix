#!/usr/bin/env bash
set -euo pipefail

# Jamshidix optional Cloudflare Tunnel client.
# Pinned to cloudflared 2026.9.3 with upstream SHA-256 verification.
VERSION="2026.9.3"
BASE="https://github.com/cloudflare/cloudflared/releases/download/${VERSION}"

ARCH="$(dpkg --print-architecture)"
case "$ARCH" in
  arm64)
    FILE="cloudflared-linux-arm64"
    EXPECTED_SHA256="3d97437c71848bd8df68041e12436b484a661d95073ea1937f01a845ce88faa3"
    ;;
  amd64)
    FILE="cloudflared-linux-amd64"
    EXPECTED_SHA256="03f1f25d1cc93b9ad6c60569d44060bc4f17ed97075760ed8cfca4b12dcd68cc"
    ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

curl --fail --silent --show-error --location \
  "$BASE/$FILE" -o "$TMP_DIR/cloudflared"

ACTUAL_SHA256="$(sha256sum "$TMP_DIR/cloudflared" | awk '{print $1}')"
if [[ "$ACTUAL_SHA256" != "$EXPECTED_SHA256" ]]; then
  echo "cloudflared SHA256 mismatch." >&2
  echo "Expected: $EXPECTED_SHA256" >&2
  echo "Actual:   $ACTUAL_SHA256" >&2
  exit 1
fi

install -m 0755 "$TMP_DIR/cloudflared" /usr/local/bin/cloudflared

/usr/local/bin/cloudflared --version
echo "cloudflared 2026.9.3 installed and hash verified."
echo "Tunnel activation is intentionally separate; do not commit a tunnel token."
