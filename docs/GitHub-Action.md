# GitHub Action

`axnic/rtunk` is a GitHub Action that installs rtunk, caches its tools and plugins, runs `rtunk
check` and reports the findings as annotations and a job summary. It is rtunk's counterpart of
`trunk-io/trunk-action`, on Linux and macOS runners.

```yaml
jobs:
  rtunk:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - uses: actions/checkout@<sha> # pin by full SHA
        with:
          fetch-depth: 0
      - uses: axnic/rtunk@<sha> # pin by full SHA
        with:
          check-mode: changed-since-base
```

## Inputs

| Input                 | Default        | Meaning                                                                                                      |
| --------------------- | -------------- | ------------------------------------------------------------------------------------------------------------ |
| `version`             | `latest`       | `latest` (newest published release), a tag such as `v0.13.2`, or `source`                                    |
| `check-mode`          | `all`          | `all` or `changed-since-base`                                                                                |
| `arguments`           | empty          | Extra arguments for `rtunk check`, split on spaces with no quoting, e.g. `--filter=shellcheck --exclude=a,b` |
| `working-directory`   | `.`            | Directory rtunk runs in, relative to the workspace                                                           |
| `require-attestation` | `false`        | Fail when the release has no signature to verify, instead of falling back to the checksum                    |
| `token`               | `github.token` | Token the `gh` CLI uses to resolve and download the release                                                  |

| Output         | Meaning                                                      |
| -------------- | ------------------------------------------------------------ |
| `version`      | The version that ran: a tag, or `source`                     |
| `exit-code`    | Exit code of `rtunk check`: `0` clean, `1` findings or error |
| `results-file` | Path of the JSON report; empty when no file was selected     |

The action sets `--format json` itself, so `arguments` must not contain `--format`.

## Choose which files are checked

- `all` runs `rtunk check .`: every file under `working-directory` that git tracks or could track.
- `changed-since-base` runs `rtunk check --from <base>`, where `<base>` is the base commit of the
  pull request, the merge-queue entry, or the previous head of a push. On any other event, or for
  the first push of a branch, there is no base and rtunk checks the changes since `HEAD` (see
  [Checking code](Checking-Code.md#choose-which-files-to-check)).

`changed-since-base` needs the base commit in the checkout: use `actions/checkout` with
`fetch-depth: 0`. When the base is missing, the action tries to fetch it, then fails with an error
that says so.

## Install and verify rtunk

With a release version, the action downloads `rtunk-<tag>-<os>-<arch>.tar.gz` and verifies it as
described in [SECURITY.md](../SECURITY.md#verifying-a-release):

1. The archive matches `checksums.txt`.
2. When the release carries `checksums.txt.sigstore.json`, `gh attestation verify` checks the
   archive's SLSA provenance against the Release workflow of `axnic/rtunk`, and `cosign
verify-blob` checks the signature of `checksums.txt` when `cosign` is on `PATH`. Any failure
   fails the job.

Releases published before signing was introduced (up to `v0.13.2`) have no signature. For them the
action emits a warning and relies on the checksum alone, which only detects a corrupted download,
not a tampered release. Set `require-attestation: true` to refuse them, and pin a recent `version`.

`version: source` builds rtunk from the checkout of the action itself with `go build` (Go is set up
automatically). It is meant for a repository that hosts rtunk, to check itself with the code under
review, and verifies nothing since there is nothing to download.

## Cache

The action restores and saves `downloads/`, `plugins/` and `registry/` of
[rtunk's cache](Cache-And-Logs.md#where-the-cache-lives) (`~/.cache/rtunk`, set through
`RTUNK_CACHE_DIR`) with `actions/cache`. Run logs are not cached.

| Part        | Value                                                                                  |
| ----------- | -------------------------------------------------------------------------------------- |
| Key         | `rtunk-v1-<os>-<arch>-<rtunk version>-<hash of the config>`                            |
| Restore key | `rtunk-v1-<os>-<arch>-<rtunk version>-`: the newest cache of the same rtunk version    |
| Config hash | `.rtunk/*.yaml`, `.rtunk/*.lock`, `.trunk/*.yaml` and `rtunk.lock` of the project root |

The project root is the nearest ancestor of `working-directory` holding `.rtunk` or `.trunk`. For
`version: source`, the rtunk version is the hash of the built binary. The cache is saved with
`if: always()`, so a run whose checks fail still keeps the tools it downloaded. An exact hit is not
saved again. Linters that a restore-key hit does not cover are downloaded and saved under the new
key.

## Results

`rtunk check` runs with `--format json`. Progress lines stay in the step log, and the action turns
the report into:

- one annotation per issue, on its file and line: `high` is an error, `medium` a warning, `low` a
  notice. A linter that failed to run is an error annotation. The log shows the first 50 issues, and
  GitHub itself displays at most 10 annotations per level and step;
- a job summary with the counts and a table of the first 100 issues.

The job fails in the last step, after the cache is saved and the results reported, with the exit
code of rtunk (see [exit codes](Checking-Code.md#use-rtunk-in-ci)). The `results-file` output
holds the whole report for a later step, such as one that uploads it.

## Limits

- Windows runners are not supported, like rtunk itself.
- `arguments` is split on spaces; an argument containing a space cannot be passed.
- The action needs the `gh` CLI (present on GitHub-hosted runners) to resolve and download a
  release, and `jq` to report results.

## Where to go next

- [Checking code](Checking-Code.md#use-rtunk-in-ci) — flags, file selection and exit codes
- [Cache and logs](Cache-And-Logs.md) — what the cache holds
- [Installation](Installation.md) — installing rtunk outside CI
