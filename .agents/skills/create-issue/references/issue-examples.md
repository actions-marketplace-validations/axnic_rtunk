# Issue examples

Illustrative: `axnic/rtunk` has no filed issues yet, so these are built from real repository facts
(merged PRs #27, #30, #32, `.commitlintrc.js`, `ROADMAP.md`) and are meant as shape references, not
as records. Replace them with real issues as they appear (`gh issue list -R axnic/rtunk`).

## Bug: evidenced, with a reproduction

Title: `rtunk check --format github loses annotations when the action runs version: latest`

Issue Type `Bug`, labels `bug`, `type::bug`, `priority::medium`.

```markdown
## Context

`action.yml` probes `rtunk check run --help` for `--format github` (#27). `latest` is v0.14.0 today,
which lacks it, so the step falls back to the human format and emits a `::warning::`. Users on
`version: latest` get no annotations and no job summary until v0.15 ships.

rtunk version: v0.14.0 (release archive), runner ubuntu-latest, workflow using `uses: axnic/rtunk@main`.

## Proposal

Reproduction:

1. Use the action with `version: latest` on a repo with a failing linter.
2. Observe: `::warning::` that annotations are unavailable, no annotations, no job summary.
3. Expected: annotations as with `version: source`.

Not fixable before the next release; track so the warning can be removed once `latest` carries
the format.

## Acceptance criteria

- [ ] After the release that ships `--format github`, `version: latest` produces annotations and a job summary
- [ ] The fallback warning is removed or reduced to the pinned-older-version case

## Notes

Follow-up to #27.
```

## Feature: Problem / Proposed behavior, grounded in `ROADMAP.md`

Title: `Add rtunk.lock to verify downloaded tools against recorded fingerprints`

Issue Type `Feature`, labels `enhancement`, `type::feature`, `priority::low`.

```markdown
## Context

A downloaded tool is trusted the first time it is fetched, with no independent check afterwards
(`AGENTS.md`, `docs/CLI-Design.md` "Download integrity"); a source compromised at that moment is not
detected. `ROADMAP.md` schedules a lock file for v1.1, after v1.0.

## Proposal

A committed `rtunk.lock` recording the expected fingerprint of every installable tool per version
and platform; a mismatch blocks the run with no fallback; a missing entry is recorded normally and
is an error under `--locked` or `CI=true`; `rtunk lock` precomputes other platforms.

## Acceptance criteria

- [ ] A tool whose fingerprint mismatches `rtunk.lock` makes the run fail with an error
- [ ] A missing entry is recorded normally, and fails under `--locked`/`CI=true`
- [ ] `rtunk lock` adds fingerprints for a platform other than the current one
- [ ] Scope check: unrelated to the daemon, Merge Queue, Flaky Tests and hosted services

## Notes

`ROADMAP.md`, section "v1.1 - Download integrity". Filed only when work on it is about to start.
```

## Task: follow-up deferred from a PR

Title: `Drop the --format github fallback warning from action.yml once latest ships it`

Issue Type `Task`, labels `type::chore`, `priority::low`.

```markdown
## Context

#27 made `action.yml` probe `rtunk check run --help` for `--format github` and fall back to the
human format with a `::warning::` on releases that lack it (v0.14.0 and earlier). Once `latest`
carries the format the probe is dead weight.

## Proposal

Remove the probe and the warning, keep the plain `rtunk check ... --format github` step. Wait for
the first release that includes #27.

## Acceptance criteria

- [ ] `action.yml` has no `--help` probe and no fallback warning
- [ ] The "Annotations and job summary" section of `docs/GitHub-Action.md` no longer mentions the fallback
- [ ] The repo's `rtunk` CI job (`uses: ./`) is green

## Notes

Follow-up to #27 and #30. Blocked until the next release (`status::blocked` once someone picks it up).
```

## Bad: what to avoid

Title: `+[docs]: Improve docs` with a 200-line body.

- Commit format in the title; the outcome is unreadable.
- No Issue Type or `type::*` label.
- Body repeats the title, lists an empty risk table, "alternatives considered" with one entry, and
  a step-by-step implementation plan that belongs in the PR.
- Filed for something already done in an open PR (no duplicate check).
