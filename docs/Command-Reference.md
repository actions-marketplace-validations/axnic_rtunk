# Command Reference

Every command and flag the `rtunk` binary ships, with its arguments and defaults. Verified against
`rtunk help --all` and each `rtunk <command> --help` of the current build. For the reasoning behind
the CLI's shape and its run semantics, see [CLI Design](CLI-Design.md).

| Command                                             | Purpose                                                               |
| --------------------------------------------------- | --------------------------------------------------------------------- |
| [`rtunk check`](#rtunk-check)                       | Run enabled checks against source files (read-only).                  |
| [`rtunk fmt`](#rtunk-fmt)                           | Run configured formatters against source files.                       |
| [`rtunk run`](#rtunk-run)                           | Run an action (shortcut for `actions run`).                           |
| [`rtunk download`](#rtunk-download)                 | Download enabled tools and runtimes ahead of time.                    |
| [`rtunk config print`](#rtunk-config-print)         | Print the fully resolved configuration.                               |
| [`rtunk plugins print`](#rtunk-plugins-print)       | Print everything available across all plugins.                        |
| [`rtunk linters list`](#rtunk-linters-list)         | List linters available for the current configuration.                 |
| [`rtunk linters enable`](#rtunk-linters-enable)     | Enable linters (interactive picker when no id is given).              |
| [`rtunk linters disable`](#rtunk-linters-disable)   | Disable linters.                                                      |
| [`rtunk actions list`](#rtunk-actions-list)         | List actions available for the current configuration.                 |
| [`rtunk actions enable`](#rtunk-actions-enable)     | Enable actions (interactive picker when no id is given).              |
| [`rtunk actions disable`](#rtunk-actions-disable)   | Disable actions.                                                      |
| [`rtunk actions run`](#rtunk-actions-run)           | Run an action on demand, or every action a git hook triggers.         |
| [`rtunk actions history`](#rtunk-actions-history)   | Show recent action runs.                                              |
| [`rtunk git-hooks sync`](#rtunk-git-hooks-sync)     | Install git hooks for enabled actions.                                |
| [`rtunk git-hooks unsync`](#rtunk-git-hooks-unsync) | Remove rtunk-installed git hooks.                                     |
| [`rtunk init`](#rtunk-init)                         | Initialize rtunk in this repository.                                  |
| [`rtunk deinit`](#rtunk-deinit)                     | Remove rtunk's configuration and installed artifacts.                 |
| [`rtunk cache clean`](#rtunk-cache-clean)           | Remove the rtunk cache subtrees (downloads, plugins, logs, registry). |
| [`rtunk cache prune`](#rtunk-cache-prune)           | Remove cache entries no repository currently needs.                   |
| [`rtunk logs list`](#rtunk-logs-list)               | List this repository's recent runs.                                   |
| [`rtunk logs show`](#rtunk-logs-show)               | Show one run's log.                                                   |
| [`rtunk logs clean`](#rtunk-logs-clean)             | Delete run logs.                                                      |
| [`rtunk renovate enable`](#rtunk-renovate-enable)   | Annotate `trunk.yaml`'s version pins for Renovate.                    |
| [`rtunk renovate disable`](#rtunk-renovate-disable) | Remove the Renovate annotations.                                      |
| [`rtunk renovate config`](#rtunk-renovate-config)   | Print the Renovate `regexManagers` config to add.                     |
| [`rtunk toolbox download`](#rtunk-download)         | Alias of `rtunk download`.                                            |
| [`rtunk toolbox where`](#rtunk-toolbox-where)       | Print a cached item's install directory.                              |
| [`rtunk toolbox exec`](#rtunk-toolbox-exec-alias-x) | Run a command from a runtime or tool.                                 |
| [`rtunk toolbox link`](#rtunk-toolbox-link)         | Rebuild the `.rtunk/{logs,tools,plugins}` symlinks.                   |

Every command also accepts `-h`/`--help` for this same information at the terminal. Commands that
are groups (`check`, `linters`, `actions`, `config`, `plugins`, `cache`, `git-hooks`, `logs`,
`renovate`, `toolbox`) show only their subcommand list: the check flags, for instance, are under
`rtunk check run --help`.

## Global flags

These apply to every command below; they are not repeated in each command's own table.

| Flag              | Type   | Default                                            | Description                                                                             |
| ----------------- | ------ | -------------------------------------------------- | --------------------------------------------------------------------------------------- |
| `-h`, `--help`    | flag   | —                                                  | Show context-sensitive help.                                                            |
| `--config`        | string | nearest `.rtunk/rtunk.yaml` or `.trunk/trunk.yaml` | Path to trunk.yaml.                                                                     |
| `--cache-dir`     | string | OS cache dir (`$RTUNK_CACHE_DIR`)                  | Plugin cache directory.                                                                 |
| `--version`       | flag   | —                                                  | Print rtunk's own version and exit.                                                     |
| `--ci`            | flag   | —                                                  | Accepted for trunk compatibility; rtunk is always CI-safe, this has no effect.          |
| `-v`, `--verbose` | flag   | —                                                  | Accepted for trunk compatibility; rtunk already prints this detail, this has no effect. |

Config-path and cache-directory precedence is detailed in [Configuration
Reference](Configuration-Reference.md#override-precedence). A bare `rtunk` prints the help.

## Everyday commands

### `rtunk check`

Run enabled checks against source files (read-only). `rtunk check [<path>...]` is shorthand for
`rtunk check run [<path>...]`; the flags below belong to `check run`. Workflow and exit codes:
[Checking Code](Checking-Code.md).

```text
rtunk check [<path>...] [flags]
```

Arguments:

- `<path>...`: paths to check (default: changed files, see `--from`).

| Flag                    | Type   | Default                                                    | Description                                                                                                     |
| ----------------------- | ------ | ---------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| `--no-progress`         | flag   | —                                                          | Do not print the per-linter progress lines on stderr.                                                           |
| `--ascii`               | flag   | —                                                          | Use ASCII glyphs in the live view.                                                                              |
| `--live-height`         | int    | half the terminal height, minimum 3 (`$RTUNK_LIVE_HEIGHT`) | Maximum height of the live view in lines.                                                                       |
| `--format`              | string | `human`                                                    | Output format: `human`, `sarif` (for CI) or `json`.                                                             |
| `--from`                | string | —                                                          | Diff base for the default file selection (e.g. `origin/main`, for CI).                                          |
| `-j`, `--jobs`          | int    | number of CPUs                                             | Number of parallel linter workers.                                                                              |
| `--format-before-check` | flag   | —                                                          | Run every formatter, then check the reformatted files.                                                          |
| `-y`, `--fix`           | flag   | —                                                          | Apply linter fixes (fix commands and finding-level autofixes) to what checking found, then report what remains. |
| `--verify-stable`       | flag   | —                                                          | With `--format-before-check`, verify the formatting result is stable instead of a single pass.                  |
| `--filter`              | string | —                                                          | Comma-separated linter id allow-list, or `--filter=-id,-id...` deny-list (trunk compatibility).                 |
| `--exclude`             | string | —                                                          | Comma-separated linter id deny-list; shorthand for an inverse `--filter` (trunk compatibility).                 |
| `--security-only`       | flag   | —                                                          | Run only commands tagged `is_security: true`, skipping every other check.                                       |
| `-n`, `--no-fix`        | flag   | —                                                          | Accepted for trunk compatibility; has no effect. `--fix` always wins if both are given.                         |
| `--print-failures`      | flag   | —                                                          | Accepted for trunk compatibility; check already always prints failures, this has no effect.                     |

```bash
rtunk check --from origin/main
```

### `rtunk fmt`

Run configured formatters against source files. Workflow: [Formatting Code](Formatting-Code.md).

```text
rtunk fmt [<path>...] [flags]
```

Arguments:

- `<path>...`: paths to format (default: changed files, see `--from`).

| Flag               | Type   | Default                                                    | Description                                                                                               |
| ------------------ | ------ | ---------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| `--no-progress`    | flag   | —                                                          | Do not print the per-linter progress lines on stderr.                                                     |
| `--ascii`          | flag   | —                                                          | Use ASCII glyphs in the live view.                                                                        |
| `--live-height`    | int    | half the terminal height, minimum 3 (`$RTUNK_LIVE_HEIGHT`) | Maximum height of the live view in lines.                                                                 |
| `--format`         | string | `human`                                                    | Output format: `human` or `json` (`sarif` is only supported by `check`).                                  |
| `--from`           | string | —                                                          | Diff base for the default file selection (e.g. `origin/main`, for CI).                                    |
| `--force`          | flag   | —                                                          | Also format files with both staged and unstaged changes (skipped with a warning by default).              |
| `-j`, `--jobs`     | int    | number of CPUs                                             | Number of parallel linter workers.                                                                        |
| `-n`, `--check`    | flag   | —                                                          | Report files that would be reformatted, without writing them. (alias: `--no-fix`)                         |
| `--verify-stable`  | flag   | —                                                          | Verify the result is stable (write, dry-run check, write+check again if needed) instead of a single pass. |
| `--filter`         | string | —                                                          | Comma-separated linter id allow-list, or `--filter=-id,-id...` deny-list (trunk compatibility).           |
| `--exclude`        | string | —                                                          | Comma-separated linter id deny-list; shorthand for an inverse `--filter` (trunk compatibility).           |
| `--print-failures` | flag   | —                                                          | Accepted for trunk compatibility; fmt already always prints failures, this has no effect.                 |

```bash
rtunk fmt --check
```

### `rtunk run`

Run a specified action (shortcut for `actions run`; identical flags and behavior, see [`rtunk
actions run`](#rtunk-actions-run)).

```text
rtunk run [<args>...] [flags]
```

Arguments:

- `<args>...`: `<action-id> [-- args...]` when `--hook` is not given; otherwise just the args to
  forward.

| Flag     | Type   | Default | Description                                                                            |
| -------- | ------ | ------- | -------------------------------------------------------------------------------------- |
| `--hook` | string | —       | Run every enabled action triggered by this git hook, instead of a single action by id. |

```bash
rtunk run <action-id>
```

### `rtunk download`

Download enabled tools and runtimes into the local cache now. `check`, `fmt` and `run` otherwise
download lazily. Items already installed are skipped; runtimes a tool needs are installed too.

```text
rtunk download [<category> [<id>...]]
```

Arguments:

- `<category>`: one of `runtime`, `tools` or `lint` (a linter id, expanded to the linter's tools).
  Omit to download every enabled tool and runtime in use. With a category, ids resolve against the
  full catalog, so a just-enabled linter matching no file can be downloaded; an unknown id fails
  with `download <category>: unknown id "<id>"`.
- `<id>...`: resource id(s), each optionally `@version`. At least one is required with a category;
  a category alone is an error.

Progress: on a terminal (stderr is a TTY, `TERM` not `dumb`) the same live install view as `check`
and `fmt` (header `Installing N% <bar> done/total · elapsed`, one row per item downloading). Off a
terminal, one `installing <category>/<id>` line per item on stderr. Then on stdout, `Downloaded N
item(s).`, or `Everything is already downloaded.` when nothing was missing. Each failure prints
`✖ <category>/<id>: <err>` on stderr and the command exits non-zero with `N of M download(s)
failed`.

`rtunk toolbox download` is the same command.

Only the global flags apply.

```bash
rtunk download
rtunk download lint shellcheck shfmt
rtunk download tools shellcheck@0.10.0
```

## Configuration inspection

The resolved configuration and the plugin catalog behind it; see [Managing
Linters](Managing-Linters.md) and [Configuration Reference](Configuration-Reference.md).

### `rtunk config print`

Print the fully resolved configuration.

```text
rtunk config print [flags]
```

| Flag       | Type   | Default | Description                      |
| ---------- | ------ | ------- | -------------------------------- |
| `--output` | string | `yaml`  | Output format: `yaml` or `json`. |

```bash
rtunk config print --output json
```

### `rtunk plugins print`

Print all configuration available across all plugins, resolved. Can be very large; useful as a
registry dump. Replaces `config print --all`, which is removed.

```text
rtunk plugins print [flags]
```

| Flag       | Type   | Default | Description                      |
| ---------- | ------ | ------- | -------------------------------- |
| `--output` | string | `yaml`  | Output format: `yaml` or `json`. |

```bash
rtunk plugins print
```

## Linters and actions

Workflows: [Managing Linters](Managing-Linters.md) and [Actions and Git
Hooks](Actions-And-Git-Hooks.md).

### `rtunk linters list`

List all linters available for the current configuration.

```text
rtunk linters list [flags]
```

| Flag       | Type   | Default | Description                                                                              |
| ---------- | ------ | ------- | ---------------------------------------------------------------------------------------- |
| `--all`    | flag   | —       | Also show the `Other` group: linters matching no file, or not suggested by `suggest_if`. |
| `--format` | string | `human` | Output format: `human` or `json`.                                                        |

A linter enabled or re-pinned by a [local override
file](Configuration-Reference.md#local-override-files) shows ` (from <file>)` after its file count
(`✔ vet  2 go files (from user.yaml)`); one kept off by `lint.disabled` shows ` (disabled by <file>)`
(`◯ mdlint  1 markdown file (disabled by rtunk.local.yaml)`). In `--format json`, items gain
`enabled_by` and `disabled_by` (file name, omitted when empty).

```bash
rtunk linters list --all
```

### `rtunk linters enable`

Enable one or more linters. With no id, opens a full-screen interactive checklist over the full
catalog (the items `linters list --all` shows, in the same groups) and applies the selection.

```text
rtunk linters enable [<id>...]
```

Arguments:

- `<id>...`: linter id(s) to enable, optionally `@version`. Omit for an interactive picker.

Unknown ids are rejected with an error and a non-zero exit; the configuration is left unchanged.
Only the base config file is edited: an id listed in `lint.disabled` of a [local override
file](Configuration-Reference.md#local-override-files) stays off even once enabled here, and
`warning: <id> stays disabled: <file> lists it under lint.disabled` is printed on stderr. The
download hint below is still printed for such an id.

Afterwards prints `Enabled: <ids>` and/or `Disabled: <ids>` and, if anything was enabled, the hint
`rtunk download lint <ids>`.

Picker keys: type to filter on id (backspace edits), up/down move, space toggles, enter confirms,
esc or ctrl-c cancels (`q` is typed into the filter).

Only the global flags apply.

```bash
rtunk linters enable shellcheck@0.10.0
```

### `rtunk linters disable`

Disable one or more linters.

```text
rtunk linters disable <id>...
```

Arguments:

- `<id>...`: linter id(s) to disable.

Only the base config file is edited. If a [local override
file](Configuration-Reference.md#local-override-files) still enables the id, stderr gets `warning:
<id> stays enabled: <file> enables it (remove it there, or list it under lint.disabled)`.

Only the global flags apply.

```bash
rtunk linters disable shellcheck
```

### `rtunk actions list`

List actions available for the current configuration.

```text
rtunk actions list [flags]
```

| Flag       | Type   | Default | Description                       |
| ---------- | ------ | ------- | --------------------------------- |
| `--format` | string | `human` | Output format: `human` or `json`. |

```bash
rtunk actions list
```

### `rtunk actions enable`

Enable one or more actions. With no id, opens the same full-screen interactive picker (keys as in
`linters enable`).

```text
rtunk actions enable [<id>...]
```

Arguments:

- `<id>...`: action id(s) to enable. Omit for an interactive picker.

Unknown ids are rejected with an error and a non-zero exit; the configuration is left unchanged.

Only the global flags apply.

```bash
rtunk actions enable <action-id>
```

### `rtunk actions disable`

Disable one or more actions.

```text
rtunk actions disable <id>...
```

Arguments:

- `<id>...`: action id(s) to disable.

Only the global flags apply.

```bash
rtunk actions disable <action-id>
```

### `rtunk actions run`

Run an action on demand, or every action a git hook triggers. `rtunk run` is a shortcut for this
command.

```text
rtunk actions run [<args>...] [flags]
```

Arguments:

- `<args>...`: `<action-id> [-- args...]` when `--hook` is not given; otherwise just the args to
  forward.

| Flag     | Type   | Default | Description                                                                            |
| -------- | ------ | ------- | -------------------------------------------------------------------------------------- |
| `--hook` | string | —       | Run every enabled action triggered by this git hook, instead of a single action by id. |

```bash
rtunk actions run <action-id>
```

### `rtunk actions history`

Show recent action runs.

```text
rtunk actions history [flags]
```

| Flag      | Type   | Default | Description                                 |
| --------- | ------ | ------- | ------------------------------------------- |
| `--id`    | string | —       | Restrict to one action id.                  |
| `--limit` | int    | `20`    | Maximum entries to show. (alias: `--count`) |

```bash
rtunk actions history --id <action-id> --limit 5
```

### `rtunk git-hooks sync`

Install git hooks for enabled actions. Idempotent: re-running it rewrites the same hook files.

```text
rtunk git-hooks sync [flags]
```

| Flag      | Type | Default | Description                                 |
| --------- | ---- | ------- | ------------------------------------------- |
| `--force` | flag | —       | Overwrite an existing, non-rtunk hook file. |

```bash
rtunk git-hooks sync
```

### `rtunk git-hooks unsync`

Remove rtunk-installed git hooks (what `sync` installed).

```text
rtunk git-hooks unsync
```

Only the global flags apply.

```bash
rtunk git-hooks unsync
```

## Setup and teardown

### `rtunk init`

Initialize rtunk in this repository. Without `.trunk/trunk.yaml`, writes the `.rtunk/rtunk.yaml`
scaffold (see [Configuration Reference](Configuration-Reference.md#minimal-example)). With one, it
migrates instead: `.trunk/trunk.yaml` moves to `.rtunk/rtunk.yaml` (content unchanged), as do
`configs/`, `user_trunk.yaml` and `user.yaml` when present, printing `migrated .trunk/<x> ->
.rtunk/<y>` for each; the rest of `.trunk/` (trunk's links into `~/.cache/trunk`, plugin checkouts,
`.gitignore`) is then removed (`removed .trunk/`). `trunk.yaml` being tracked, git keeps the
original.

Then prints `initialized rtunk at <path>`. On a terminal (stdin and stdout), it runs the linters
picker, the actions picker (as in `linters enable` / `actions enable`, with the `Enabled:` /
`Disabled:` summary but no download hint), then `rtunk download` of everything enabled with the
live install view; a download failure is only a warning, `check` and `fmt` retry later. Off a
terminal it prints `next: rtunk linters enable, rtunk actions enable, rtunk download`. Finally it
links `.rtunk/{logs,tools,plugins}` and prints `then: rtunk git-hooks sync to install the git hooks of
the enabled actions`. The generated `.rtunk/.gitignore` ignores `logs`, `tools`, `plugins`,
`user_trunk.yaml`, `user.yaml` and `rtunk.local.yaml` (the [local override
files](Configuration-Reference.md#local-override-files)).

```text
rtunk init [flags]
```

| Flag      | Type | Default | Description                                                                                                 |
| --------- | ---- | ------- | ----------------------------------------------------------------------------------------------------------- |
| `--force` | flag | —       | Overwrite an existing `.rtunk/rtunk.yaml` (required when it exists; the scaffold or migration replaces it). |

```console
$ rtunk init
initialized rtunk at <repo>/.rtunk/rtunk.yaml
next: rtunk linters enable, rtunk actions enable, rtunk download
linked <repo>/.rtunk
then: rtunk git-hooks sync to install the git hooks of the enabled actions
```

### `rtunk deinit`

Remove rtunk's configuration and installed artifacts.

```text
rtunk deinit [flags]
```

| Flag          | Type | Default | Description                                                                 |
| ------------- | ---- | ------- | --------------------------------------------------------------------------- |
| `-y`, `--yes` | flag | —       | Accepted for trunk compatibility; deinit never prompts, this has no effect. |

```bash
rtunk deinit
```

## Cache and logs

Where the cache lives and what a run log contains: [Cache and Logs](Cache-And-Logs.md).

### `rtunk cache clean`

Remove the rtunk cache: the `downloads`, `plugins`, `logs` and `registry` subtrees of the cache
root, and nothing else in it. On a terminal, one line per existing subtree with a spinner that turns
into a green `✔ <name>  <size> freed` (red `✖` with the error on failure); removals run in parallel
and the result stays on screen. Off a terminal, `removed <name> (<size>)` per subtree. When none
exists, prints `The cache is already empty.`

```text
rtunk cache clean
```

Only the global flags apply.

```bash
rtunk cache clean
```

### `rtunk cache prune`

Remove cache entries no repository currently needs.

```text
rtunk cache prune
```

Only the global flags apply.

```bash
rtunk cache prune
```

### `rtunk logs list`

List this repository's recent runs.

```text
rtunk logs list
```

Only the global flags apply.

```bash
rtunk logs list
```

### `rtunk logs show`

Show one run's log (default: the latest).

```text
rtunk logs show [<run>] [flags]
```

Arguments:

- `<run>`: run name (or a unique prefix of it) as printed by `logs list`, or `latest`.

| Flag     | Type | Default | Description                                        |
| -------- | ---- | ------- | -------------------------------------------------- |
| `--json` | flag | —       | Print the raw JSONL instead of the text rendering. |

```bash
rtunk logs show latest
```

### `rtunk logs clean`

Delete this repository's run logs.

```text
rtunk logs clean [flags]
```

| Flag    | Type | Default | Description                                          |
| ------- | ---- | ------- | ---------------------------------------------------- |
| `--all` | flag | —       | Delete every repository's logs, not just this one's. |

```bash
rtunk logs clean
```

## Renovate

Workflow: [Keeping Tools Up To Date](Keeping-Tools-Up-To-Date.md). Both `enable` and `disable` warn
on stderr when no Renovate config file at the repository root contains the regexManager `renovate
config` prints.

### `rtunk renovate enable`

Annotate `trunk.yaml`'s version pins for Renovate.

```text
rtunk renovate enable
```

Only the global flags apply.

```bash
rtunk renovate enable
```

### `rtunk renovate disable`

Remove the Renovate annotations from `trunk.yaml`.

```text
rtunk renovate disable
```

Only the global flags apply.

```bash
rtunk renovate disable
```

### `rtunk renovate config`

Print the Renovate `regexManagers` config to add.

```text
rtunk renovate config
```

Only the global flags apply.

```bash
rtunk renovate config
```

## Toolbox (internal commands)

`toolbox` is callable but absent from the default `rtunk --help`; `rtunk help --all` lists it. The
`<category>` argument of `where` and `exec` is `runtime` or `tools`. `toolbox download` is the
public [`rtunk download`](#rtunk-download).

### `rtunk toolbox where`

Print a cached item's install directory (not its shim).

```text
rtunk toolbox where <category> <id>
```

Arguments:

- `<category>`: resource category, one of `runtime` or `tools`.
- `<id>`: resource id, optionally `@version`.

Only the global flags apply.

```bash
rtunk toolbox where tools shellcheck@0.10.0
```

### `rtunk toolbox exec` (alias `x`)

Run a command from a runtime or tool, downloading it first if missing. `<cmd>` is the item's shim
when equal to `<id>`, else an executable in its install directory.

```text
rtunk toolbox exec <category> <id> <args>... [flags]
```

Arguments:

- `<category>`: resource category, one of `runtime` or `tools`.
- `<id>`: resource id, optionally `@version`.
- `<args>...`: command to run, then its arguments.

| Flag            | Type | Default | Description                            |
| --------------- | ---- | ------- | -------------------------------------- |
| `--interactive` | flag | —       | Bind stdin and stdout to the terminal. |

```bash
rtunk toolbox x runtime python@3.14 -- python3 --version
```

### `rtunk toolbox link`

(Re)build `.rtunk/logs`, `.rtunk/tools/<id>` and `.rtunk/plugins/<source-id>` as symlinks into
rtunk's cache, plus a `.rtunk/.gitignore` covering them and the local override files (`user_trunk.yaml`, `user.yaml`,
`rtunk.local.yaml`); missing entries are appended to an existing `.gitignore`, existing lines are kept. The links give shells, editors and other
tooling a fixed local path to the logs, the resolved tool shims and the plugin checkouts. A link may
dangle until something populates its target; a `local:` plugin source is not linked. See [Cache and
Logs](Cache-And-Logs.md).

```text
rtunk toolbox link
```

Only the global flags apply.

```console
$ rtunk toolbox link
linked <repo>/.rtunk
```

## Where to go next

- [Checking Code](Checking-Code.md) and [Formatting Code](Formatting-Code.md) — the everyday
  workflow behind `check` and `fmt`
- [Configuration Reference](Configuration-Reference.md) — the keys these commands read and write
- [CLI Design](CLI-Design.md) — why the CLI is shaped this way, and its run semantics
