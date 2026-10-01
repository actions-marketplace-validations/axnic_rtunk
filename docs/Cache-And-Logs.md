# Cache And Logs

rtunk downloads tools and plugin sources into a cache outside your repository and writes one log
per run. Override the location with `--cache-dir` or `RTUNK_CACHE_DIR`; free the space with
`rtunk cache clean`.

```bash
RTUNK_CACHE_DIR=/data/rtunk-cache rtunk check
```

## Where the cache lives

The location resolves in this order:

1. `--cache-dir <path>`
2. the `RTUNK_CACHE_DIR` environment variable
3. the OS cache directory, under `rtunk` (`~/Library/Caches` on macOS, `$XDG_CACHE_HOME` or
   `~/.cache` on Linux)

There is no setting for it in `rtunk.yaml`. The cache root holds these subdirectories:

| Directory    | Content                                                                   |
| ------------ | ------------------------------------------------------------------------- |
| `downloads/` | Downloaded and extracted tools and runtimes, plus the shims that run them |
| `plugins/`   | Fetched plugin sources (the `trunk` plugin repository, for example)       |
| `registry/`  | Parsed plugin source definitions                                          |
| `logs/`      | Run logs, one subdirectory per repository                                 |

> [!NOTE]
> The cache is shared by all your repositories. A tool pinned to the same version in two
> repositories is downloaded once.

## Clean and prune

```bash
rtunk cache prune
rtunk cache clean
```

| Command             | Effect                                                            |
| ------------------- | ----------------------------------------------------------------- |
| `rtunk cache prune` | Removes entries no existing, configured repository needs any more |
| `rtunk cache clean` | Removes the downloads, plugins, logs and registry subtrees        |

`prune` succeeds silently. `clean` shows one line per removed subtree with the space freed (`removed <name> (<size>)` when piped), or `The cache is already empty.` `prune` has no age option: an entry is kept while a repository that still
exists on disk lists it. `clean` removes only those four subtrees, leaving other files in the cache directory alone; the next run downloads everything again.

> [!WARNING]
> `cache clean` also deletes the run logs of every repository.

## Run logs

Runs such as `rtunk check` write a log. List and read them with `rtunk logs`:

```console
$ rtunk logs list
2026-10-01T09:40:24+02:00  check         ok                 0s  20261001T074024.005283000Z-check

$ rtunk logs show
run 2026-10-01T07:40:24.005494Z  (rtunk dev)
  argv: rtunk check
  cwd: /home/me/project
  repo: /home/me/project
  config: /home/me/project/.rtunk/rtunk.yaml
  concurrency: 10

status: ok in 0ms
```

`logs list` prints the time, the command, the status, the duration and the run name. `logs show`
takes a run name (or a unique prefix of it), or `latest`, which is the default. Add `--json` to
print the raw JSON Lines instead of the text rendering.

`rtunk logs clean` deletes this repository's logs; `--all` deletes every repository's logs.
With no logs, `logs list` and `logs show` report `runlog: no runs logged for this repository`.

## The `.rtunk/` symlinks

`rtunk init` also creates symlinks inside `.rtunk/` that point into the cache, so you can browse
logs, installed tools and plugin sources from the repository:

| Path                  | Points to                          |
| --------------------- | ---------------------------------- |
| `.rtunk/logs`         | This repository's log directory    |
| `.rtunk/tools/<name>` | The shim of each installed tool    |
| `.rtunk/plugins/<id>` | The checkout of each plugin source |

`.rtunk/.gitignore` lists `logs`, `tools` and `plugins`, so the links stay out of git. After
`cache clean` or a cache move, the links dangle; `rtunk toolbox link` rebuilds them.

```console
$ rtunk toolbox link
linked /home/me/project/.rtunk
```

## Where to go next

- [Command Reference](Command-Reference.md#rtunk-cache-clean) — every flag of `cache` and `logs`
- [Configuration Reference](Configuration-Reference.md#cache-directory) — cache directory
  resolution
- [Actions And Git Hooks](Actions-And-Git-Hooks.md) — `actions history` shows action runs
