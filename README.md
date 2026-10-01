# rtunk

**One command for your linters, formatters and security scanners. Open source, 100% local.**

[![Go version](https://img.shields.io/github/go-mod/go-version/xunleii/rtunk)](go.mod)
[![License: MIT](https://img.shields.io/github/license/xunleii/rtunk)](LICENSE)
[![Docs: wiki](https://img.shields.io/badge/docs-wiki-blue)](https://github.com/axnic/rtunk/wiki)

rtunk is a from-scratch rewrite of [trunk.io's Code Quality
CLI](https://docs.trunk.io/code-quality/overview): it orchestrates existing tools behind a single
declarative config.

```bash
go install github.com/xunleii/rtunk/cmd/rtunk@latest   # needs Go on PATH; see Installation
cd your-git-repo && rtunk init && rtunk linters enable yamllint
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

`rtunk check` exits `1` when it finds issues and `0` when the tree is clean, so it works as a CI
gate. It downloads each linter on first use, so that run takes longer.

> [!NOTE]
> Already use trunk? Skip `rtunk init`: it creates `.rtunk/rtunk.yaml`, which then takes precedence
> over `.trunk/trunk.yaml`. Run `rtunk check` directly instead.

## Why rtunk

trunk's orchestration model is good, but `trunk` ships as a closed-source binary, which is hard to
audit or get approved in a professional or regulated environment. rtunk keeps the model and drops
the opacity: no telemetry, no cloud account, no daemon, no self-upgrade.

## Features

| Feature | What it does |
| ------- | ------------ |
| Linters, formatters, scanners | Runs the tools defined by the [trunk plugins](https://github.com/trunk-io/plugins) ecosystem: `rtunk check`, `rtunk fmt` |
| trunk-compatible config | Reads `.trunk/trunk.yaml` as is, or `.rtunk/rtunk.yaml` |
| Per-project tool versions | Linters are pinned in the config and downloaded on demand |
| Git-aware | With no paths, `check` only looks at changed files |
| Normalized output | One report format for every tool: `--format human`, `json` or `sarif` |
| Autofix | `rtunk check --fix` applies linter fixes, then reports what remains |
| Actions and git hooks | `rtunk actions`, `rtunk git-hooks sync` |
| Renovate pins | `rtunk renovate enable` annotates version pins for Renovate |
| Caching and logs | Downloads are cached; `rtunk logs` lists and shows past runs |

## Coming from trunk?

rtunk reads your existing `.trunk/trunk.yaml`. It is not a drop-in clone: plain `check` never runs
formatters, Windows is not supported, and actions cannot trigger on file changes or a schedule. The
[migration guide](https://github.com/axnic/rtunk/wiki/Migrating-From-Trunk) lists what carries
over, what changes and why.

## Documentation

Full documentation is on the [wiki](https://github.com/axnic/rtunk/wiki) (sources in
[docs/](docs/)).

- [Quickstart](https://github.com/axnic/rtunk/wiki/Quickstart): from an empty repository to a report
- [Checking code](https://github.com/axnic/rtunk/wiki/Checking-Code): select files, apply fixes, exit codes
- [Configuration Reference](https://github.com/axnic/rtunk/wiki/Configuration-Reference): schema and overrides
- [Command Reference](https://github.com/axnic/rtunk/wiki/Command-Reference): every command and flag
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
[SECURITY.md](SECURITY.md) to report a vulnerability. Release history is in
[CHANGELOG.md](CHANGELOG.md).

## License

[MIT](LICENSE), copyright Alexandre NICOLAIE.
