# Formatting Code

`rtunk fmt` runs every enabled formatter on the files you changed and rewrites them in place.

```bash
rtunk fmt
```

## Format your changes

Without arguments, `rtunk fmt` selects files the same way as
[`rtunk check`](Checking-Code.md#choose-which-files-to-check): the files that differ from the
upstream merge base (or from `HEAD` without an upstream), including untracked files. Pass paths to
format specific files or directories, and `--from <ref>` to set the diff base.

```console
$ rtunk fmt
▲ taplo            done     1 file changed
REFORMATTED   1 file

  c.toml

Checked 1 file with 1 linter in 0.1s
✔ 1 file reformatted
```

With nothing to format, rtunk prints `rtunk: no files to format` and exits `0`. A `fmt` that
rewrites files also exits `0`: reformatting is not an error.

`fmt` writes to the working tree only. It never stages anything, so review the result with
`git diff` and add what you want to keep.

## Preview without writing

`--check` (`-n`) reports the files that would change and leaves them untouched. It exits `1` when
any file needs reformatting, which makes it suitable for CI:

```console
$ rtunk fmt --check
▲ taplo            done     1 file would change
WOULD REFORMAT   1 file

  c.toml

Checked 1 file with 1 linter in 0.0s
✖ 1 file would be reformatted
rtunk: rtunk: fmt --check found 1 file(s) needing reformatting
```

## Files with staged and unstaged changes

A file that is partly staged has changes in both the index and the working tree. Because `fmt`
rewrites the working copy, formatting it could mix your staged and unstaged edits. rtunk skips such
files with a warning:

```console
$ rtunk fmt
rtunk: skipping partially staged file /path/to/c.toml (use --force to format it)
rtunk: no files to format
```

Add `--force` to format them anyway. The index is still left untouched.

## Formatting and checking

`fmt` and `check` are separate on purpose: `check` reads files, `fmt` writes them.

| You want to                                      | Run                                 |
| ------------------------------------------------ | ----------------------------------- |
| Rewrite files with formatters                    | `rtunk fmt`                         |
| Know which files a formatter would change        | `rtunk fmt --check`                 |
| Format, then lint the formatted files in one run | `rtunk check --format-before-check` |

Other flags (`--format human|json`, `--filter`, `--exclude`, `--verify-stable`, `--jobs`) are
listed in [`rtunk fmt`](Command-Reference.md#rtunk-fmt). `--format sarif` is only available for
`check`; `fmt` refuses it with exit `1` before running anything.

> [!NOTE]
> If a formatter reports `warning: ... was reformatted again ... possible formatter instability`,
> run again with `--verify-stable`: rtunk then repeats the pass until the output stops changing.

## Where to go next

- [Checking Code](Checking-Code.md) — lint your changes and combine with `--format-before-check`
- [Command Reference](Command-Reference.md#rtunk-fmt) — every flag of `rtunk fmt`
- [Ignoring Issues](Ignoring-Issues.md) — suppress findings you accept
