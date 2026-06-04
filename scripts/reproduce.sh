#!/usr/bin/env bash
set -euo pipefail

upstream_repo="${UPSTREAM_REPO:-https://github.com/QuilibriumNetwork/monorepo.git}"
upstream_branch="${UPSTREAM_BRANCH:-v2.1.0.23}"
platform="${1:-linux-amd64}"
work_dir="${WORK_DIR:-./reproduce-work}"

rm -rf "$work_dir"
mkdir -p "$work_dir"

git clone --branch "$upstream_branch" --single-branch "$upstream_repo" "$work_dir/upstream"
(
  cd "$work_dir/upstream"
  git rev-parse HEAD > ../UPSTREAM_COMMIT
)

UPSTREAM_BASE_VERSION="$upstream_branch" ./scripts/build-node.sh "$platform" "$work_dir/upstream" "$work_dir/dist"

if command -v sha256sum >/dev/null 2>&1; then
  sha256sum "$work_dir"/dist/node-*
else
  shasum -a 256 "$work_dir"/dist/node-*
fi

echo "Upstream repo: $upstream_repo"
echo "Upstream branch: $upstream_branch"
echo "Upstream commit: $(cat "$work_dir/UPSTREAM_COMMIT")"
echo "Artifact directory: $work_dir/dist"
