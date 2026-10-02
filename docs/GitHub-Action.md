# GitHub Action

`axnic/rtunk` is a GitHub Action that installs rtunk, verifies it, caches its tools and plugins,
runs `rtunk check` and reports the findings as annotations and a job summary. It is rtunk's
counterpart of `trunk-io/trunk-action`, on Linux and macOS runners. It is a composite action
defined by [`action.yml`](../action.yml) at the root of the repository.

```yaml
name: rtunk
on: pull_request

permissions: {}

jobs:
  rtunk:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - uses: actions/checkout@<sha> # pin by full SHA
        with:
          fetch-depth: 0 # changed-since-base needs the base commit
      - uses: axnic/rtunk@v0.14.0
        with:
          version: v0.14.0
          check-mode: changed-since-base
          require-attestation: true
```

The repository must have an rtunk configuration (`.rtunk/rtunk.yaml`, or an existing
`.trunk/trunk.yaml`): without a `.rtunk` or `.trunk` ancestor, `rtunk check` refuses to run and the
job fails.

## Set up a workflow

1. **Pin the action.** `uses: axnic/rtunk@v0.14.0` pins by tag. Tags can be moved, so a workflow
   that must be immutable pins the commit instead, with the tag in a comment:

   ```yaml
   - uses: axnic/rtunk@2e4dd9dc340083cf8d318e576ef0114b65a398e0 # v0.14.0
   ```

2. **Pin rtunk itself.** The action version and the rtunk version are independent: `uses:` selects
   the code of the action, `version:` selects the rtunk release it installs, and `version` defaults
   to `latest`. Set both so a new release never changes your pipeline unannounced.
3. **Grant `contents: read`.** It is the only permission the action asks for: the cache, the
   annotations and the summary do not use the `GITHUB_TOKEN`. Start from `permissions: {}` at the
   top of the workflow and grant it to the job.
