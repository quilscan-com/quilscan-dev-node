# Linux Warm Builder Image Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Add a signed Linux warm builder image flow that reuses a prior `/opt/monorepo` checkout and Cargo target state without keeping Docker buildx fallback paths.

**Architecture:** Add a standalone `deps/linux-warm-builder` directory and `Build Linux Warm Builder` workflow. `build-dev-node.yml` has a `linux_warm_builder_base_url` input, loads the signed warm image, and builds inside `/opt/monorepo` while only mounting an output directory. The Linux build uses a local expansion of the upstream Docker `gen-rust` and `build-node` stages.

**Tech Stack:** GitHub Actions, Docker, Bash, zstd, Go manifest signing tools, Ed25519 signatures.

---

### Task 1: Warm Builder Scripts

**Files:**
- Create: `deps/linux-warm-builder/build-image.sh`
- Create: `deps/linux-warm-builder/download-signed.sh`
- Create: `deps/linux-warm-builder/README.md`

- [x] **Step 1: Create `build-image.sh`**

Create a script that copies the checked-out upstream source into a temporary Docker build context, builds from `deps/linux-warm-builder/Dockerfile`, copies the source into `/opt/monorepo`, runs the expanded upstream Docker `gen-rust` and `build-node` steps once, and exports a compressed image archive.

- [x] **Step 2: Create `download-signed.sh`**

Create a script matching the signed manifest verifier pattern, but expecting manifest name `quilibrium-linux-warm-builder` and loading image tag `quilibrium-linux-warm-builder:linux-amd64-v1`.

- [x] **Step 3: Create `README.md`**

Document the bucket path, uploaded files, and trust model for the warm builder image.

### Task 2: Warm Builder Workflow

**Files:**
- Create: `.github/workflows/build-linux-warm-builder.yml`

- [x] **Step 1: Create workflow dispatch inputs**

Use `upstream_repo`, `upstream_branch`, `builder_version`, and `platform`.

- [x] **Step 2: Build and sign artifact**

Checkout scripts, checkout upstream, free disk, install `zstd`, call `deps/linux-warm-builder/build-image.sh`, sign the archive manifest with name `quilibrium-linux-warm-builder`, and upload the image archive plus manifest files.

### Task 3: Dev Node Workflow Integration

**Files:**
- Modify: `.github/workflows/build-dev-node.yml`

- [x] **Step 1: Add input**

Add `linux_warm_builder_base_url` with default `https://releases.quilscan.com/deps/linux-warm-builder/amd64`.

- [x] **Step 2: Adjust Linux source checkout**

Skip external upstream checkout when `linux_warm_builder_base_url` is set. The warm image owns `/opt/monorepo`.

- [x] **Step 3: Add warm-image build path**

When `linux_warm_builder_base_url` is non-empty, download and verify the warm image, run `git fetch`, checkout the requested branch, `git pull --ff-only`, `git lfs pull`, run the expanded upstream Docker `gen-rust` and `build-node` steps, copy `/opt/monorepo/node/build/amd64_linux/node` to a mounted output directory, and write `UPSTREAM_COMMIT`.

- [x] **Step 4: Remove existing clean fallback path**

Linux node builds use the signed warm builder image path only.

- [x] **Step 5: Include warm URL in metadata**

Add `linux_warm_builder_base_url` to Linux and macOS `build-info.json`.

### Task 4: Verification

**Files:**
- No new files.

- [x] **Step 1: Run shell syntax checks**

Run `bash -n` over all new shell scripts.

- [x] **Step 2: Run Go tests**

Run `go test ./...`.

- [x] **Step 3: Inspect workflow YAML**

Read the changed workflow sections and confirm conditions are mutually exclusive.

- [x] **Step 4: Commit and push**

Commit implementation and push to `main`.
