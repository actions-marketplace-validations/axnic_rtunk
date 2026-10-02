# Installation

rtunk ships no prebuilt binary yet: build it from source with Go on macOS or Linux.

```bash
go install github.com/xunleii/rtunk/cmd/rtunk@latest
rtunk --version
```

`--version` prints `dev`: expected, not an error. A local `go build` or `go install` carries no
version metadata, and only a tagged, properly built release reports its real version
(`resolveVersion()` in [`cmd/rtunk/main.go`](../cmd/rtunk/main.go)). There is no `.goreleaser.yml`,
no `Makefile` and no GitHub Releases artifact, so there is no "just download it" option today.

## Check that your platform is supported

macOS and Linux only. Windows is not a supported host platform, a deliberate divergence recorded in
[ROADMAP.md](../ROADMAP.md#deliberate-divergences-from-trunk), so neither path below is verified
there. Both paths work on any macOS or Linux architecture Go supports (for example `darwin/arm64`,
`darwin/amd64`, `linux/amd64`, `linux/arm64`), since Go cross-compiles.

## Install with `go install`

This is the quickest path if you already have Go on `PATH`. It requires Go 1.27.0 or later, the
floor declared in [`go.mod`](../go.mod).

```bash
go install github.com/xunleii/rtunk/cmd/rtunk@latest
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
git clone https://github.com/xunleii/rtunk.git
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
