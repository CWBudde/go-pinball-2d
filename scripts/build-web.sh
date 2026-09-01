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

required=(
    index.html
    main.wasm
    wasm_exec.js
    assets/README.md
    assets/images/favicon.png
    assets/images/background.png
    assets/images/logo.png
    assets/images/ball.png
    assets/images/flipper.png
    assets/images/bumper.png
    assets/images/post.png
    assets/images/target.png
    assets/images/lane-light.png
    assets/images/plunger.png
    assets/images/glow.png
    assets/images/particle.png
    assets/audio/flipper.wav
    assets/audio/bumper.wav
    assets/audio/target.wav
    assets/audio/launch.wav
    assets/audio/jackpot.wav
    assets/audio/drain.wav
    assets/audio/game-over.wav
)

for path in "${required[@]}"; do
    if [[ ! -s "${dist_dir}/${path}" ]]; then
        echo "web build is missing required non-empty file: ${path}" >&2
        exit 1
    fi
done

if [[ ! -f "${dist_dir}/.nojekyll" ]]; then
    echo "web build is missing .nojekyll" >&2
    exit 1
fi

go run ./cmd/verifydist -repo "${root_dir}" -dist dist

echo "Built browser game in ${dist_dir}"
