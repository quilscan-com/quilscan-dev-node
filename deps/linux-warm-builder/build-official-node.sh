#!/usr/bin/env bash
set -euo pipefail

repo_dir="${1:-.}"
repo_dir="$(cd "$repo_dir" && pwd)"

cd "$repo_dir"

cargo build --release \
  -p channel \
  -p vdf \
  -p ferret \
  -p bls48581 \
  -p rpm \
  -p verenc \
  -p ed448-bulletproofs

for dir in channel vdf ferret bls48581 rpm verenc bulletproofs; do
  uniffi-bindgen-go "crates/$dir/src/lib.udl" -o "$dir/generated"
done

./node/build.sh
test -f node/build/amd64_linux/node
