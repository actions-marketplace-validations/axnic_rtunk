# Checking Code

`rtunk check` runs every enabled linter against the files you changed and reports what it finds. It
never modifies files unless you ask with `--fix`.

```bash
rtunk check
```

## Choose which files to check

Without arguments, `rtunk check` checks the files that differ from a base commit, including staged,
unstaged and untracked files that git does not ignore. The base depends on the repository:

| Situation                      | Files checked                                           |
| ------------------------------ | ------------------------------------------------------- |
| In git, branch has an upstream | Changes since the merge base of the upstream and `HEAD` |
| In git, no upstream            | Changes since `HEAD`                                    |
| Not in git                     | Nothing: rtunk exits with an error, pass explicit paths |

Precisely, the default file set is the union of:

- the files in `git diff <base>` against the working tree, so staged and unstaged changes alike
  (added, modified and renamed; deleted files are never selected);
- untracked files that `.gitignore` does not exclude.

`<base>` is the merge base of the reference and `HEAD`: the reference is `--from` when given, else
the branch's upstream (`@{upstream}`). With neither, `<base>` is `HEAD` itself, so everything since
the last commit counts; in a repository with no commit yet, every file counts. A `--from` ref that
cannot be resolved is an error. Symlinks are dropped from the selection (their target is selected
under its own path).

