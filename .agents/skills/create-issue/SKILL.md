---
name: create-issue
description: Use when work in the rtunk repository surfaces something that should be tracked as a GitHub issue ("this deserves its own issue", "we should do X later", "file a bug for this", "open an issue", "track this follow-up"): a bug found in passing, a follow-up that does not fit the current PR, a proposal worth recording, an upstream blocker. Runs the duplicate check with `gh`, sets a sentence-form title, the Issue Type (Bug, Feature, Task), the `type::*` and optional `priority::*` labels, and a Context / Proposal / Acceptance criteria / Notes body that respects `.github/ISSUE_TEMPLATE/`. Never for security vulnerabilities (private advisory instead).
---

# Create an issue

Repo: `axnic/rtunk`; always pass `-R axnic/rtunk` to `gh`. Issues are for work and information the
codebase does not already carry; sizing and prioritization beyond an obvious `priority::*` belong to
the `triage-issues` skill.

## 0. Never file these publicly

A suspected vulnerability (`SECURITY.md`): report it through
[GitHub Security Advisories](https://github.com/axnic/rtunk/security/advisories/new), never an issue,
PR or discussion. Stop and hand it to the user.

Out of scope for rtunk, closed without discussion (`AGENTS.md`, "Permanently out of scope"): the
trunk daemon, Merge Queue, Flaky Tests, hosted/SaaS dashboards, `login`/`logout`/`whoami`/cloud
accounts. Do not file a proposal for them.

## 1. When to file

All three must hold:

1. **Not in scope for the current change.** If it fits the current PR without growing it, do it
   there.
2. **Specific enough to act on.** "Improve performance" is not an issue; "`rtunk check` exits 0
   on a config with an unknown linter, with no warning" is.
3. **Carries information the repo does not.** A symptom, a decision rationale, a changed
   constraint. Restating `git log` or `ROADMAP.md` is noise (a `ROADMAP.md` item needs an issue only
   once someone is about to work on it).

Good triggers: a bug seen while doing something else; a follow-up from a PR that cannot ship in it;
a design tradeoff that needs a tracked next iteration; an upstream blocker to link back to.

Do not file: refactor ideas with no concrete trigger; "X is broken, I'll look later" without
evidence; questions the code answers.

## 2. Duplicate check

```bash
gh issue list -R axnic/rtunk --state all --search "<symptom or feature keywords>" -L 10
gh pr list    -R axnic/rtunk --state all --search "<keywords>" -L 5
```

Related issue found: comment on it, or file a distinct one and cross-link. Fixed already by a merged
PR: do not file.

## 3. Templates

`.github/ISSUE_TEMPLATE/` has `bug_report.md` (label `bug`) and `feature_request.md` (label
`enhancement`); `config.yml` sets `blank_issues_enabled: false`. Through the web UI the templates
are mandatory. With `gh issue create --body-file`, the template is not applied, so mirror its
sections:

- Bug: `rtunk version` (`rtunk --version`), `Platform`, `Configuration` (relevant excerpt of
  `.rtunk/rtunk.yaml` or `.trunk/trunk.yaml`), `Linter or plugin involved`, `Command run`,
  `Expected behavior`, `Actual behavior`, `Minimal reproduction`. Windows is not a supported host.
- Feature: `Problem`, `Proposed behavior`, and the `Scope check` checkbox (not a permanently
  out-of-scope item).
- Task (follow-up, chore, docs work): no template; use the body skeleton below.

Read the template files again before filing if they may have changed.

## 4. Title

Sentence form, read in an issue list one at a time. The commit format `type[scope]: Subject` is for
commits and PR titles only; do not use it here.

- Verb first for proposed work ("Add ...", "Reuse ...", "Document ..."); the symptom first for a bug
  ("`rtunk check` hangs on ...").
- Sentence case, no trailing period, at most 100 characters.
- Says what changes for a user when it closes. Read it without the body: if the outcome is unclear,
  rewrite it.

| Bad                        | Better                                                       |
| -------------------------- | ------------------------------------------------------------ |
| `+[check]: Add github fmt` | `Add a github output format to rtunk check`                  |
| `Fix cache`                | `Cache shim for a tool is rewritten on every run`            |
| `Improve performance`      | `Reuse parsed config across linters during rtunk check`      |
| `Docs`                     | `Document the .rtunk/ versus .trunk/ precedence in the wiki` |

## 5. Issue Type and labels

Set exactly one Issue Type (`gh issue create --type`; the repo has `Bug`, `Feature`, `Task`), one
`type::*` label, and a `priority::*` label only when the priority is obvious from the evidence
(otherwise leave it for `triage-issues`). Do not create labels; `size::*` does not exist.

| Intent                                          | Issue Type | `type::*` label  |
| ----------------------------------------------- | ---------- | ---------------- |
| Something behaves wrongly                       | `Bug`      | `type::bug`      |
| New capability or command/flag/config           | `Feature`  | `type::feature`  |
| Documentation only                              | `Task`     | `type::docs`     |
| Tooling, CI, refactor, dependency, maintenance  | `Task`     | `type::chore`    |
| Security hardening (not a vulnerability report) | `Task`     | `type::security` |

Ambiguous pairs:

| Pair                | Ask                                                                   | Resolution                                                                 |
| ------------------- | --------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| `Feature` vs `Task` | Could users already do this, just by a worse mechanism?               | Yes: `Task` (only the implementation changes). No: `Feature`.              |
| `Docs` vs `Task`    | Does closing it change anything besides a document?                   | No: `type::docs`. Yes: `type::chore`/`bug`, with the doc as a side effect. |
| `Bug` vs `Feature`  | Did rtunk ever promise this (docs, trunk compat claim, `ROADMAP.md`)? | Yes and it fails: `Bug`. No: `Feature`.                                    |

The issue templates also attach the default `bug`/`enhancement` labels. Keep them as GitHub sets
them; add the `type::*` label next to them. Do not add `status::*` labels at filing time
(`status::in-progress|needs-review|blocked` describe work in flight).

`priority::*` quick guide (full procedure in `triage-issues`): `high` when something users rely on
is broken now or a release/CI path is broken; `medium` for a real, recurring annoyance or a
needed enhancement; `low` for nice-to-have with no stated reason. Unsure: omit.

## 6. Body

Four sections, in this order; drop one only if it truly does not apply. Prose is not wrapped at 80
columns (that limit is for commit messages).

```markdown
## Context

What exists today, what is broken or missing, with evidence: verbatim error text, command and
output, versions, links to runs, PRs, `ROADMAP.md` or doc sections. Facts only.

## Proposal

For work: what to do, with options only when the choice is genuinely open (pick one and say why
otherwise). For a bug: exact steps, observed versus expected (the bug template sections may replace
this when they fit better).

## Acceptance criteria

- [ ] Observable outcomes checkable by running a command or reading a file/CI result

## Notes

Cross-links (issues, PRs, upstream), related `ROADMAP.md` item. Optional.

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

- Acceptance criteria are checkable: "`rtunk check` on the repro config exits 0 and prints no
  warning" passes, "works better" fails.
- No implementation step lists, no empty risk or alternatives tables, no restating the title.
- A focused 25-line body with two links beats a 200-line one that fills every field.
- Evidence must be real: run the command and paste the output; do not reconstruct from memory.

## 7. Create

Write the body to a file in the session scratchpad and use `--body-file`; never inline in the shell.

```bash
gh issue create -R axnic/rtunk \
  --title "<sentence-form title>" \
  --body-file <scratchpad>/issue.md \
  --type Bug \
  --label "bug" --label "type::bug"
```

`--type` needs a recent `gh`; if it is rejected, create without it and set the type with
`gh issue edit <n> -R axnic/rtunk --type Bug`, and if that fails too, say so in the report rather
than skipping silently. The command prints the URL; the number is its last path segment.

## 8. Cross-link

- Deferred from a PR or another issue: mention `#N` in both directions (a comment on the PR is
  enough: `gh pr comment`).
- Part of a larger piece of work: reference the parent in Notes.
- A PR that fully resolves it says `Closes #N` in its `Related` section (see `create-pr`).

## Rules summary

- Duplicate check first; one Issue Type + one `type::*` label; existing labels only.
- Sentence-form title, no commit symbols; body Context / Proposal / Acceptance criteria / Notes.
- Respect the templates' sections and the out-of-scope list; vulnerabilities go to the private
  advisory form.
- Filing is the whole action: do not start the fix from here.

Examples: `references/issue-examples.md`.
