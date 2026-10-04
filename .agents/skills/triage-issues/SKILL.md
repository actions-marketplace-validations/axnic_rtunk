---
name: triage-issues
description: Use when triaging the rtunk issue backlog or when an issue needs a priority::* label ("triage the issues", "what priority should this be", "how urgent is this", "set priorities", "check labels on #N"). Applies a calibrated decision procedure to set exactly one of priority::high, priority::medium, priority::low and to verify the type::* label and GitHub Issue Type, using only labels that exist in the repo. Priority tracks whether something is broken now and whether there is a stated reason to act, not how dangerous the area is. Proposes a table first and applies labels only after confirmation.
---

# Triage issues

Repo: `axnic/rtunk` (always `-R axnic/rtunk`). A triaged issue has exactly one `priority::*`
(`high`, `medium`, `low`), exactly one `type::*` (`bug`, `feature`, `docs`, `chore`, `security`)
consistent with its Issue Type (`Bug`, `Feature`, `Task`). No other labels are created or applied
here: there is no `size::*` label, `status::*` belongs to the person doing the work, and `bug`,
`enhancement`, `documentation` are GitHub defaults left as they are.

Filing new issues is the `create-issue` skill; this one sets priority on new and existing issues.

## Priority procedure

Priority answers one question: is there an actual, current reason to do this now? Danger of the
area is not a reason. Walk the rungs in order and stop at the first that applies. Precedents in
`references/calibration-examples.md` (cited as C1 to C5).

1. **Out of scope or a vulnerability report: stop.** A permanently out-of-scope request (daemon,
   Merge Queue, Flaky Tests, hosted/SaaS, cloud accounts: `AGENTS.md`) is not prioritized: propose
   closing it (`wontfix`) to the user. A public issue that reads as a vulnerability report: tell the
   user to move it to a private advisory (`SECURITY.md`); do not label it.
2. **Broken now and relied on: `high`.** Something that currently fails and that users or the
   maintainer depend on: `check`/`fmt` giving wrong or missing results, a crash, a broken
   release or CI path (red `main`, failing release workflow). Currently failing, not "could fail
   eventually" (C1). The same breakage in something nobody uses is `low`, and the right move is
   usually to propose removal (C3).
3. **Security exposure scales multiplicatively.** A hardening gap with nothing exploiting it, or a
   dependency CVE that `govulncheck` shows as not reachable: `medium`. The same gap on the path that
   downloads, verifies, or executes tools, or on the release pipeline, or a reachable CVE:
   `high` (C4).
4. **Expiry follows a curve** for any certificate, token or signing identity: more than about 30
   days out `medium`; about 15 days or less `high`; expired `high` (there is no higher band).
5. **A recurring annoyance that never blocks: `medium`.** A workaround exists (a flag, a manual
   step) but users keep hitting it.
6. **A stated reason moves work up; none keeps it `low`.** A concrete trigger (a `ROADMAP.md`
   milestone being worked on now, a user report, another issue that needs it) makes a feature or
   refactor `medium`, or `high` if it blocks the milestone. A refactor, cleanup or feature with no
   rationale, or scheduled for a later milestone and not started, is `low` however large it is (C2).
7. **Dependency bumps** (Renovate, Dependabot security updates and manual): driven by a feature you need `medium`; by a
   reachable CVE `high` (rung 3); routine `low`. A bump of `go.mod`'s Go version or a core tool the
   CI depends on is `medium` even when routine.
8. **Documentation:** docs that mislead users into doing the wrong thing, or contradict the
   behavior (verify against the code first) `medium`; wording, structure, typos `low` (C5).
9. **Default `low`** when nothing above applies. Never raise priority because the issue is long,
   confidently written, or filed by a bot.

## Type check

Verify, do not rewrite blindly: one `type::*` label that matches the Issue Type.

| Issue Type | `type::*` allowed                         |
| ---------- | ----------------------------------------- |
| `Bug`      | `bug` (or `security` if hardening a flaw) |
| `Feature`  | `feature`                                 |
| `Task`     | `docs`, `chore`, `security`               |

Use the resolution tests in `create-issue` for ambiguous cases. A missing or mismatched type is
fixed in the same pass and listed in the table.

## Workflow

1. Gather untriaged issues (no priority label), or take the issues the user names:

   ```bash
   gh issue list -R axnic/rtunk --state open \
     --search '-label:"priority::high" -label:"priority::medium" -label:"priority::low"' \
     --json number,title,labels,issueType,createdAt
   ```

2. Read each body (`gh issue view <n> -R axnic/rtunk --comments`), not just the title: the
   signal (is it broken now, is a reason stated, who is affected) is in Context.
3. Verify claims that decide the rung against the repo or CI (a "broken" workflow: look at the
   run; a "docs are wrong": read the code). A claim you cannot verify is stated as such in the
   rationale.
4. Draft the table: number, title, current labels, proposed `priority::*`, proposed `type::*` change
   if any, rung and one-line rationale (`reason stated: ...` or `no reason stated`, `broken now: ...`).
5. Show the table and wait for confirmation. Bulk label edits on the backlog are visible and tedious
   to undo.
6. Apply, replacing any stale priority label, then re-read the labels to confirm (a multi-flag
   `gh issue edit` can silently no-op):

   ```bash
   gh issue edit <n> -R axnic/rtunk --remove-label "priority::low" --add-label "priority::high"
   gh issue view <n> -R axnic/rtunk --json labels --jq '[.labels[].name]'
   ```

   `--remove-label` for a label the issue does not carry is harmless; omit it when there is none.

7. A case no rung settles: do not force it. Use the nearest calibration example, else ask the user,
   and add the resolved case to `references/calibration-examples.md`.

## Quick reference

| Question                                                            | Priority                       |
| ------------------------------------------------------------------- | ------------------------------ |
| Out-of-scope request                                                | none: propose closing          |
| Looks like a vulnerability report                                   | none: move to private advisory |
| Failing now, relied on (wrong results, red `main`, release)         | `high`                         |
| Failing now, nobody uses it                                         | `low` (propose removal)        |
| Security gap, no exploitation, off the download/verify/release path | `medium`                       |
| Security gap on download/verify/release path, or reachable CVE      | `high`                         |
| Recurring annoyance with a workaround                               | `medium`                       |
| Feature/refactor with a concrete current trigger                    | `medium` (`high` if it blocks) |
| Feature/refactor with no stated reason or later milestone           | `low`                          |
| Docs that mislead / docs that are cosmetic                          | `medium` / `low`               |
| Dependency: needed feature / reachable CVE / routine                | `medium` / `high` / `low`      |

## References

- `references/calibration-examples.md`: the calibrated precedents C1 to C5
- `.agents/skills/create-issue/SKILL.md`: filing, Issue Type and `type::*` rules
- `SECURITY.md`, `AGENTS.md` ("Permanently out of scope")