`rtunk fmt` uses the same selection, and additionally skips files with both staged and unstaged
changes (see [Formatting Code](Formatting-Code.md#files-with-staged-and-unstaged-changes)).

When nothing changed, rtunk says so and exits `0`:

```console
$ rtunk check
rtunk: no files to check
```

Pass paths to check them regardless of git state. A directory means every file under it that git
tracks or could track (ignored files are skipped):

```bash
rtunk check .
rtunk check scripts/run.sh
```

Explicit paths bypass the diff: the file set is every file under them that git tracks or could
track, whether or not it changed.

In CI, a checkout is often a detached `HEAD` with no upstream, so the default base is `HEAD` and
nothing is selected on a clean checkout. Force the base with `--from`:

```bash
rtunk check --from origin/main
```

> [!NOTE]
> Outside a git repository, `rtunk check` without paths fails with
> `rtunk: outside a git repository, explicit paths are required` and exits `1`. A path that does
> not exist also exits `1`.

## Read the output

Each linter prints one progress line on stderr, then findings are grouped by file on stdout:

```console
$ rtunk check run.sh
▲ shellcheck       done     3 issues
run.sh  (3)
  2:6  low     Double quote to prevent globbing and word splitting.  shellcheck/SC2086
  3:4  medium  f is referenced but not assigned.                     shellcheck/SC2154
  3:4  low     Double quote to prevent globbing and word splitting.  shellcheck/SC2086

Checked 1 file with 1 linter in 0.5s
✖ 3 issues (0 high · 0 medium · 1 low)
rtunk: rtunk: check found 3 issue(s)
```

Each finding line reads `line:column`, severity, message, then `linter/rule`. The summary counts
issues per severity. Findings silenced by an ignore comment are not listed; the summary shows how
many were hidden (`3 issues ... · 4 suppressed`). See [Ignoring issues](Ignoring-Issues.md).

> [!NOTE]
> `rtunk check` runs linters only. A formatter whose job is to rewrite files does not report
> unformatted code here; use [`rtunk fmt --check`](Formatting-Code.md#preview-without-writing) or
> `--format-before-check` below.

Two flags change what appears on screen:

- `--no-progress` drops the per-linter progress lines and the live view, and keeps warnings and
  errors.
- `--ascii` replaces the live view's glyphs with ASCII. rtunk also does this on its own when the
  locale is not UTF-8.

Color is added only when stdout is a terminal and `NO_COLOR` is empty.

To restrict a run to some linters, use `--filter=a,b` (only these), `--filter=-a,-b` or
`--exclude=a,b` (all but these), or `--security-only` for commands tagged as security checks. Full
flag list: [`rtunk check`](Command-Reference.md#rtunk-check).

## Choose an output format

`--format` selects what rtunk writes to stdout. Progress lines stay on stderr whatever the format.

| Format   | Use it for                                                     |
| -------- | -------------------------------------------------------------- |
| `human`  | Reading in a terminal (default)                                |
| `json`   | Scripts: one JSON document with issues, failures, run metadata |
| `sarif`  | CI code-scanning tools that read SARIF 2.1.0                   |
| `github` | GitHub Actions: one annotation per finding, plus a job summary |

```console
$ rtunk check --format json run.sh 2>/dev/null
{
  "version": 1,
  "command": "check",
  "elapsed_ms": 56,
  "run_log": "20261001T074014.933051000Z-check",
  "files_checked": 1,
  "linters": 1,
  "suppressed": 0,
  "issues": [
    {
      "file": "run.sh",
      "line": 2,
      "column": 6,
      "severity": "low",
      "message": "Double quote to prevent globbing and word splitting.",
      "linter": "shellcheck",
      "rule": "SC2086",
      "url": "https://github.com/koalaman/shellcheck/wiki/SC2086"
    }
  ],
  "failures": [],
  "skipped": [],
  "changed": []
}
```

(Output trimmed to one issue.) Every key is always present, empty lists are `[]`, and `line` and
`column` are `0` when the linter does not report them. `failures` lists linters that could not run.
With no files selected, rtunk writes no document at all and exits `0`.

### GitHub Actions annotations

`--format github` writes one [workflow command](https://docs.github.com/actions/reference/workflow-commands-for-github-actions)
per finding on stdout, which GitHub turns into an annotation on the pull request:

```console
$ rtunk check --format github README.md 2>/dev/null
::error file=README.md,line=2,title=markdownlint/MD001::Heading levels should only increment by one level at a time
```

- **Level.** `high` findings (`error`) become `::error`, `medium` (`warning`) become `::warning`,
  anything else `::notice`.
- **Location.** `file` is relative to `$GITHUB_WORKSPACE` (the repository root when unset) with
  forward slashes. `line` and `col` are present only when the linter reports them, and a finding
  without a file becomes an annotation that is not attached to any file. rtunk does not record end
  positions, so `endLine` and `endColumn` are never written. `title` is `linter/rule`, or `linter`.
- **Failed and skipped linters** become `::error title=<linter>::linter failed to run: ...` and
  `::warning title=<linter>::linter skipped: ...`.
- **Escaping.** `%`, CR and LF are escaped in messages, and `:` and `,` also in properties, as
  GitHub requires.
- **No truncation.** rtunk emits every finding. GitHub shows at most 10 errors and 10 warnings per
  step, and 50 annotations per job; the rest only appear in the step log.

When `GITHUB_STEP_SUMMARY` is set (GitHub sets it in every step), rtunk also appends a Markdown job
summary to that file: totals, counts per severity and per linter, a table of the first 50 findings
with an `N more not shown` line, and the failed and skipped linters. When it is unset, nothing is
written. The format is never selected automatically: pass `--format github` explicitly. The
exit code is the same as for every other format.

```bash
rtunk check --format sarif --from origin/main > rtunk.sarif
```

## Fix problems automatically

| Flag                    | What it does                                                                           |
| ----------------------- | -------------------------------------------------------------------------------------- |
| `--fix` (`-y`)          | Applies linters' own fix commands and per-finding autofixes, then reports what remains |
| `--format-before-check` | Runs every enabled formatter first, then checks the reformatted files                  |

`--fix` never runs formatters, and `--format-before-check` never applies fixes. Combine them to get
the full sequence: format, check, apply fixes, check again.

```console
$ rtunk check --format-before-check
▲ taplo            done     1 file changed
REFORMATTED   1 file

  c.toml

Checked 1 file with 1 linter in 0.1s
✔ 1 file reformatted
✔ taplo            done     clean
...
```

Both flags rewrite files in your working tree. Commit or stash first so you can review the diff.
`--verify-stable` applies to the formatting pass of `--format-before-check`: it re-runs the
formatters until the result stops changing, to catch unstable formatters.

> [!NOTE]
> `-n`/`--no-fix` is accepted for trunk compatibility and has no effect. If you pass both, `--fix`
> wins.

## Use rtunk in CI

rtunk is always CI-safe: `--ci` is accepted for trunk compatibility and changes nothing. A
typical job:

```bash
rtunk check --from origin/main --format sarif > rtunk.sarif
```

Exit codes:

| Code | Meaning                                                                         |
| ---- | ------------------------------------------------------------------------------- |
| `0`  | No findings, or no files to check                                               |
| `1`  | Findings, a linter that failed to run, an invalid config, or a nonexistent path |

There is no separate code for "findings" versus "error". The code does not depend on `--format`.
When a linter fails to run, the report has a `FAILURES` section and the run log id; read it with
`rtunk logs show <id>`.

On GitHub Actions, the [GitHub Action](GitHub-Action.md) wraps this: install, cache, annotations.

## Where to go next

- [Command Reference](Command-Reference.md#rtunk-check) — every flag of `rtunk check`
- [Formatting Code](Formatting-Code.md) — write formatter output to your files with `rtunk fmt`
- [Ignoring Issues](Ignoring-Issues.md) — suppress a finding you accept
