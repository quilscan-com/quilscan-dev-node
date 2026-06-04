#!/usr/bin/env bash
set -euo pipefail

platform="${PLATFORM:-linux-amd64}"
if [[ "$platform" != "linux-amd64" ]]; then
  echo "unsupported Linux warm builder platform: $platform" >&2
  exit 1
fi

upstream_dir="${1:?upstream source directory is required}"
dist_dir="${2:-dist}"
warm_image_tag="${LINUX_WARM_BUILDER_IMAGE_TAG:-quilibrium-linux-warm-builder:linux-amd64-v1}"
archive_name="${LINUX_WARM_BUILDER_ARCHIVE:-quilibrium-linux-warm-builder-linux-amd64.tar.zst}"

mkdir -p "$dist_dir"
dist_dir="$(cd "$dist_dir" && pwd)"
upstream_dir="$(cd "$upstream_dir" && pwd)"
context_dir="${RUNNER_TEMP:-/tmp}/quilscan-linux-warm-builder-context"

rm -rf "$context_dir"
mkdir -p "$context_dir/upstream"
cp deps/linux-warm-builder/build-official-node.sh "$context_dir/build-official-node.sh"

echo "[$(date -u '+%Y-%m-%dT%H:%M:%SZ')] Copying upstream checkout into warm image context"
(
  cd "$upstream_dir"
  tar \
    --exclude='./target' \
    --exclude='./node/build' \
    -cf - .
) | (
  cd "$context_dir/upstream"
  tar -xf -
)

echo "[$(date -u '+%Y-%m-%dT%H:%M:%SZ')] Building warm image $warm_image_tag from zero"
docker build \
  --platform linux/amd64 \
  --target linux-warm-builder \
  --file deps/linux-warm-builder/Dockerfile \
  --tag "$warm_image_tag" \
  "$context_dir"

echo "[$(date -u '+%Y-%m-%dT%H:%M:%SZ')] Saving and compressing $warm_image_tag to $archive_name"
docker image inspect "$warm_image_tag" >/dev/null
docker save "$warm_image_tag" | zstd -T0 -3 -f -o "$dist_dir/$archive_name"

printf "%s\n" "$archive_name" > "$dist_dir/linux-warm-builder-archive-name.txt"
printf "%s\n" "$warm_image_tag" > "$dist_dir/linux-warm-builder-image-tag.txt"

echo "[$(date -u '+%Y-%m-%dT%H:%M:%SZ')] Built $dist_dir/$archive_name for $warm_image_tag"
