#!/usr/bin/env bash
set -euo pipefail

base_url="${1:?base URL is required}"
download_dir="${2:-linux-warm-builder-download}"
platform="${PLATFORM:-linux-amd64}"
public_key_path="${PUBLIC_KEY_PATH:-keys/build-signing-public-key.txt}"
image_tag="${LINUX_WARM_BUILDER_IMAGE_TAG:-quilibrium-linux-warm-builder:linux-amd64-v1}"

base_url="${base_url%/}"
mkdir -p "$download_dir"

curl -fsSL --retry 3 "$base_url/manifest.json" -o "$download_dir/manifest.json"
curl -fsSL --retry 3 "$base_url/manifest.json.sig" -o "$download_dir/manifest.json.sig"

archive="$(go run ./cmd/dependency-manifest-verify \
  -manifest "$download_dir/manifest.json" \
  -signature "$download_dir/manifest.json.sig" \
  -public-key "$public_key_path" \
  -expect-name quilibrium-linux-warm-builder \
  -expect-platform "$platform" \
  -print-archive)"
case "$archive" in
  ""|*/*|*..*)
    echo "unsafe archive name in manifest: $archive" >&2
    exit 1
    ;;
esac

curl -fsSL --retry 3 "$base_url/$archive" -o "$download_dir/$archive"

go run ./cmd/dependency-manifest-verify \
  -manifest "$download_dir/manifest.json" \
  -signature "$download_dir/manifest.json.sig" \
  -public-key "$public_key_path" \
  -archive-dir "$download_dir" \
  -expect-name quilibrium-linux-warm-builder \
  -expect-platform "$platform"

zstd -dc "$download_dir/$archive" | docker load
docker image inspect "$image_tag" >/dev/null

echo "Loaded signed Linux warm builder image $image_tag"
