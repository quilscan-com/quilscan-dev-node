#!/usr/bin/env bash
set -euo pipefail

platform="${1:?platform is required: linux-amd64 or darwin-arm64}"
repo_dir="${2:-upstream}"
dist_dir="${3:-dist}"
skip_build="${4:-}"

repo_dir="$(cd "$repo_dir" && pwd)"
mkdir -p "$dist_dir"
dist_dir="$(cd "$dist_dir" && pwd)"

run_official_macos_node_build() {
  vdf/generate.sh
  bls48581/generate.sh
  verenc/generate.sh
  bulletproofs/generate.sh
  ferret/generate.sh
  channel/generate.sh
  rpm/generate.sh
  node/build.sh -o build/arm64_macos/node
}

cd "$repo_dir"
upstream_commit="$(git rev-parse HEAD)"
short_commit="${upstream_commit:0:7}"

case "$platform" in
  linux-amd64)
    if [[ "$skip_build" != "--skip-build" ]]; then
      if command -v task >/dev/null 2>&1; then
        task build_node_amd64_linux
      else
        docker build \
          --ulimit stack=-1:-1 \
          --platform linux/amd64 \
          -f docker/Dockerfile.source \
          --output node/build/amd64_linux \
          --target=node \
          .
      fi
    fi
    src="node/build/amd64_linux/node"
    ;;
  darwin-arm64)
    if [[ "$skip_build" != "--skip-build" ]]; then
      run_official_macos_node_build
    fi
    src="node/build/arm64_macos/node"
    if [[ ! -f "$src" && -f "build/arm64_macos/node" ]]; then
      src="build/arm64_macos/node"
    fi
    ;;
  *)
    echo "unsupported platform: $platform" >&2
    exit 1
    ;;
esac

if [[ ! -f "$src" ]]; then
  echo "built node binary not found: $src" >&2
  exit 1
fi

base_version="${UPSTREAM_BASE_VERSION:-${GITHUB_REF_NAME:-dev}}"
base_version="${base_version#v}"
name="node-${base_version}-${short_commit}-${platform}"
cp "$src" "$dist_dir/$name"
chmod +x "$dist_dir/$name"

echo "built $dist_dir/$name"
