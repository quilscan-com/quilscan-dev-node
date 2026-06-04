# Linux Warm Builder Image Design

## Goal

Speed up public Linux dev-node builds while preserving a clear trust story.

The Linux node build uses a signed warm builder image as its only GitHub
Actions build path. The workflow does not use Docker buildx fallback paths or
GitHub Actions build caches.

## Current State

The warm builder image is now the Linux base image. It contains:

- Ubuntu base system
- Go and Rust toolchains
- GMP, FLINT, EMP, and build tools
- A cloned upstream monorepo at `/opt/monorepo`
- A successful prior expanded upstream Docker `gen-rust` and `build-node` result
- Generated bindings from the upstream Rust leaf crates
- The Cargo `target/` cache from that prior build

## Proposed Architecture

Use a single Linux warm builder image flow.

The warm builder image contains:

- A cloned upstream monorepo at `/opt/monorepo`
- The `.git` directory for that repository
- A successful prior build output
- Generated bindings from the upstream Docker `gen-rust` stage
- The Cargo `target/` cache from that prior build

The image is exported, its manifest is signed, and it is uploaded to:

```text
https://releases.quilscan.com/deps/linux-warm-builder/amd64/
```

## Build Flow

### Warm Image Build

The warm-image workflow should:

1. Clone the configured upstream repository and branch.
2. Run Git LFS pull.
3. Build the warm image from zero in Docker.
4. Run the local expanded form of the upstream Docker `gen-rust` stage.
5. Run the local expanded form of the upstream Docker `build-node` stage.
6. Keep `/opt/monorepo/.git`, generated bindings, `/opt/monorepo/target`, and build outputs in the image.
7. Export the image as a compressed tar archive.
8. Write a manifest containing source repo, branch, commit, archive SHA, image tag, and workflow run URL.
9. Sign the manifest with the repository signing private key.
10. Upload the archive, manifest, and signature as a GitHub artifact for manual upload to the bucket.

### Dev Node Build From Warm Image

The dev-node workflow should:

1. Download `manifest.json`, `manifest.json.sig`, and the warm image archive.
2. Verify the manifest signature with the public key in the repository.
3. Verify the warm image archive SHA from the manifest.
4. Load the image with Docker.
5. Run the image without mounting over `/opt/monorepo`.
6. Inside the container:

```bash
cd /opt/monorepo
git fetch origin
git checkout "$UPSTREAM_BRANCH"
git pull --ff-only
git lfs pull
quilscan-build-node-linux-official-local .
cp node/build/amd64_linux/node /out/node
```

7. Mount only the output directory from the host:

```bash
-v "$PWD/warm-output:/out"
```

8. Package, checksum, sign, and upload the final node binary.

## Trust Model

The warm builder image improves speed but is not as clean as a fresh source build.

The public trust statement should be precise:

- The warm image is built by a public GitHub Actions workflow.
- The warm image manifest is signed.
- The final build still fetches the configured upstream branch and runs the
  expanded local form of the upstream Linux Docker `gen-rust` and `build-node`
  stages.
- Cargo decides which cached artifacts are reusable based on its fingerprints.
- A fresh reproduction check should use a separate local or public clean-build
  workflow, not the default Linux node workflow.

Do not describe warm-image builds as fully clean builds.

## Risks

The main risk is stale or incompatible cached build state.

Mitigations:

- Keep the upstream commit used to create the warm image in the manifest.
- Keep the final upstream commit in the node build metadata.
- Rebuild the warm image when upstream changes significantly.

## Recommended Naming

Use one warm builder input:

```text
linux_warm_builder_base_url
```

Use one warm builder bucket path:

```text
deps/linux-warm-builder/amd64
```
