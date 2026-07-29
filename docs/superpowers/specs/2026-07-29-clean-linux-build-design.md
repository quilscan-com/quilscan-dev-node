# Clean Linux Dev Node Build

## Goal

Build a requested Quilibrium upstream branch in a fresh GitHub-hosted
`ubuntu-24.04` runner without loading the signed warm-builder image or reusing
its Cargo target directory.

## Selected Approach

Keep `.github/workflows/build-dev-node.yml` as the dispatch entry point, but
replace the Linux build steps on the `test` branch with a clean build:

1. Check out the Dev Node build repository.
2. Free unused runner disk space.
3. Install a pinned Go Task release.
4. Clone the requested upstream repository and branch into a new directory.
5. Pull Git LFS objects and record the exact upstream commit.
6. Replace only the unreachable `gmplib.org` GMP download URL in the upstream
   Dockerfile with the official GNU mirror at `ftp.gnu.org`.
7. Run `task build_node_amd64_linux` from the upstream checkout.
8. Package, checksum, sign, and upload the resulting binary using the existing
   release steps.

The macOS build remains unchanged.

## Alternatives Considered

- Rebuild the warm-builder image for every requested branch. This is clean but
  duplicates the official Docker build and adds a large intermediate artifact.
- Delete `/opt/monorepo/target` inside the existing warm-builder. This removes
  Cargo artifacts but still reuses an older toolchain and native-library image,
  so it does not provide a fully fresh official build environment.

## Inputs and Outputs

The workflow continues accepting `upstream_repo`, `upstream_branch`, and
`platform`. The Linux artifact naming, build metadata, checksums, signing, and
artifact upload format remain compatible with the current publishing pipeline.

The build metadata will identify the exact upstream commit, workflow run, clean
build method, and GMP mirror URL. It will no longer claim a warm-builder URL for
clean Linux builds.

## Failure Handling

The job fails if cloning, Git LFS, Task installation, the official Task build,
or binary discovery fails. No fallback to warm-builder is allowed because a
fallback would invalidate the clean-build test.

## Verification

- Validate the workflow YAML locally.
- Confirm the Linux job contains no warm-builder download or `docker load`.
- Confirm the build runs from a newly cloned upstream directory.
- Confirm the command is exactly `task build_node_amd64_linux`.
- Dispatch the workflow from the `test` branch with the desired upstream branch.
- Compare the produced upstream commit and SHA256 against build metadata.
