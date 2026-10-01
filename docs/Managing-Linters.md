# Managing Linters

`rtunk linters` lists the linters available to your repository and adds them to or removes them
from the configuration; `rtunk config print` shows what the configuration resolves to.

```bash
rtunk linters enable shellcheck
```

## List linters

`rtunk linters list` groups linters as enabled, available for this repository, and (with `--all`)
the rest. A linter counts as available when its `suggest_if` says so: by default (`files_present`, or
unset) when it matches at least one file in the repository, with `config_present` when one of its
`direct_configs` exists, never with `never`. Groups are separated by a blank line.

```console
$ rtunk linters list
Available for this repo (not enabled)
  ◯ checkov         1 file
  ◯ git-diff-check  4 files
  ◯ markdownlint    1 markdown file
  ◯ prettier        2 files
  ◯ shellcheck      1 shell file
  ◯ shfmt           1 shell file
  ◯ trufflehog      4 files
  ◯ yamllint        1 yaml file
(121 other linters — rtunk linters list --all)

Enable one with: rtunk linters enable <id>
```

Pass `--all` to show the remaining linters as a third group, `Other` (linters matching no file, and
linters matching files that `suggest_if` does not suggest), or `--format json` for machine-readable
output (`enabled`, `available` and `other` arrays with `id`, `version`, `files` and `description`).

On a terminal, group headers are bold, enabled marks are green and linters matching no file are
dimmed; piped output, or any output with `NO_COLOR` set, is plain text. The collapsed hint reads
`(1 other linter — ...)` for a single linter.

## Enable linters

Pass one or more ids. Append `@<version>` to pin a version.

```console
$ rtunk linters enable shellcheck shfmt@3.7.0
$ rtunk linters list
Enabled
  ✔ shellcheck      1 shell file
  ✔ shfmt@3.7.0     1 shell file
Available for this repo (not enabled)
  ◯ checkov         1 file
  ...
```

The command edits `lint.enabled` in `.rtunk/rtunk.yaml`:

```yaml
lint:
  enabled:
    - shellcheck
    - shfmt@3.7.0
```

Either way, the command then prints what changed and how to fetch the tools:

```console
$ rtunk linters enable shellcheck shfmt@3.7.0
Enabled: shellcheck, shfmt

Download them now with:
  rtunk download lint shellcheck shfmt
```

`Disabled: <ids>` is printed when the picker unchecks enabled linters; the download hint appears only
if something was enabled. Downloading is optional: `check`, `fmt` and `run` download lazily.

Without ids, `rtunk linters enable` opens a full-screen interactive picker (alternate screen, the
terminal is restored on exit) over the same groups as `linters list --all`, separated by blank
lines. The list scrolls to the terminal height and follows resizes; `↓ N more` counts the items
below the viewport. Typing filters on the id (backspace edits; a `filter:` line shows
`(visible/total)`), up/down move, space toggles, enter confirms, esc or ctrl-c cancels. `q` does not
cancel: it is typed into the filter. Linters matching no file are dimmed. The picker needs a
terminal; without one the command stops with `interactive mode requires a terminal; pass explicit
id(s) instead`.

> [!NOTE]
> `enable` rejects ids no plugin defines (the part before any `@version`): `rtunk linters enable nope`
> exits non-zero with `unknown linter id(s): nope`, and the configuration is left unchanged. Run
> `rtunk linters list --all` to see the available ids.

## Disable linters

```bash
rtunk linters disable shfmt
```

This removes the linter from the configuration. Disabled linters no longer run in
[`rtunk check`](Checking-Code.md) or [`rtunk fmt`](Formatting-Code.md).

## Inspect the resolved configuration

Two commands print YAML (or JSON with `--output json`):

| Command               | Prints                                                                              |
| --------------------- | ----------------------------------------------------------------------------------- |
| `rtunk config print`  | The fully resolved configuration: your `rtunk.yaml` merged with every plugin source |
| `rtunk plugins print` | All configuration available across all plugins, resolved                            |

```console
$ rtunk config print
version: "0.1"
cli:
    version: ""
plugins:
    sources:
        trunk:
            id: trunk
            uri: https://github.com/trunk-io/plugins
            ref: v1.11.0
...
lint:
    enabled:
        - shellcheck
```

The output is long (the `trunk` plugin source alone defines about 130 linters); pipe it to a pager
or `grep`. Use it to see which commands, files and tools a linter will actually use.

## Where to go next

- [Checking code](Checking-Code.md) — run the linters you enabled
- [Command Reference](Command-Reference.md#linters-and-actions) — every flag of the `linters`
  commands
- [Configuration Reference](Configuration-Reference.md) — the `lint` keys in `rtunk.yaml`
