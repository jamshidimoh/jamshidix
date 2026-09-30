#!/usr/bin/env bash
set -euo pipefail

: "${CLOUDFLARE_EMAIL:?Set CLOUDFLARE_EMAIL in the environment.}"
: "${CLOUDFLARE_API_KEY:?Set CLOUDFLARE_API_KEY in the environment.}"

response="$(curl --fail --silent --show-error \
  --request GET \
  --url "https://api.cloudflare.com/client/v4/user" \
  --header "X-Auth-Email: ${CLOUDFLARE_EMAIL}" \
  --header "X-Auth-Key: ${CLOUDFLARE_API_KEY}" \
  --header "Content-Type: application/json")"

python3 - "$response" <<'PY'
import json
import sys
d = json.loads(sys.argv[1])
if d.get("success") is not True:
    print("Cloudflare authentication failed.", file=sys.stderr)
    raise SystemExit(1)
print("Cloudflare authentication succeeded.")
PY