4. **Check out with enough history** when `check-mode` is `changed-since-base`: `actions/checkout`
   with `fetch-depth: 0` (see [Choose which files are checked](#choose-which-files-are-checked)).
5. **Use a Linux or macOS runner** (`amd64` or `arm64`). The runner needs the `gh` CLI and `jq`,
   both present on GitHub-hosted runners.

## Inputs and outputs

Defaults are those of [`action.yml`](../action.yml).

| Input                 | Default        | Meaning                                                                                                      |
| --------------------- | -------------- | ------------------------------------------------------------------------------------------------------------ |
| `version`             | `latest`       | `latest` (newest published release), a tag such as `v0.14.0`, or `source`                                    |
| `check-mode`          | `all`          | `all` or `changed-since-base`                                                                                |
| `arguments`           | empty          | Extra arguments for `rtunk check`, split on spaces with no quoting, e.g. `--filter=shellcheck --exclude=a,b` |
| `working-directory`   | `.`            | Directory rtunk runs in, relative to the workspace                                                           |
| `require-attestation` | `false`        | Fail when the release has no signature bundle to verify, instead of falling back to the checksum alone       |
| `token`               | `github.token` | Token the `gh` CLI uses to resolve, download and verify the release                                          |

| Output         | Meaning                                                                                        |
| -------------- | ---------------------------------------------------------------------------------------------- |
| `version`      | The version that ran: a tag, or `source`                                                       |
| `exit-code`    | Exit code of `rtunk check`: `0` clean, `1` findings or error                                   |
| `results-file` | Path of the JSON report (`rtunk-results.json` in `RUNNER_TEMP`); empty if no file was selected |

The action sets `--format json` itself, so `arguments` must not contain `--format`.

## Choose which files are checked

- `all` runs `rtunk check .`: every file under `working-directory` that git tracks or could track.
- `changed-since-base` runs `rtunk check --from <base>`, where `<base>` is the base commit of the
  pull request (`pull_request`, `pull_request_target`), of the merge-queue entry (`merge_group`), or
  the previous head of a push (`push`). On any other event, or for the first push of a branch (an
  all-zero previous head), there is no base and rtunk checks the changes since `HEAD` (see
  [Checking code](Checking-Code.md#choose-which-files-to-check)).

> [!WARNING]
> `changed-since-base` needs the base commit in the checkout: use `actions/checkout` with
> `fetch-depth: 0`. When the base is missing, the action tries to fetch it (unshallowing a shallow
> clone), then fails with an error that says so.

On `schedule` and `workflow_dispatch` there is no base: use `check-mode: all`, the default.

## Choose the rtunk version

| `version` | What runs                                                                            |
| --------- | ------------------------------------------------------------------------------------ |
| `latest`  | The tag of the newest published release (`gh api repos/axnic/rtunk/releases/latest`) |
| `vX.Y.Z`  | That release (a semver prerelease suffix such as `v1.0.0-rc.1` is accepted)          |
| `source`  | rtunk built with `go build` from the action's own checkout (Go is set up for you)    |

Any other value fails the job. `latest` moves with every release: pin an exact tag for
reproducible runs. Use `source` only in a repository that hosts rtunk, to check itself with the
code under review; this repository's own CI does exactly that, and `source` verifies nothing since
nothing is downloaded.

Releases before `v0.14.0` can still be installed through the action, but they are unsigned (see
below) and declare the previous Go module path.

## Install and verify rtunk

With a release version, the action downloads `rtunk-<tag>-<os>-<arch>.tar.gz` and `checksums.txt`
and verifies the archive as described in [SECURITY.md](../SECURITY.md#verifying-a-release):

1. The archive matches `checksums.txt`.
2. When the release carries `checksums.txt.sigstore.json`, `gh attestation verify` checks the
   archive's SLSA build provenance against `axnic/rtunk`, and `cosign verify-blob` checks the
   signature of `checksums.txt` against the Release workflow's identity when `cosign` is on `PATH`.
   Any failure fails the job.

`v0.14.0` is the first release signed with cosign and carrying SBOMs and SLSA provenance, so every
release from `v0.14.0` on has a bundle to verify. Releases up to `v0.13.2` have none: the action
warns (`rtunk <tag> is not signed: verified against checksums.txt only`) and relies on the checksum,
which only detects a corrupted download, not a tampered release.

> [!TIP]
> For `v0.14.0` and later, set `require-attestation: true`. A release without a bundle is then
> refused instead of accepted with a warning, so a release stripped of its signature cannot slip
> through. The input only decides what happens when the bundle is absent; a bundle that is present
> is always verified.

Hosted runners do not necessarily ship `cosign`. To also check the signature of `checksums.txt`,
install it before the action, for example with `sigstore/cosign-installer`; the attestation check
runs either way.

## Cache

The action restores and saves `downloads/`, `plugins/` and `registry/` of
[rtunk's cache](Cache-And-Logs.md#where-the-cache-lives) (`~/.cache/rtunk`, set through
`RTUNK_CACHE_DIR`) with `actions/cache`. Run logs are not cached.

| Part        | Value                                                                                  |
| ----------- | -------------------------------------------------------------------------------------- |
| Key         | `rtunk-v1-<os>-<arch>-<rtunk version>-<hash of the config>`                            |
| Restore key | `rtunk-v1-<os>-<arch>-<rtunk version>-`: the newest cache of the same rtunk version    |
| Config hash | `.rtunk/*.yaml`, `.rtunk/*.lock`, `.trunk/*.yaml` and `rtunk.lock` of the project root |

The project root is the nearest ancestor of `working-directory` holding `.rtunk` or `.trunk`. With
no config file the hash is `no-config`. For `version: source`, the rtunk version is a hash of the
built binary. The cache is saved with `if: always()`, so a run whose checks fail still keeps the
tools it downloaded. An exact hit is not saved again (cache keys are immutable). Linters that a
restore-key hit does not cover are downloaded and saved under the new key.

Cache entries follow GitHub's scoping rules: a pull request reads caches saved on its base branch
and saves its own under its merge ref, so a run on `main` (a push, or a nightly job) is what warms
the cache for later pull requests.

## Results

`rtunk check` runs with `--format json`. Progress lines stay in the step log, and the action turns
the report into:

- one annotation per issue, on its file and line, titled `<linter>/<rule>`: `high` is an error,
  `medium` a warning, `low` a notice. A linter that failed to run is an error annotation. The log
  receives the first 50 issues, and GitHub itself displays at most 10 annotations per level and
  step; on a pull request they appear in the Files changed tab and the checks list;
- a job summary titled `rtunk check`: counts of files, linters, issues, failures and suppressed
  issues, the run time, then a table of the first 100 issues.

The job fails in the last step, after the cache is saved and the results reported, with the exit
code of rtunk (see [exit codes](Checking-Code.md#use-rtunk-in-ci)). The `results-file` output
holds the whole report for a later step, such as one that uploads it.

## Examples

### Pull requests: changed files only

The common setup: check what the pull request changes, and what the merge queue is about to merge.

```yaml
name: rtunk
on:
  pull_request:
  merge_group:

permissions: {}

jobs:
  rtunk:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - uses: actions/checkout@<sha>
        with:
          fetch-depth: 0
      - uses: axnic/rtunk@v0.14.0
        with:
          version: v0.14.0
          check-mode: changed-since-base
          require-attestation: true
```

### Nightly: every file

A scheduled run has no base. It checks the whole tree and keeps the cache warm for pull requests.

```yaml
name: rtunk nightly
on:
  schedule:
    - cron: 0 3 * * *
  workflow_dispatch:

permissions: {}

jobs:
  rtunk:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - uses: actions/checkout@<sha>
      - uses: axnic/rtunk@v0.14.0
        with:
          version: v0.14.0
```

### Monorepo: one job per project

`working-directory` selects where rtunk runs. The configuration is the nearest `.rtunk` or `.trunk`
above that directory, up to the workspace root, and the cache key hashes that configuration.

```yaml
jobs:
  rtunk:
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      matrix:
        project: [services/api, services/web]
    permissions:
      contents: read
    steps:
      - uses: actions/checkout@<sha>
        with:
          fetch-depth: 0
      - uses: axnic/rtunk@v0.14.0
        with:
          version: v0.14.0
          working-directory: ${{ matrix.project }}
          check-mode: changed-since-base
```

Legs that resolve to the same project root share one cache key.

### Narrow the run and keep the report

`arguments` forwards flags to `rtunk check`; `results-file` hands the JSON report to later steps.

```yaml
- uses: axnic/rtunk@v0.14.0
  id: rtunk
  with:
    version: v0.14.0
    arguments: --filter=shellcheck,yamllint --security-only
- if: always() && steps.rtunk.outputs.results-file != ''
  uses: actions/upload-artifact@<sha>
  with:
    name: rtunk-report
    path: ${{ steps.rtunk.outputs.results-file }}
```

### Forks and private repositories

- **Pull requests from forks** run with a read-only `GITHUB_TOKEN` and no secrets. That is enough:
  the action only reads the release, and annotations and the summary are workflow commands, not API
  calls. Use the `pull_request` trigger. With `pull_request_target`, the workflow runs with the
  base repository's privileges, and rtunk executes the linters and plugin sources that the checked
  out configuration declares: never check out and run untrusted pull request code under that
  trigger.
- **Private repositories** work the same way: the action downloads rtunk from the public
  `axnic/rtunk` releases and never sends your code anywhere. It configures no credentials for
  private plugin sources; a `plugins.sources` entry that rtunk cannot clone with the checkout's
  credentials fails the run.
- **Self-hosted runners** must provide `gh` and `jq`, run Linux or macOS, and be `amd64` or
  `arm64`.

## Troubleshooting

| Message or symptom                                                        | Cause and fix                                                                                                                                                  |
| ------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `rtunk supports Linux and macOS runners only`                             | Windows runner. Use `ubuntu-latest` or `macos-latest`.                                                                                                         |
| `the gh CLI is required to ...`                                           | The runner has no `gh`. Install it, or use a GitHub-hosted image.                                                                                              |
| `no published rtunk release found; pin a version or use version: source`  | `latest` could not be resolved: no non-draft release, or the token cannot read it. Pin `version: vX.Y.Z`.                                                      |
| `version must be 'latest', 'source' or a tag like v0.13.2`                | Malformed `version`. Use the tag with its `v` prefix, as `v0.14.0`.                                                                                            |
| `base commit <sha> is not in the checkout; use actions/checkout with ...` | Shallow checkout with `changed-since-base`. Set `fetch-depth: 0` on `actions/checkout`.                                                                        |
| `rtunk <tag> has no signature bundle and require-attestation is true`     | The release predates signing (up to `v0.13.2`). Move to `v0.14.0` or later, or set `require-attestation: false`.                                               |
| `rtunk <tag> is not signed: verified against checksums.txt only`          | Warning for a release up to `v0.13.2`. Pin a newer version and set `require-attestation: true`.                                                                |
| `gh attestation verify` or `cosign verify-blob` fails                     | The archive or `checksums.txt` was not produced by the Release workflow. Do not use it; report it per [SECURITY.md](../SECURITY.md).                           |
| `check-mode must be 'all' or 'changed-since-base'`                        | Typo in `check-mode`.                                                                                                                                          |
| Job fails with `rtunk check exited with code 1`, no annotation            | A linter failed to run, the config is invalid, or no `.rtunk`/`.trunk` was found above `working-directory`. Read the step log and the summary's Failures list. |
| Green job, nothing checked (`results-file` empty)                         | No file was selected. With `changed-since-base`, no relevant file changed; on `schedule` use `all`.                                                            |
| Fewer annotations than issues                                             | GitHub shows at most 10 per level and step. The summary lists the first 100; `results-file` holds all.                                                         |
| Every run downloads the linters again                                     | The cache key changed (config or rtunk version) or the run is on a new branch with no base-branch cache. Warm it with a run on `main`.                         |

## Limits

- Windows runners are not supported, like rtunk itself.
- `arguments` is split on spaces; an argument containing a space cannot be passed.
- `token` is used by the `gh` CLI only: it does not authenticate plugin source clones.

## Where to go next

- [Checking code](Checking-Code.md#use-rtunk-in-ci) — flags, file selection and exit codes
- [Cache and logs](Cache-And-Logs.md) — what the cache holds
- [Installation](Installation.md) — installing rtunk outside CI
- [SECURITY.md](../SECURITY.md#verifying-a-release) — verify a release by hand
