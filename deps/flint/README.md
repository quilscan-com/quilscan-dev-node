# FLINT macOS Install

This directory contains the macOS Apple Silicon install script for the static
FLINT dependency used by the dev-node workflow.

`build-dev-node.yml` resolves the configured FLINT source ref, restores the
`vendor/flint` cache when available, and otherwise builds `libflint.a` from
source into `vendor/flint`.

Default source:

```text
repo: https://github.com/flintlib/flint.git
ref: flint-3.0
```

The install script records source metadata next to the installed dependency:

```text
.quilscan-flint-source-repo
.quilscan-flint-source-ref
.quilscan-flint-source-commit
```
