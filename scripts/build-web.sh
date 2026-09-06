#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dist_dir="${root_dir}/dist"
root_wasm="${root_dir}/main.wasm"
root_exec="${root_dir}/wasm_exec.js"

cleanup() {
    rm -f "${root_wasm}" "${root_exec}"
}
trap cleanup EXIT

cd "${root_dir}"
go run ./cmd/genassets -check

go run github.com/gonutz/prototype/cmd/drawsm@v1.9.2 build

rm -rf "${dist_dir}"
mkdir -p "${dist_dir}"
cp "${root_dir}/index.html" "${dist_dir}/index.html"
cp "${root_wasm}" "${dist_dir}/main.wasm"
cp "${root_exec}" "${dist_dir}/wasm_exec.js"
cp -R "${root_dir}/assets" "${dist_dir}/assets"
touch "${dist_dir}/.nojekyll"

go run ./cmd/verifydist -repo "${root_dir}" -dist dist

echo "Built browser game in ${dist_dir}"
