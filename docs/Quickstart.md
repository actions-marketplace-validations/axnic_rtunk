# Quickstart

Go from an empty git repository to your first `rtunk check` report in a few minutes: initialize,
enable one linter, run the check, then fix or ignore what it finds. This page assumes
[rtunk is installed](Installation.md) and that you have `git` and network access (rtunk downloads
the linter on first use).

```bash
rtunk init && rtunk linters enable yamllint && rtunk check
```

## Initialize rtunk in a repository

Run `init` at the root of a git repository. This guide uses a repository holding one badly
formatted YAML file.

```console
$ printf 'name:   test\nlist: [a,  b]\n' > example.yaml
$ rtunk init
initialized rtunk at /path/to/repo/.rtunk/rtunk.yaml
next: rtunk linters enable <linter>, rtunk actions enable <action>, rtunk git-hooks sync
linked /path/to/repo/.rtunk
```

`init` creates `.rtunk/rtunk.yaml` with no linter enabled. It pins the
[trunk plugins](https://github.com/trunk-io/plugins) repository as the source of linter
definitions.

> [!WARNING]
> Coming from trunk? Do not run `init`. If `.trunk/trunk.yaml` exists, the new `.rtunk/rtunk.yaml`
> takes precedence and `.trunk/trunk.yaml` stops being read. Run `rtunk check` directly instead; see
> [Migrating from trunk](Migrating-From-Trunk.md).

## Enable a linter

```bash
rtunk linters enable yamllint
```

The command prints nothing on success and adds `yamllint` to `lint.enabled` in
`.rtunk/rtunk.yaml`:

```yaml
lint:
  enabled:
    - yamllint
```

## Run the check

With no paths, `rtunk check` checks the files that changed. In a repository without commits, that
is every file. rtunk downloads and installs yamllint on the first run, so that run takes longer.

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
rtunk: rtunk: check found 4 issue(s)
```

Each finding reads `line:column  severity  message  linter/rule`. The command exits `1` when it
finds issues and `0` when the tree is clean, so it works as a CI gate. `check` is read-only: it
never modifies files.

## Fix or ignore what it found

Fix the file, then check it again. Pass a path to check one file regardless of git state:

```console
$ printf -- '---\nname: test\nlist: [a, b]\n' > example.yaml
$ rtunk check example.yaml
✔ yamllint         done     clean
Checked 1 file with 1 linter in 0.1s
✔ no issues
```

To accept a finding instead, add an `rtunk-ignore` comment on the offending line, naming the rule:

```console
$ cat example.yaml
---
name: test
list: [a,  b]  # rtunk-ignore(yamllint/commas): demo
$ rtunk check example.yaml
✔ yamllint         done     clean
Checked 1 file with 1 linter in 0.1s
✔ no issues · 1 suppressed
```

> [!NOTE]
> yamllint requires two spaces before an inline comment, as above. `trunk-ignore(...)` comments
> work the same way. Every form is in the
> [ignore syntax reference](Ignoring-Issues.md).

## Where to go next

- [Checking code](Checking-Code.md) — file selection, `--fix`, output formats and exit codes
- [Managing linters](Managing-Linters.md) — list, enable and disable linters
- [Formatting code](Formatting-Code.md) — formatters run through `rtunk fmt`, not `check`
- [Actions and git hooks](Actions-And-Git-Hooks.md) — run `check` automatically before a commit
