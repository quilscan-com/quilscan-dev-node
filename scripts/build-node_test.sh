#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fixture_root="$(mktemp -d)"
trap 'rm -rf "$fixture_root"' EXIT

upstream_dir="$fixture_root/upstream"
dist_dir="$fixture_root/dist"
mkdir -p "$upstream_dir/node" "$dist_dir"

cat > "$upstream_dir/node/build.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

mkdir -p node/build/arm64_macos
printf 'official-rust-node\n' > node/build/arm64_macos/node
chmod +x node/build/arm64_macos/node
EOF
chmod +x "$upstream_dir/node/build.sh"

git -C "$upstream_dir" init -q
git -C "$upstream_dir" config user.name "build-node test"
git -C "$upstream_dir" config user.email "build-node-test@example.invalid"
git -C "$upstream_dir" add node/build.sh
git -C "$upstream_dir" commit -q -m "fixture"

upstream_commit="$(git -C "$upstream_dir" rev-parse HEAD)"
expected_artifact="$dist_dir/node-2.1.0.25-${upstream_commit:0:7}-darwin-arm64"

UPSTREAM_BASE_VERSION=v2.1.0.25 \
  "$repo_root/scripts/build-node.sh" darwin-arm64 "$upstream_dir" "$dist_dir"

test -x "$expected_artifact"
test "$(cat "$expected_artifact")" = "official-rust-node"

echo "build-node macOS official-entrypoint test passed"
