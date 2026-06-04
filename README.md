# Quilscan Dev Node Builds

Public GitHub Actions workflows for building Quilscan Dev Node binaries from
Quilibrium upstream source.

Each build records:

- upstream repo, branch, and commit
- platform
- binary SHA256
- GitHub workflow run URL
- Ed25519 signature for the binary

Default upstream:

```text
repo: https://github.com/QuilibriumNetwork/monorepo.git
branch: v2.1.0.23
```

## Workflows

### Build Linux Warm Builder

Creates the Linux warm Docker image used by later Linux node builds.

Upload the workflow artifact files to:

```text
https://releases.quilscan.com/deps/linux-warm-builder/amd64/
```

Required files:

```text
quilibrium-linux-warm-builder-linux-amd64.tar.zst
manifest.json
manifest.json.sig
```

### Build Quilibrium Dev Node

Builds the Dev Node binary for one platform.

Supported platforms:

```text
linux-amd64
darwin-arm64
```

The workflow artifact contains:

```text
node-<version>-<platform>
node-<version>-<platform>.sig
SHA256SUMS
build-info.json
public-key.txt
```

## Signing

The `.sig` file is a base64 Ed25519 signature over the exact node binary bytes.

The private key is stored in GitHub Secrets:

```text
DEV_NODE_SIGNING_PRIVATE_KEY
```

`build-info.json` is for transparency only. The signature proves the binary
itself was signed by the private key.

## Reproduce Locally

Linux:

```bash
./scripts/reproduce.sh linux-amd64
```

macOS Apple Silicon:

```bash
./scripts/reproduce.sh darwin-arm64
```

Compare the local SHA256 with `SHA256SUMS` from the public workflow artifact.
