#!/usr/bin/env bash
set -euo pipefail

# Uses a scoped API Token (never the Global API Key).
# Create one at https://dash.cloudflare.com/profile/api-tokens with the minimum permissions needed.
: "${CLOUDFLARE_API_TOKEN:?Set CLOUDFLARE_API_TOKEN in the environment.}"

response="$(curl --fail --silent --show-error \
  --request GET \
  --url "https://api.cloudflare.com/client/v4/user/tokens/verify" \
  --header "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  --header "Content-Type: application/json")"

python3 - "$response" <<'PY'
import json
import sys
d = json.loads(sys.argv[1])
if d.get("success") is not True or d.get("result", {}).get("status") != "active":
    print("Cloudflare token verification failed.", file=sys.stderr)
    raise SystemExit(1)
print("Cloudflare token is active.")
PY
