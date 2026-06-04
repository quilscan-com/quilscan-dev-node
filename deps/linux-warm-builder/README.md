# Linux Warm Builder Dependency

This directory contains scripts for a Linux AMD64 warm builder image used by
public Quilibrium dev-node builds.

The warm builder image starts from Ubuntu, builds the required Linux toolchain
and native libraries, then keeps a previously built upstream source checkout
at:

```text
/opt/monorepo
```

That checkout includes `.git`, generated bindings, and Cargo `target/` state
from a successful prior build. The dev-node workflow loads the signed image,
fetches the requested upstream branch inside `/opt/monorepo`, runs the local
expanded form of the upstream Docker `gen-rust` and `build-node` stages, and
copies only the final node binary out through a mounted output directory.

Default bucket path:

```text
https://releases.quilscan.com/deps/linux-warm-builder/amd64
```

Upload these files from the `Build Linux Warm Builder` workflow artifact:

```text
quilibrium-linux-warm-builder-linux-amd64.tar.zst
manifest.json
manifest.json.sig
```

The dev-node workflow verifies `manifest.json` using
`keys/build-signing-public-key.txt`, verifies the archive SHA256, runs
`docker load`, and then builds the node binary inside the loaded warm image.

This is not a fully clean build path. It is an incremental public build path
that reuses Cargo build state from a signed, public workflow artifact.
