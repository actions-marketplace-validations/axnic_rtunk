# rtunk

rtunk is a from-scratch, open-source rewrite of the [trunk.io Code Quality
CLI](https://docs.trunk.io/code-quality/overview), for developers and teams who want one command
to run their linters, formatters and security scanners.

It orchestrates existing tools behind a single declarative config, with per-project tool versions,
git-aware "only check what changed" runs, normalized output and caching. `trunk` itself is a
closed-source binary, which is hard to audit or get approved in a regulated environment; rtunk is
the same orchestration model as an auditable, 100% local tool.

> [!IMPORTANT]
> This project is 100% vibe-coded: every line of code, commit and doc was written by an AI coding
> agent (Claude Code). The author's contribution is direction, architecture review and code review.

| rtunk is                                                        | rtunk is not                                                             |
| --------------------------------------------------------------- | ------------------------------------------------------------------------ |
| Configuration-compatible with trunk (reads `.trunk/trunk.yaml`) | A drop-in clone: [deliberate differences](Migrating-From-Trunk.md) exist |
| Open source and auditable                                       | A closed binary                                                          |
| 100% local: no telemetry, no cloud account                      | A service with a daemon or a background process                          |
| Consumer of the trunk-io/plugins ecosystem                      | A reimplementation of linter definitions                                 |
| Supported on macOS and Linux                                    | Supported on Windows                                                     |

## Getting started

For anyone installing rtunk or moving a trunk repository to it.

- [Installation](Installation.md) — build and install rtunk, supported platforms, upgrading
- [Quickstart](Quickstart.md) — from an empty repository to your first `rtunk check` report
- [Migrating from trunk](Migrating-From-Trunk.md) — what carries over and what changes

## Using rtunk

For everyday use, organized by what you are trying to do.

- [Checking code](Checking-Code.md) — run linters, select files, apply fixes, read exit codes
- [Formatting code](Formatting-Code.md) — run formatters with `rtunk fmt`
- [Ignoring issues](Ignoring-Issues.md) — suppress findings with `rtunk-ignore` comments
- [Managing linters](Managing-Linters.md) — list, enable and disable linters and inspect plugins
- [Actions and git hooks](Actions-And-Git-Hooks.md) — run actions and trigger them from git hooks
- [Keeping tools up to date](Keeping-Tools-Up-To-Date.md) — maintain version pins with Renovate
- [Cache and logs](Cache-And-Logs.md) — manage downloaded tools and read past run logs
- [GitHub Action](GitHub-Action.md) — run rtunk in GitHub Actions with caching, annotations and a job summary

## Reference

For looking up exact flags, keys and answers.

- [Command Reference](Command-Reference.md) — every command and flag
- [Configuration Reference](Configuration-Reference.md) — config schema and override precedence
- [FAQ](FAQ.md) — short answers to common questions

## Internals

For contributors: how rtunk is designed and why. These pages describe design, not usage.

- [Architecture](Architecture.md) — package layout and responsibilities
- [Plugin Model](Plugin-Model.md) — how plugin definitions become runnable linters
- [Plugin Sources](Plugin-Sources.md) — how plugin sources are resolved
- [Cache Architecture](Cache-Architecture.md) — the content-addressed cache
- [Run Flows](Run-Flows.md) — how a run executes end to end
- [CLI Design](CLI-Design.md) — the command surface and run semantics
- [Terminal UX Design](Terminal-UX-Design.md) — output rendering and the live view
- [Decision Log](Decision-Log.md) — maintainers' recorded decisions

The project's source, license and contribution guide are in the
[repository](https://github.com/xunleii/rtunk); see also [CONTRIBUTING.md](../CONTRIBUTING.md).
