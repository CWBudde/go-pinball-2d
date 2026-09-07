#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
preview_wasm="$(mktemp "${TMPDIR:-/tmp}/neon-relay-preview.XXXXXX.wasm")"
trap 'rm -f "${preview_wasm}"' EXIT
cd "${root_dir}"

GOOS=js GOARCH=wasm go build -o "${preview_wasm}" ./cmd/renderpreview
runtime_dir="$(go env GOROOT)/lib/wasm"
# Go's WASM startup reserves limited space for argv/env. Keep the Node runner
# environment small; it needs neither a browser nor graphics system libraries.
env -i PATH="${PATH}" node "${runtime_dir}/wasm_exec_node.js" "${preview_wasm}" "$@"
