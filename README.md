<!-- markdownlint-disable MD033 -->

<h1 align="center">rtunk</h1>

<p align="center"><strong>One command for your linters, formatters and security scanners. Open source, 100% local.</strong></p>

<p align="center">
<a href="go.mod"><img alt="Go version" src="https://img.shields.io/github/go-mod/go-version/axnic/rtunk"></a>
<a href="LICENSE"><img alt="License: MIT" src="https://img.shields.io/github/license/axnic/rtunk"></a>
<a href="https://www.bestpractices.dev/projects/15174"><img alt="OpenSSF Best Practices" src="https://www.bestpractices.dev/projects/15174/badge"></a>
<a href="https://github.com/axnic/rtunk/wiki"><img alt="Docs: wiki" src="https://img.shields.io/badge/docs-wiki-blue"></a>
</p>

rtunk is a from-scratch rewrite of [trunk.io's Code Quality
CLI](https://docs.trunk.io/code-quality/overview): it orchestrates existing tools behind a single
declarative config.

```bash
go install github.com/axnic/rtunk/cmd/rtunk@latest   # needs Go on PATH; see Installation
cd your-git-repo && rtunk init   # on a terminal, pick yamllint in the linters picker
rtunk check
```

```console
$ rtunk check
▲ yamllint         done     4 issues
.rtunk/rtunk.yaml  (1)
  1:1  medium  missing document start "---"  yamllint/document-start

example.yaml  (3)
  1:1   medium  missing document start "---"  yamllint/document-start
  1:8   high    too many spaces after colon   yamllint/colons
  2:11  high    too many spaces after comma   yamllint/commas

Checked 2 files with 1 linter in 0.2s
✖ 4 issues (2 high · 2 medium · 0 low)
```

Other ways to install: a release archive, mise, aqua, or the GitHub Action, see
[Installation](https://github.com/axnic/rtunk/wiki/Installation). `v0.14.0` is the first release
installable with `go install` (the Go module path moved to `github.com/axnic/rtunk`) and the first
signed one: its archives carry a cosign signature, an SBOM and SLSA build provenance, verifiable
with the steps in [SECURITY.md](SECURITY.md#verifying-a-release).

`rtunk check` exits `1` when it finds issues and `0` when the tree is clean, so it works as a CI
gate. It downloads each linter on first use, so that run takes longer.

> [!NOTE]
> Already use trunk? Run `rtunk check` directly: rtunk reads your existing `.trunk/trunk.yaml`
> (`rtunk init` migrates `.trunk/` into `.rtunk/` when you want to switch). It is not a drop-in
> clone: plain `check` never runs formatters, Windows is not supported, and actions cannot trigger
> on file changes or a schedule. See the [migration guide](https://github.com/axnic/rtunk/wiki/Migrating-From-Trunk).

## Use it in GitHub Actions

```yaml
permissions:
  contents: read
steps:
  - uses: actions/checkout@<sha>
    with:
      fetch-depth: 0
  - uses: axnic/rtunk@v0.14.0
    with:
      version: v0.14.0
      check-mode: changed-since-base
      require-attestation: true
```

The action installs and verifies rtunk, caches its tools, runs `rtunk check` and reports findings
as annotations and a job summary. Inputs, caching, verification and examples:
[GitHub Action](https://github.com/axnic/rtunk/wiki/GitHub-Action).

## Why rtunk

trunk's orchestration model is good, but `trunk` ships as a closed-source binary, which is hard to
audit or get approved in a professional or regulated environment. rtunk keeps the model and drops
the opacity: no telemetry, no cloud account, no daemon, no self-upgrade.

## Features

| Feature                       | What it does                                                                                                                        |
| ----------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| Linters, formatters, scanners | Runs the tools defined by the [trunk plugins](https://github.com/trunk-io/plugins) ecosystem: `rtunk check`, `rtunk fmt`            |
| trunk-compatible config       | Reads `.trunk/trunk.yaml` as is, or `.rtunk/rtunk.yaml`                                                                             |
| Per-project tool versions     | Linters are pinned in the config and downloaded on demand                                                                           |
| Git-aware                     | With no paths, `check` only looks at [changed files](https://github.com/axnic/rtunk/wiki/Checking-Code#choose-which-files-to-check) |
| Normalized output             | One report format for every tool: `--format human`, `json` or `sarif`                                                               |
| Autofix                       | `rtunk check --fix` applies linter fixes, then reports what remains                                                                 |
| Actions and git hooks         | `rtunk actions`, `rtunk git-hooks sync`                                                                                             |
| Renovate pins                 | `rtunk renovate enable` annotates version pins for Renovate                                                                         |
| GitHub Action                 | `axnic/rtunk` installs, verifies and caches rtunk, then reports findings as annotations                                             |
| Caching and logs              | Downloads are cached; `rtunk logs` lists and shows past runs                                                                        |

## Documentation

Full documentation is on the [wiki](https://github.com/axnic/rtunk/wiki) (sources in
[docs/](docs/)).

- [Quickstart](https://github.com/axnic/rtunk/wiki/Quickstart): from an empty repository to a report
- [Checking code](https://github.com/axnic/rtunk/wiki/Checking-Code): select files, apply fixes, exit codes
- [Configuration Reference](https://github.com/axnic/rtunk/wiki/Configuration-Reference): schema and overrides
- [Command Reference](https://github.com/axnic/rtunk/wiki/Command-Reference): every command and flag
- [Installation](https://github.com/axnic/rtunk/wiki/Installation): archive, mise, aqua, `go install`, dev container
- [GitHub Action](https://github.com/axnic/rtunk/wiki/GitHub-Action): run rtunk in CI
- [Migrating from trunk](https://github.com/axnic/rtunk/wiki/Migrating-From-Trunk)
- [FAQ](https://github.com/axnic/rtunk/wiki/FAQ)

## How this was built

> [!IMPORTANT]
> **This project is 100% vibe-coded.** Every line of code, commit, and doc was written by an AI
> coding agent (Claude Code); the author's contribution is direction, architecture review, and code
> review, not code. It serves two goals: a working, auditable alternative to `trunk`, and a
> personal experiment on whether vibecoding is practical in 2026 (setup cost, remaining friction,
> how the result holds up against 2025-era vibe-coded projects). The author overrode the agent on
> several architectural and technical decisions and reworked a handful of simple features by hand.
> Findings from that experiment are a personal opinion and are deliberately not published here.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the dev environment, tests and commit conventions, and
[SECURITY.md](SECURITY.md) to report a vulnerability. Release history is on the
[GitHub Releases](https://github.com/axnic/rtunk/releases) page.

## License

[MIT](LICENSE), copyright Alexandre NICOLAIE.
