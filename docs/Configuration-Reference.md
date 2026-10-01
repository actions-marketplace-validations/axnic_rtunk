# Configuration Reference

Every key `rtunk` reads from its config file (native and trunk-compatible), the config file
discovery rules, and every override with its precedence. Verified against the key definitions in
`pkg/trunk/config`, not transcribed from prose.

| Section | Purpose |
| --- | --- |
| [Config file discovery](#config-file-discovery) | Which file `rtunk` reads, and in what order. |
| [Schema overview](#schema-overview) | Every top-level key, with type and default. |
| [`plugins.sources`](#pluginssources) | Where linter, tool, runtime and action definitions come from. |
| [`actions.disabled`](#actionsdisabled) | A record of turned-off actions; suppresses nothing by itself. |
| [Examples](#minimal-example) | The `rtunk init` scaffold and a fully populated file. |
| [Override precedence](#override-precedence) | Config path, cache directory and environment variables. |
| [Inspecting the resolved configuration](#inspecting-the-resolved-configuration) | `config print` and `plugins print`. |

This page covers the config file itself: the keys a repository writes into `.rtunk/rtunk.yaml` or
`.trunk/trunk.yaml`. For what a plugin repository contributes (linters, tools, runtimes, actions and
how they reference each other), see [Plugin Model](Plugin-Model.md). For command and flag syntax,
see [Command Reference](Command-Reference.md).

## Config file discovery

`rtunk` looks for a config file starting at the current directory and walking upward to the git
repository root (or the filesystem root outside a git repository). At each directory it checks, in
order, `.rtunk/rtunk.yaml` then `.trunk/trunk.yaml`; the first match stops the search. `--config
<path>` bypasses discovery entirely and reads that file instead. There is no environment variable
for the config path.

When both `.rtunk/rtunk.yaml` and `.trunk/trunk.yaml` exist in the same directory, `.rtunk` wins and
`.trunk` is not read at all: the two are never merged. `rtunk init` in a repository that has a
`.trunk/trunk.yaml` avoids this by migrating it to `.rtunk/` and removing `.trunk/` (see
[`rtunk init`](Command-Reference.md#rtunk-init)). Linters' config files (`direct_configs`) are
looked up in `.rtunk/configs` first, then `.trunk/configs`, so a repository still on trunk keeps
working.

> [!NOTE]
> `.rtunk/user.yaml` (a git-ignored local override) is a design goal, not shipped behavior.
> Discovery only ever checks the two paths above, and nothing reads or merges a `user.yaml`.

## Schema overview

A config file has six top-level sections. A missing section decodes to its zero value, but a
`version` line is what every real-world config and `rtunk init`'s own scaffold write in practice.

| Key | Type | Default | Meaning |
| --- | --- | --- | --- |
| `version` | string | `""` | Schema version marker. Parsed and stored, not otherwise validated or acted on. |
| `cli.version` | string | `""` | Trunk CLI version marker, kept for trunk compatibility. Parsed and stored, not otherwise validated or acted on. |
| `plugins.sources` | list of source | `[]` | Plugin repositories to merge linter, tool, runtime and action definitions from. |
| `runtimes.enabled` | list of string | `[]` | Runtime ids (optionally `id@version`) to activate. |
| `lint.enabled` | list of string | `[]` | Linter ids (optionally `id@version`) to activate. |
| `actions.enabled` | list of string | `[]` | Action ids (optionally `id@version`) to activate. |
| `actions.disabled` | list of string | `[]` | Action ids `rtunk actions disable` records as turned off. See [below](#actionsdisabled). |

An `enabled:` entry is either a bare id (`golangci-lint2`) or `id@version` (`golangci-lint2@2.13.2`)
to pin which version of that id's downloads get resolved; the `@version` suffix is stripped before
checking that the id exists. Version pinning does not choose among a linter's several `commands:`
variants; nothing in the config file does that.

`rtunk linters enable` and `rtunk actions enable` edit these lists for you; see [Managing
Linters](Managing-Linters.md) and [Actions and Git Hooks](Actions-And-Git-Hooks.md).

### `plugins.sources`

Each entry is either a **git source** (`id`, `uri`, `ref`) or a **local source** (`id`, `local`),
mutually exclusive. An entry with neither `local:` nor `uri:` set fails to resolve with `plugin
source "<id>": neither local nor uri is set`.

| Field | Applies to | Required | Meaning |
| --- | --- | --- | --- |
| `id` | both | yes | Source identifier, referenced nowhere else in the config file but used to key caching and error messages. |
| `uri` | git source | yes | Git remote URL, cloned via the system `git` binary (`init`/`fetch --depth 1`/`checkout`). |
| `ref` | git source | yes in practice | Tag or commit SHA to fetch, passed directly to `git fetch origin <ref>`; an empty value fails the fetch. Never a branch. |
| `local` | local source | yes | Filesystem path to the plugin repository, resolved relative to the config file's own directory. Must already exist: unlike a git source, a missing local path is a hard error (`plugin source "<id>": local path <path> does not exist`), never fetched. |

A git source's parsed definitions (not its raw checkout) are cached under the resolved cache
directory, keyed by `uri`+`ref`, so a pinned ref is fetched from the network only once. See [Cache
directory](#cache-directory) for where that cache lives and [Plugin Sources](Plugin-Sources.md) for
how sources are resolved.

### `actions.disabled`

`actions.disabled` is parsed and stored, but nothing reads it back: it has no effect on which
actions run. The actual gate is `actions.enabled`, which trims the action catalog down to exactly
what it lists. `rtunk actions disable <id>` removes `<id>` from `enabled:` (which is what turns it
off) and separately adds it to `disabled:` for record-keeping; `rtunk actions enable` does the
reverse. Treat `disabled:` as a log of what was turned off, not a key that itself suppresses
anything.

### Minimal example

The exact scaffold `rtunk init` writes to a fresh `.rtunk/rtunk.yaml`:

```yaml
version: "0.1"
plugins:
  sources:
    - id: trunk
      uri: https://github.com/trunk-io/plugins
      ref: v1.11.0
```

Enabled lists are intentionally omitted from the scaffold (an absent `enabled:` decodes to the same
empty list as `enabled: []`) and grown from there with `rtunk linters enable` and `rtunk actions
enable`.

### Fuller example

This repository's own `.trunk/trunk.yaml` (paraphrased, same shape) shows every section populated,
including `id@version` pinning and `actions.disabled`:

```yaml
version: "0.1"
cli:
  version: 1.25.0
plugins:
  sources:
    - id: trunk
      ref: v1.11.0
      uri: https://github.com/trunk-io/plugins
runtimes:
  enabled:
    - go@1.27.0
    - node@22.16.0
lint:
  enabled:
    - gofmt@1.20.4
    - golangci-lint2@2.13.2
    - markdownlint@0.49.1
    - prettier@3.9.6
actions:
  disabled:
    - trunk-announce
  enabled:
    - commitlint
    - trunk-check-pre-push
```

## Override precedence

### Config file path

1. `--config <path>`: read exactly that file, no discovery.
2. Automatic discovery: nearest `.rtunk/rtunk.yaml`, else nearest `.trunk/trunk.yaml`, walking from
   the current directory up to the git root (see [Config file discovery](#config-file-discovery)).

There is no environment variable for the config path.

### Cache directory

1. `--cache-dir <path>`: explicit flag value.
2. `RTUNK_CACHE_DIR` environment variable: the same underlying option as `--cache-dir`, so the flag
   wins only because an explicit `--cache-dir` overrides the value taken from the environment.
3. OS-default cache directory: `os.UserCacheDir()/rtunk/downloads` for downloads and
   `os.UserCacheDir()/rtunk/plugins` for plugin sources; the XDG cache dir on Linux,
   `~/Library/Caches` on macOS.

> [!NOTE]
> There is no `cache.dir` config-file key. The cache directory is sourced only from the flag, the
> environment variable or the OS default, never from the resolved config. A `cache.dir` field is a
> design goal, not the current schema.

When `--cache-dir` or `RTUNK_CACHE_DIR` is set to `<dir>`, downloads live under `<dir>/downloads`
and the plugin source cache under `<dir>/plugins`: the same two-subdirectory split as the OS
default, rooted differently. See [Cache and Logs](Cache-And-Logs.md) and [Cache
Architecture](Cache-Architecture.md).

### Environment

Beyond `RTUNK_CACHE_DIR`, `rtunk` reads a handful of environment variables that affect output
rendering rather than configuration resolution:

| Variable | Effect |
| --- | --- |
| `RTUNK_LIVE_HEIGHT` | Maximum height of the live view, in lines. Same option as `--live-height`; see [`rtunk check`](Command-Reference.md#rtunk-check). |
| `NO_COLOR` | Any value disables ANSI color in output. See `--ascii` and `--no-progress` in [`rtunk check`](Command-Reference.md#rtunk-check) for related output-shaping flags. |
| `TERM` | `TERM=dumb` disables the live terminal view; output falls back to plain progress lines. |
| `LC_ALL`, `LC_CTYPE`, `LANG` | Checked in that order; the first non-empty one that doesn't contain `utf-8` or `utf8` triggers the ASCII glyph fallback in the live view, the same effect as passing `--ascii`. |

None of these has a config-file equivalent: they are read directly from the process environment, not
from `.rtunk/rtunk.yaml` or `.trunk/trunk.yaml`. See [Terminal UX Design](Terminal-UX-Design.md) for
how the live view uses them.

## Ignoring issues

Inline `rtunk-ignore` and `trunk-ignore` comments are documented in [Ignoring
Issues](Ignoring-Issues.md), not in the config file.

## Inspecting the resolved configuration

`rtunk config print` prints the fully resolved, enabled configuration; `rtunk plugins print` prints
every definition available across all configured plugin sources, enabled or not. Neither command
runs the `check`/`fmt` deprecation validation, so a broken configuration can still be inspected in
order to fix it. See [`rtunk config print`](Command-Reference.md#rtunk-config-print) and [`rtunk
plugins print`](Command-Reference.md#rtunk-plugins-print) for their flags.

## Where to go next

- [Command Reference](Command-Reference.md) — the commands that read and edit these keys
- [Managing Linters](Managing-Linters.md) — enabling, pinning and disabling linters in practice
- [Ignoring Issues](Ignoring-Issues.md) — suppressing individual findings with inline comments
- [Plugin Model](Plugin-Model.md) — what a plugin source contributes
