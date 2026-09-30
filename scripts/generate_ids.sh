#!/usr/bin/env bash
set -euo pipefail

command -v openssl >/dev/null || { echo "openssl is required" >&2; exit 1; }

UUID="$(cat /proc/sys/kernel/random/uuid 2>/dev/null || uuidgen)"
SHORT_ID="$(openssl rand -hex 4)"

printf 'UUID=%s\nSHORT_ID=%s\n' "$UUID" "$SHORT_ID"
echo
echo "Generate the REALITY keypair on the server:"
echo "  sing-box generate reality-keypair"
echo "Keep the private key on the server only."
