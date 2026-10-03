#!/usr/bin/env bash
set -euo pipefail

CONF_DIR="/etc/sing-box"
CONF="$CONF_DIR/server.json"
SB="/usr/local/bin/sing-box"

[[ -x "$SB" ]] || { echo "sing-box is not installed: $SB" >&2; exit 1; }
sudo mkdir -p "$CONF_DIR"

# Regenerating keys invalidates every existing client link, so require intent.
if [[ -s "$CONF" && "${FORCE:-0}" != "1" ]]; then
  echo "$CONF already exists. Re-run with FORCE=1 to replace it (existing client links stop working)." >&2
  exit 1
fi

read -r -p "REALITY handshake host (hostname only): " HANDSHAKE_HOST

if [[ ! "$HANDSHAKE_HOST" =~ ^[A-Za-z0-9.-]+$ || "$HANDSHAKE_HOST" == .* || "$HANDSHAKE_HOST" == *. || "$HANDSHAKE_HOST" != *.* ]]; then
  echo "Invalid hostname. Example: www.example.com" >&2
  exit 1
fi

UUID="$(cat /proc/sys/kernel/random/uuid)"
KEYPAIR="$(sudo "$SB" generate reality-keypair)"
PRIVATE_KEY="$(printf '%s\n' "$KEYPAIR" | awk -F': ' '/PrivateKey/{print $2}')"
PUBLIC_KEY="$(printf '%s\n' "$KEYPAIR" | awk -F': ' '/PublicKey/{print $2}')"
SHORT_ID="$(openssl rand -hex 4)"

[[ -n "$PRIVATE_KEY" && -n "$PUBLIC_KEY" ]] || {
  echo "Failed to generate REALITY keypair." >&2
  exit 1
}

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
      "users": [
        { "uuid": "$UUID", "flow": "xtls-rprx-vision" }
      ],
      "tls": {
        "enabled": true,
        "server_name": "$HANDSHAKE_HOST",
        "reality": {
          "enabled": true,
          "handshake": {
            "server": "$HANDSHAKE_HOST",
            "server_port": 443
          },
          "private_key": "$PRIVATE_KEY",
          "short_id": ["$SHORT_ID"]
        }
      }
    }
  ],
  "outbounds": [
    { "type": "direct", "tag": "direct" }
  ],
  "route": { "final": "direct" }
}
JSON

sudo chmod 600 "$CONF"
sudo "$SB" check -c "$CONF"

PUBLIC_IP="${PUBLIC_IP:-}"
if [[ -z "$PUBLIC_IP" ]]; then
  for url in https://api.ipify.org https://ifconfig.me/ip https://icanhazip.com; do
    candidate="$(curl -4 -fsS --max-time 8 "$url" 2>/dev/null | tr -d '[:space:]' || true)"
    if [[ "$candidate" =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]]; then
      PUBLIC_IP="$candidate"
      break
    fi
  done
fi
LINK_HOST="${PUBLIC_IP:-REPLACE_WITH_SERVER_IP}"
LINK="vless://${UUID}@${LINK_HOST}:443?encryption=none&flow=xtls-rprx-vision&security=reality&sni=${HANDSHAKE_HOST}&fp=chrome&pbk=${PUBLIC_KEY}&sid=${SHORT_ID}&type=tcp#Jamshidix"
printf '%s\n' "$LINK" | sudo tee "$CONF_DIR/client-link.txt" >/dev/null
sudo chmod 600 "$CONF_DIR/client-link.txt"

cat <<OUT

SERVER CONFIG READY
UUID=$UUID
PUBLIC_KEY=$PUBLIC_KEY
SHORT_ID=$SHORT_ID
HANDSHAKE_HOST=$HANDSHAKE_HOST

Private key was written only to $CONF.
Do not commit $CONF or the link below to Git.

CLIENT LINK (copy this whole line, then double-click Jamshidix.exe on Windows):

$LINK

Saved (root only) in $CONF_DIR/client-link.txt

Next:
  sudo systemctl enable --now sing-box
  sudo systemctl status sing-box --no-pager
OUT
