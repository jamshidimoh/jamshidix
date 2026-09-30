#!/usr/bin/env bash
set -euo pipefail

# Jamshidix - server bootstrap.
# Stable release pinned to sing-box v1.14.1.
SING_BOX_VERSION="1.14.1"
SING_BOX_BASE="https://github.com/SagerNet/sing-box/releases/download/v${SING_BOX_VERSION}"

sudo apt-get update
sudo DEBIAN_FRONTEND=noninteractive apt-get install -y curl ca-certificates jq openssl tar iptables

ARCH="$(dpkg --print-architecture)"
case "$ARCH" in
  arm64) SB_ARCH="arm64" ;;
  amd64) SB_ARCH="amd64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT
ARCHIVE="sing-box-${SING_BOX_VERSION}-linux-${SB_ARCH}.tar.gz"
URL="${SING_BOX_BASE}/${ARCHIVE}"

curl --fail --silent --show-error --location "$URL" -o "$TMP_DIR/$ARCHIVE"
tar -xzf "$TMP_DIR/$ARCHIVE" -C "$TMP_DIR"

SB_BIN="$(find "$TMP_DIR" -type f -name sing-box -perm -u+x | head -n 1)"
[[ -n "$SB_BIN" ]] || { echo "sing-box binary not found" >&2; exit 1; }

sudo install -m 0755 "$SB_BIN" /usr/local/bin/sing-box
sudo mkdir -p /etc/sing-box /var/lib/sing-box
sudo touch /etc/sing-box/server.json
sudo chmod 600 /etc/sing-box/server.json

# IPv4 forwarding/NAT only. The gateway does not expose a routed IPv6 subnet.
cat <<SYS | sudo tee /etc/sysctl.d/99-jamshidix.conf >/dev/null
net.ipv4.ip_forward=1
net.ipv6.conf.all.forwarding=0
SYS
sudo sysctl --system >/dev/null

WAN_IF="$(ip route show default | awk 'NR==1{print $5}')"
[[ -n "$WAN_IF" ]] || { echo "Could not detect WAN interface" >&2; exit 1; }

sudo iptables -t nat -C POSTROUTING -o "$WAN_IF" -j MASQUERADE 2>/dev/null || \
  sudo iptables -t nat -A POSTROUTING -o "$WAN_IF" -j MASQUERADE

sudo tee /etc/systemd/system/sing-box.service >/dev/null <<SERVICE
[Unit]
Description=Jamshidix sing-box gateway
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/sing-box run -c /etc/sing-box/server.json
Restart=on-failure
RestartSec=3
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/sing-box
CapabilityBoundingSet=CAP_NET_ADMIN CAP_NET_BIND_SERVICE
AmbientCapabilities=CAP_NET_ADMIN CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
SERVICE

sudo systemctl daemon-reload

echo "sing-box ${SING_BOX_VERSION} installed."
echo "Next: generate REALITY keys, render server.json, validate it, then enable the service."
