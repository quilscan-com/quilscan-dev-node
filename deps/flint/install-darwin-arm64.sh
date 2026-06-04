#!/usr/bin/env bash
set -euo pipefail

platform="${PLATFORM:-darwin-arm64}"
if [[ "$platform" != "darwin-arm64" ]]; then
  echo "unsupported FLINT platform: $platform" >&2
  exit 1
fi

dest_dir="${1:-vendor/flint}"
work_dir="${WORK_DIR:-flint-work}"
flint_repo="${FLINT_REPO:-https://github.com/flintlib/flint.git}"
flint_ref="${FLINT_REF:-flint-3.0}"
expected_commit="${FLINT_SOURCE_COMMIT:-}"

mkdir -p "$dest_dir" "$work_dir"
dest_dir="$(cd "$dest_dir" && pwd)"
work_dir="$(cd "$work_dir" && pwd)"

metadata_repo="$dest_dir/.quilscan-flint-source-repo"
metadata_ref="$dest_dir/.quilscan-flint-source-ref"
metadata_commit="$dest_dir/.quilscan-flint-source-commit"

if [[ -f "$dest_dir/lib/libflint.a" && -f "$metadata_repo" && -f "$metadata_ref" && -f "$metadata_commit" ]]; then
  cached_repo="$(cat "$metadata_repo")"
  cached_ref="$(cat "$metadata_ref")"
  cached_commit="$(cat "$metadata_commit")"
  if [[ "$cached_repo" == "$flint_repo" && "$cached_ref" == "$flint_ref" && ( -z "$expected_commit" || "$cached_commit" == "$expected_commit" ) ]]; then
    echo "Using cached FLINT dependency at $dest_dir from $cached_commit"
    exit 0
  fi
fi

src_dir="$work_dir/source"
rm -rf "$src_dir" "$dest_dir"
mkdir -p "$dest_dir"

git clone --branch "$flint_ref" --depth 1 "$flint_repo" "$src_dir"
source_commit="$(git -C "$src_dir" rev-parse HEAD)"
if [[ -n "$expected_commit" && "$source_commit" != "$expected_commit" ]]; then
  echo "FLINT commit mismatch: expected $expected_commit, got $source_commit" >&2
  exit 1
fi

cd "$src_dir"
./bootstrap.sh
./configure \
  --prefix="$dest_dir" \
  --enable-static \
  --disable-shared \
  --with-gmp="$(brew --prefix gmp)" \
  --with-mpfr="$(brew --prefix mpfr)" \
  CFLAGS="-O3 -fPIC"
make -j"$(sysctl -n hw.ncpu)"
make install

test -f "$dest_dir/lib/libflint.a"
printf "%s\n" "$flint_repo" > "$metadata_repo"
printf "%s\n" "$flint_ref" > "$metadata_ref"
printf "%s\n" "$source_commit" > "$metadata_commit"

echo "Installed FLINT dependency into $dest_dir from $source_commit"
