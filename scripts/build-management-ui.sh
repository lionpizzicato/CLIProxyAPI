#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
UI_DIR="$ROOT/frontend/management-center"
OUTPUT="$ROOT/internal/managementasset/management.html"

if [ ! -d "$UI_DIR/node_modules" ]; then
  (
    cd "$UI_DIR"
    npm ci
  )
fi

(
  cd "$UI_DIR"
  npm run build
)

cp "$UI_DIR/dist/index.html" "$OUTPUT"
echo "management UI synced to $OUTPUT"
