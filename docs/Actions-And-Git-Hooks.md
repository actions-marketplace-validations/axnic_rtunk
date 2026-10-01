# Actions And Git Hooks

An action is a command that rtunk runs on demand or when a git event fires, such as `pre-commit`.
List actions with `rtunk actions list`, enable them, then install the hooks with
`rtunk git-hooks sync`.

```console
$ rtunk actions enable trunk-check-pre-commit
$ rtunk git-hooks sync
installed: pre-commit
```

## List actions

```console
$ rtunk actions list
Available (not enabled)
  ◯ buf-gen                      Run 'buf generate' anytime a .proto file changes. ...
  ◯ commitlint                   Enforce git commit message standards
  ◯ go-mod-tidy                  Runs go mod tidy when changes are detected to go.mod
  ...

Enable one with: rtunk actions enable <id>
```

Actions are grouped as `Enabled` and `Available (not enabled)`. `--format json` prints the same
data as JSON.

## Enable and disable actions

```bash
rtunk actions enable trunk-check-pre-commit go-mod-tidy
rtunk actions disable go-mod-tidy
```

Both commands edit `.rtunk/rtunk.yaml`. Disabling moves the id from `actions.enabled` to
`actions.disabled`:

```yaml
actions:
  enabled:
    - trunk-check-pre-commit
  disabled:
    - go-mod-tidy
```

As with linters, `rtunk actions enable` without ids opens an interactive picker (terminal only).

## Install git hooks

`rtunk git-hooks sync` writes a hook script into `.git/hooks/` for every git hook that an enabled
action triggers. The script calls back into rtunk (`rtunk actions run --hook <hook>`), so a hook
runs every enabled action bound to it.

```console
$ rtunk git-hooks sync
installed: pre-commit
$ rtunk git-hooks unsync
removed: pre-commit
```

`sync` is idempotent: running it again rewrites the same files. `unsync` removes only hooks that
rtunk installed.

> [!NOTE]
> rtunk never overwrites a hook it did not write. `sync` prints `skipped (foreign hook, use
--force): pre-commit` and leaves the file alone; add `--force` to replace it.

The hook script embeds the path of the rtunk binary that ran `sync`. Re-run `sync` after moving or
reinstalling rtunk.

## Run actions

Run one action by id, or every action a git hook triggers:

```bash
rtunk actions run trunk-announce
rtunk run trunk-announce
rtunk run --hook pre-commit
```

`rtunk run` is a shortcut for `rtunk actions run`. Arguments after `--` are forwarded to the
action. Only enabled actions run: `rtunk run go-mod-tidy` fails with `unknown action
"go-mod-tidy"` while that action is disabled. A failing action makes the command fail:

```console
$ rtunk run trunk-announce
rtunk: actions: trunk-announce: exit code 109
```

## Review past runs

```console
$ rtunk actions history
2026-10-01T09:40:18+02:00  trunk-check-pre-commit  pre-commit  exit 1  434ms
2026-10-01T09:40:17+02:00  trunk-announce          manual      exit 109  535ms
```

Each line shows the time, the action, what triggered it (`manual` or the git hook), the exit code
and the duration. Filter with `--id <action-id>` and cap the output with `--limit` (default 20).

## Where to go next

- [Command Reference](Command-Reference.md#rtunk-actions-run) — every flag of `actions`, `run` and
  `git-hooks`
- [Cache And Logs](Cache-And-Logs.md) — where run logs are stored
- [Managing Linters](Managing-Linters.md) — the same enable and disable workflow for linters
