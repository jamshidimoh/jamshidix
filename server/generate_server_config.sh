#!/usr/bin/env bash
set -euo pipefail

CONF_DIR="/etc/sing-box"
CONF="$CONF_DIR/server.json"
SB="/usr/local/bin/sing-box"

[[ -x "$SB" ]] || { echo "sing-box is not installed: $SB" >&2; exit 1; }
sudo mkdir -p "$CONF_DIR"

read -r -p "REALITY handshake host (hostname only): " HANDSHAKE_HOST
[[ "$HANDSHAKE_HOST" != *"/"* ]] || { echo "Use hostname only, not URL" >&2; exit 1; }

UUID="$(cat /proc/sys/kernel/random/uuid)"
KEYPAIR="$(sudo "$SB" generate reality-keypair)"
PRIVATE_KEY="$(printf '%s\n' "$KEYPAIR" | awk -F': ' '/PrivateKey/{print $2}')"
PUBLIC_KEY="$(printf '%s\n' "$KEYPAIR" | awk -F': ' '/PublicKey/{print $2}')"
SHORT_ID="$(openssl rand -hex 4)"

[[ -n "$PRIVATE_KEY" && -n "$PUBLIC_KEY" ]] || { echo "Failed to generate REALITY keypair" >&2; exit 1; }

sudo tee "$CONF" >/dev/null <<JSON
{
  "\$schema": "https://sing-box.sagernet.org/schema.json",
  "log": { "level": "warn" },
  "inbounds": [
    {
      "type": "vless",
      "tag": "vless-reality-in",
      "listen": "0.0.0.0",
      "listen_port": 443,
      "users": [{ "uuid": "$UUID", "flow": "xtls-rprx-vision" }],
      "tls": {
        "enabled": true,
        "server_name": "$HANDSHAKE_HOST",
        "reality": {
          "enabled": true,
          "handshake": { "server": "$HANDSHAKE_HOST", "server_port": 443 },
          "private_key": "$PRIVATE_KEY",
          "short_id": ["$SHORT_ID"]
        }
      }
    }
  ],
  "outbounds": [{ "type": "direct", "tag": "direct" }],
  "route": { "final": "direct" }
}
JSON

sudo chmod 600 "$CONF"
sudo "$SB" check -c "$CONF"

cat <<OUT

SERVER CONFIG READY
UUID=$UUID
PUBLIC_KEY=$PUBLIC_KEY
SHORT_ID=$SHORT_ID
HANDSHAKE_HOST=$HANDSHAKE_HOST

Private key was written only to $CONF.
Do not commit $CONF to Git.

Next:
  sudo systemctl enable --now sing-box
  sudo systemctl status sing-box --no-pager
OUT
