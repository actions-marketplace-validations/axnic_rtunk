# FAQ

## Does rtunk read my existing `.trunk/trunk.yaml`?

Yes. rtunk understands `.trunk/trunk.yaml` as-is, so a trunk repository can run `rtunk check`
without any rewrite. See [Migrating from trunk](Migrating-From-Trunk.md) and
[config file discovery](Configuration-Reference.md#config-file-discovery).

## What happens if both `.rtunk` and `.trunk` exist?

`.rtunk/rtunk.yaml` wins and `.trunk/trunk.yaml` is not read; the two are never merged.
`rtunk init` avoids it by migrating `.trunk/` into `.rtunk/`. See
[config file discovery](Configuration-Reference.md#config-file-discovery).

## Why did `check` not format my files?

Plain `check` only runs checking commands and never runs formatters, unlike trunk. Run `rtunk fmt`
to format, or pass `--format-before-check` to `check` to reproduce trunk's behavior. See
[Migrating from trunk](Migrating-From-Trunk.md) and [Formatting code](Formatting-Code.md).

## Why does `check` only look at some of my files?

With no paths, `check` selects the files that changed relative to git; `--from` sets the diff base
(for example `origin/main` in CI), and explicit paths override the selection. See
[Checking code](Checking-Code.md).

## Where are tools downloaded?

Into rtunk's cache, in the OS cache directory by default (`~/Library/Caches/rtunk` on macOS, the
XDG cache directory on Linux). Override it with `--cache-dir` or `RTUNK_CACHE_DIR`. See
[Cache and logs](Cache-And-Logs.md) and
[cache directory](Configuration-Reference.md#cache-directory).

## Does rtunk send telemetry?

No. rtunk makes no network call you did not trigger: only linter, runtime and tool downloads, and
cloning the plugin sources declared in config. See [Architecture](Architecture.md).

## Does rtunk update itself?

No. There is no `rtunk upgrade` command; reinstall the way you installed it. See
[Installation](Installation.md#upgrade-rtunk).

## Is Windows supported?

No. macOS and Linux are the supported host platforms, and there is no Windows build. See
[Installation](Installation.md#check-that-your-platform-is-supported).

## How do I silence one finding?

Add an `rtunk-ignore(<linter>/<rule>)` comment on the line, or `rtunk-ignore-all(...)` for a whole
file. Existing `trunk-ignore` comments keep working. See [Ignoring issues](Ignoring-Issues.md).

## Why does `rtunk check --help` not list the flags?

`check` is a command group; its flags are on `rtunk check run --help`. See the
[Command Reference](Command-Reference.md).

## Do I need a trunk account or a cloud service?

No. rtunk is 100% local: no cloud account, no daemon, no telemetry. See [Home](README.md).

## Can rtunk run as a daemon or trigger actions on file changes?

No. Actions cannot trigger on file changes or a schedule, because both need a background daemon
and rtunk refuses to run as one. See [Actions and git hooks](Actions-And-Git-Hooks.md).
