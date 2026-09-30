#!/usr/bin/env bash
set -euo pipefail

VERSION="1.14.1"
BASE="https://github.com/SagerNet/sing-box/releases/download/v${VERSION}"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

curl --fail --silent --show-error --location \
  "$BASE/sing-box-${VERSION}-linux-amd64.tar.gz" \
  -o "$TMP/sing-box.tar.gz"
tar -xzf "$TMP/sing-box.tar.gz" -C "$TMP"

SB="$(find "$TMP" -type f -name sing-box -perm -u+x | head -n1)"
[[ -n "$SB" ]] || { echo "sing-box binary not found" >&2; exit 1; }

KEYS="$("$SB" generate reality-keypair)"
PRIVATE_KEY="$(printf '%s\n' "$KEYS" | awk -F': ' '/PrivateKey/{print $2}')"
PUBLIC_KEY="$(printf '%s\n' "$KEYS" | awk -F': ' '/PublicKey/{print $2}')"
UUID="$("$SB" generate uuid)"
SHORT_ID="0123abcd"

[[ -n "$PRIVATE_KEY" && -n "$PUBLIC_KEY" && -n "$UUID" ]] || {
  echo "Failed to generate validation credentials." >&2
  exit 1
}

CHECK_DIR="$TMP/config-check"
mkdir -p "$CHECK_DIR"
export PRIVATE_KEY PUBLIC_KEY UUID SHORT_ID CHECK_DIR

python3 - <<'PY'
import json
import os
from pathlib import Path

root = Path(".")
out = Path(os.environ["CHECK_DIR"])

server = json.loads((root / "config/server.config.template.json").read_text())
server["inbounds"][0]["users"][0]["uuid"] = os.environ["UUID"]
server["inbounds"][0]["tls"]["server_name"] = "www.cloudflare.com"
server["inbounds"][0]["tls"]["reality"]["handshake"]["server"] = "www.cloudflare.com"
server["inbounds"][0]["tls"]["reality"]["private_key"] = os.environ["PRIVATE_KEY"]
server["inbounds"][0]["tls"]["reality"]["short_id"] = [os.environ["SHORT_ID"]]
(out / "server.json").write_text(json.dumps(server), encoding="utf-8")

client = json.loads((root / "config/client.config.template.json").read_text())
client["outbounds"][0]["server"] = "192.0.2.1"
client["outbounds"][0]["uuid"] = os.environ["UUID"]
client["outbounds"][0]["tls"]["server_name"] = "www.cloudflare.com"
client["outbounds"][0]["tls"]["reality"]["public_key"] = os.environ["PUBLIC_KEY"]
client["outbounds"][0]["tls"]["reality"]["short_id"] = os.environ["SHORT_ID"]
(out / "client.json").write_text(json.dumps(client), encoding="utf-8")
PY

"$SB" check -c "$CHECK_DIR/server.json"
"$SB" check -c "$CHECK_DIR/client.json"

echo "sing-box ${VERSION} schema validation passed."
