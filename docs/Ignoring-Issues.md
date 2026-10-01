# Ignoring Issues

Add an `rtunk-ignore` comment in the file to silence a finding you accept. The comment names the
linter, optionally a rule, and a reason.

```sh
# rtunk-ignore(shellcheck/SC2086): argument is a single word by contract
echo $2
```

rtunk drops the matching findings from the report and counts them (`3 issues ... · 4 suppressed`
in the summary, `suppressed` in `--format json`). Every `trunk-ignore` comment you already have
works identically, so migrating from trunk needs no edits.

## Forms

| Form                                                                        | Suppresses                                                                     |
| --------------------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| `rtunk-ignore(<targets>): <reason>`                                         | The comment's own line, or the next line when the comment is alone on its line |
| `rtunk-ignore-all(<targets>): <reason>`                                     | The whole file, including findings that have no line number                    |
| `rtunk-ignore-begin(<targets>): <reason>` ... `rtunk-ignore-end(<targets>)` | Every line from the `-begin` line through the `-end` line, inclusive           |

The `: <reason>` part is for readers; rtunk does not read it. Write one anyway.

## Name the targets

`<targets>` is a comma-separated list. Each entry is a linter id or `linter/rule`:

| Targets                            | Suppresses                                                                         |
| ---------------------------------- | ---------------------------------------------------------------------------------- |
| `shellcheck`                       | Every rule of shellcheck                                                           |
| `shellcheck/SC2086`                | Only rule `SC2086`                                                                 |
| `eslint/no-console,no-unused-vars` | Two rules of eslint: a bare entry after a `linter/rule` entry stays on that linter |
| `eslint,prettier`                  | Both linters entirely                                                              |

Only the first `/` of an entry separates linter from rule, so rule ids that contain a slash work:
`eslint/@typescript-eslint/no-unused-vars`, `markdownlint/MD013/line-length`.

## One line

A comment that shares its line with code applies to that line. A comment alone on its line (only
whitespace and the comment opener before it) applies to the next line.

```sh
#!/bin/bash
echo $1
# rtunk-ignore(shellcheck/SC2086): argument is a single word by contract
echo $2
echo $3 # rtunk-ignore(shellcheck/SC2086): trailing form
echo $4
```

Running `rtunk check` on this file reports only lines 2 and 6 (`echo $1` and `echo $4`); lines 4
and 5 are suppressed.

## A whole file

`rtunk-ignore-all` suppresses the linter for the file, wherever the comment sits. It is the only
form that can silence findings without a line number, such as a linter that only reports
pass or fail for a file.

```sh
#!/bin/bash
# rtunk-ignore-all(shellcheck): legacy script
echo $1
rm $f
```

```console
$ rtunk check all.sh
✔ shellcheck       done     clean
Checked 1 file with 1 linter in 0.1s
✔ no issues · 3 suppressed
```

In a Markdown file, use an HTML comment:

```html
<!-- rtunk-ignore-all(markdownlint/MD013): long reference table, wrapping hurts readability -->
```

## A range of lines

`-begin` opens a range for each of its targets and the next `-end` naming the same target closes
it.

```sh
# rtunk-ignore-begin(shellcheck/SC2086): legacy block
echo $5
echo $6
# rtunk-ignore-end(shellcheck/SC2086)
echo $7
```

The `-end` must name the same targets as the `-begin`. Ranges for different linters or rules do
not interfere, and nested ranges close innermost first.

> [!WARNING]
> A `-begin` without a matching `-end` suppresses nothing, and a stray `-end` is ignored. rtunk
> prefers showing a finding you meant to hide over hiding everything after a typo. If findings
> reappear, check that both ends of the range are present and spelled identically.

## Comment markers

A directive only counts when a comment opener appears before it on the line. This stops a string
literal that looks like a directive from silencing anything:

```sh
echo "rtunk-ignore(shellcheck): not a comment, no leader"
echo $4
```

Here `echo $4` is still reported. The accepted openers (`#`, `//`, `<!--` and so on) come from the
comment formats declared by the plugin definitions of your enabled linters; you cannot set them in
your own configuration. The check uses all openers of all enabled linters, not only those of the
file's language.

## Where to go next

- [Checking Code](Checking-Code.md) — run checks and read the report
- [Migrating From Trunk](Migrating-From-Trunk.md) — what carries over from trunk
- [Configuration Reference](Configuration-Reference.md) — enable and configure linters
