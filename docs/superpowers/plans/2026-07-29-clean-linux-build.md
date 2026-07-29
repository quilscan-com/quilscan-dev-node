# Clean Linux Dev Node Build Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a selected upstream Quilibrium branch on a fresh GitHub `ubuntu-24.04` runner by invoking the upstream `task build_node_amd64_linux` target.

**Architecture:** Keep the existing workflow and release pipeline, but replace only the Linux warm-builder preparation and execution steps. The clean job clones upstream into the ephemeral runner, installs a pinned Task binary, invokes the upstream Docker build through Task, and passes the resulting node binary into the existing packaging and signing steps.

**Tech Stack:** GitHub Actions, Ubuntu 24.04, Git LFS, Go Task, Docker BuildKit, Bash

---

### Task 1: Replace the warm Linux build

**Files:**
- Modify: `.github/workflows/build-dev-node.yml:31`

- [ ] **Step 1: Confirm the current workflow is incremental**

Run:

```bash
rg -n "warm builder|LINUX_WARM_BUILDER|quilscan-build-node-linux-official-local" \
  .github/workflows/build-dev-node.yml
```

Expected: matches for the signed warm-builder download, Docker image, and local
warm-build helper.

- [ ] **Step 2: Add a fresh upstream checkout**

Insert this Linux-only step after runner disk cleanup:

```yaml
      - name: Checkout upstream
        run: |
          set -euo pipefail
          git clone \
            --branch "${{ inputs.upstream_branch }}" \
            --single-branch \
            "${{ inputs.upstream_repo }}" \
            upstream
          cd upstream
          git lfs install --local
          git lfs pull
          git rev-parse HEAD > ../UPSTREAM_COMMIT
          git status --short
```

- [ ] **Step 3: Install a pinned Task binary**

Add this step after `actions/setup-go`:

```yaml
      - name: Install Task
        run: |
          set -euo pipefail
          go install github.com/go-task/task/v3/cmd/task@v3.39.2
          task --version
```

- [ ] **Step 4: Invoke the official upstream Task**

Replace the signed warm-builder download and Docker run with:

```yaml
      - name: Build node with official upstream Task
        run: |
          set -euo pipefail
          cd upstream
          task build_node_amd64_linux
          test -x node/build/amd64_linux/node
          mkdir -p ../clean-output
          cp node/build/amd64_linux/node ../clean-output/node
          chmod +x ../clean-output/node
```

Update the existing packaging step to copy `clean-output/node`.

- [ ] **Step 5: Correct Linux build metadata**

Replace the Linux build-info warm-builder field with:

```json
"linux_build_method": "fresh_github_runner_official_task"
```

Keep `upstream_repo`, `upstream_branch`, `upstream_commit`, platform, workflow
run URL, checksums, signature, and public key unchanged.

- [ ] **Step 6: Validate workflow invariants**

Run:

```bash
ruby -e 'require "yaml"; YAML.load_file(".github/workflows/build-dev-node.yml", aliases: true); puts "yaml ok"'
rg -n "task build_node_amd64_linux|clean-output/node|fresh_github_runner_official_task" \
  .github/workflows/build-dev-node.yml
if sed -n '52,190p' .github/workflows/build-dev-node.yml |
  rg "download-signed|docker load|quilscan-build-node-linux-official-local"; then
  exit 1
fi
git diff --check
```

Expected: YAML parses, all three clean-build markers exist, no warm-builder
execution remains in the Linux job, and `git diff --check` reports nothing.

- [ ] **Step 7: Commit the workflow**

```bash
git add .github/workflows/build-dev-node.yml
git commit -m "Build Linux dev node in clean runner"
```

### Task 2: Publish and exercise the test branch

**Files:**
- Verify: `.github/workflows/build-dev-node.yml`

- [ ] **Step 1: Push the isolated branch**

Run:

```bash
git push -u origin test
```

Expected: remote branch `test` points at the clean workflow commit.

- [ ] **Step 2: Dispatch a clean `.25` build**

Run:

```bash
gh workflow run build-dev-node.yml \
  --ref test \
  -f upstream_repo=https://github.com/QuilibriumNetwork/monorepo.git \
  -f upstream_branch=v2.1.0.25 \
  -f platform=linux-amd64 \
  -f dispatch_id=clean-runner-test
```

Expected: GitHub accepts the dispatch for the `test` ref.

- [ ] **Step 3: Inspect the workflow result**

Run:

```bash
gh run list \
  --workflow build-dev-node.yml \
  --branch test \
  --limit 1
```

Expected: the newest run uses branch `test`; when complete, its Linux job shows
the fresh checkout and `task build_node_amd64_linux` steps.
