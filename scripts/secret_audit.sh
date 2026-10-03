#!/usr/bin/env bash
set -euo pipefail

patterns=(
  'BEGIN[[:space:]]+(RSA|OPENSSH|EC|DSA|PRIVATE) KEY'
  'ghp_[A-Za-z0-9_]+'
  'github_pat_[A-Za-z0-9_]+'
  'xox[baprs]-'
  'AKIA[0-9A-Z]{16}'
  'cfk_[A-Za-z0-9_-]{20,}'
  'vless://[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}@'
)

found=0
for pattern in "${patterns[@]}"; do
  matches="$(git grep -nEI "$pattern" -- . ':!.git' ':!cmd/jamshidix/link_test.go' || true)"
  if [[ -n "$matches" ]]; then
    printf '%s\n' "$matches"
    found=1
  fi
done

if [[ "$found" -ne 0 ]]; then
  echo "Potential secret material found in tracked files." >&2
  exit 1
fi

echo "Secret audit passed."
