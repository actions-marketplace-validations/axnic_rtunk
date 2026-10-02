# Installation

rtunk installs from a release archive, with a version manager (mise, aqua) or from source with Go, on macOS or Linux.

```bash
go install github.com/axnic/rtunk/cmd/rtunk@latest
rtunk --version
```

`--version` reports the release tag for a release archive. A local `go build` carries no version
metadata and prints `dev`: expected, not an error (`resolveVersion()` in
[`cmd/rtunk/main.go`](../cmd/rtunk/main.go)). Releases are built by the
[Release workflow](https://github.com/axnic/rtunk/blob/main/.github/workflows/workflow_dispatch.release.yaml)
with [GoReleaser](https://goreleaser.com) ([`.goreleaser.yml`](../.goreleaser.yml)) and published on
the [Releases page](https://github.com/axnic/rtunk/releases).

## Check that your platform is supported

macOS and Linux only. Windows is not a supported host platform, a deliberate divergence recorded in
[ROADMAP.md](../ROADMAP.md#deliberate-divergences-from-trunk), so neither path below is verified
there. Release archives exist for `darwin/amd64`, `darwin/arm64`, `linux/amd64` and `linux/arm64`;
building from source works on any macOS or Linux architecture Go supports, since Go cross-compiles.

## Install from a release archive

Each release at <https://github.com/axnic/rtunk/releases> ships one archive per platform, named
`rtunk-<tag>-<os>-<arch>.tar.gz` (for example `rtunk-v0.13.2-darwin-arm64.tar.gz`), plus
`checksums.txt`. Download the archive for your platform and `checksums.txt`, check the archive
against the checksums, then extract the `rtunk` binary onto `PATH`:

```bash
tag=v0.13.2 os=darwin arch=arm64    # os: darwin|linux, arch: amd64|arm64
gh release download "$tag" -R axnic/rtunk -p "rtunk-$tag-$os-$arch.tar.gz" -p checksums.txt
sha256sum --ignore-missing -c checksums.txt        # shasum -a 256 -c on macOS
tar -xzf "rtunk-$tag-$os-$arch.tar.gz" rtunk
install -m 0755 rtunk ~/.local/bin/rtunk           # any directory on PATH
rtunk --version
```

The `sha256sum` check only proves the archive matches `checksums.txt`; it says nothing about who
produced that file. Releases signed with cosign and carrying SLSA build provenance can also be
verified against the Release workflow's identity: the steps are in
[SECURITY.md](../SECURITY.md#verifying-a-release). Releases published before signing was added
(`v0.13.0` to `v0.13.2`) carry `checksums.txt` only, so the checksum comparison is the only check
available for them.

## Install with a version manager

rtunk is not in the central registry of any version manager yet, so both tools below read the
GitHub release archives directly. Pin an exact version: it is what makes the install reproducible.

### mise

[mise](https://mise.jdx.dev) installs release archives through its `github` backend. In
`mise.toml` (or `.mise.toml`):

```toml
[tools]
"github:axnic/rtunk" = "0.13.2"
```

or one-off, from the command line:

```bash
mise use "github:axnic/rtunk@0.13.2"      # adds it to mise.toml and installs it
mise exec "github:axnic/rtunk@0.13.2" -- rtunk --version
```

Run `mise lock` to record the archive checksums in `mise.lock`, so every machine and CI run
installs the same bytes.

### aqua

[aqua](https://aquaproj.github.io) has no `axnic/rtunk` entry in its standard registry, so declare
the package in a local registry. `registry.yaml`:

```yaml
packages:
  - type: github_release
    repo_owner: axnic
    repo_name: rtunk
    asset: rtunk-{{.Version}}-{{.OS}}-{{.Arch}}.tar.gz
    format: tar.gz
    files:
      - name: rtunk
    checksum:
      type: github_release
      asset: checksums.txt
      algorithm: sha256
```

`aqua.yaml`:

```yaml
registries:
  - type: local
    name: rtunk
    path: registry.yaml
packages:
  - name: axnic/rtunk@v0.13.2
    registry: rtunk
```

aqua refuses packages from a registry it has not been told to trust, so add the registry to your
[policy file](https://aquaproj.github.io/docs/reference/security/policy) (`aqua-policy.yaml`, then
`aqua policy allow`):

```yaml
registries:
  - type: local
    name: rtunk
    path: registry.yaml
packages:
  - registry: rtunk
```

Then `aqua i` installs rtunk and verifies the archive against `checksums.txt`.

Neither tool checks the cosign signature or the build provenance: for that, follow
[SECURITY.md](../SECURITY.md#verifying-a-release) on a downloaded archive. Releases before signing
(`v0.13.0` to `v0.13.2`) have no signature to check anyway.

## Install with `go install`

This is the quickest path if you already have Go on `PATH`. It requires Go 1.27.0 or later, the
floor declared in [`go.mod`](../go.mod).

```bash
go install github.com/axnic/rtunk/cmd/rtunk@latest
```

This installs `rtunk` to `$(go env GOBIN)`, or to `$(go env GOPATH)/bin` if `GOBIN` is unset. Make
sure that directory is on `PATH`, then verify:

```bash
rtunk --version
rtunk help
```

`help` prints the full command list; [Command Reference](Command-Reference.md) is the authoritative
description of each command.

## Build from a clone (contributor path)

Use this path to contribute, or to get the toolchain versions the project develops against. It is
the setup [CONTRIBUTING.md](../CONTRIBUTING.md) assumes, and it requires
[mise](https://mise.jdx.dev/).

```bash
git clone https://github.com/axnic/rtunk.git
cd rtunk
mise trust    # if mise prompts about this repo's .mise.toml
mise install  # installs the Go toolchain and dev tools .mise.toml declares
go build -o rtunk ./cmd/rtunk
./rtunk --version
```

`mise install` resolves the toolchain declared in [`.mise.toml`](../.mise.toml): the Go compiler
(matching the `go 1.27.0` floor in `go.mod`) and the dev tools. rtunk dogfoods itself: the binary
you just built runs the lint stack of [`.rtunk/rtunk.yaml`](../.rtunk/rtunk.yaml) on its own source
and docs (`./rtunk fmt`, `./rtunk check`, or `mise run rtunk`), downloading those linters into
its cache on first use. The output of `--version` and `help` is the same as with `go install`.

## Upgrade rtunk

Repeat the install option you used: re-run `go install .../rtunk@latest`, or `git pull` and
rebuild. There is no `rtunk upgrade` command, by design: rtunk stays a local tool that makes no
network call you did not explicitly trigger.

## Where to go next

- [Quickstart](Quickstart.md) — run your first `rtunk check` in a few minutes
- [Migrating from trunk](Migrating-From-Trunk.md) — switch an existing trunk repository to rtunk
- [Command Reference](Command-Reference.md) — every command and flag
