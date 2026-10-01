# Managing Linters

`rtunk linters` lists the linters available to your repository and adds them to or removes them
from the configuration; `rtunk config print` shows what the configuration resolves to.

```bash
rtunk linters enable shellcheck
```

## List linters

`rtunk linters list` groups linters as enabled, available for this repository, and (with `--all`)
the rest. A linter counts as available when it matches at least one file in the repository.

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
(121 other linters don't match any file here — rtunk linters list --all)

Enable one with: rtunk linters enable <id>
```

Pass `--all` to include the linters that match no file, or `--format json` for machine-readable
output (`enabled` and `available` arrays with `id`, `version`, `files` and `description`).

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

Without ids, `rtunk linters enable` opens an interactive picker over the available linters. The
picker needs a terminal; without one the command stops with `interactive mode requires a terminal;
pass explicit id(s) instead`.

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
