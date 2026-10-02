---
name: auto-file-bug
description: Use when work on one task surfaces a second, unrelated bug in the rtunk repository that is evidenced (not suspected), serious enough that recurrence causes real harm (wrong or silently skipped lint results, data loss in the cache, a broken release or CI path), and out of scope for the current change. File the issue autonomously without asking, instead of writing "this deserves its own issue" and moving on. Delegates the drafting to a fresh general-purpose subagent running the `create-issue` skill so the investigation stays off the main thread. Never for security vulnerabilities.
---

# Auto-file a bug

Work on one task regularly surfaces a second, real problem. `create-issue` already covers "you would
otherwise say this deserves its own issue". This skill is the narrower, higher bar: severe enough
that waiting for the user to notice is not acceptable, not severe enough to derail the current task.
File it, report in one line, and keep going.

## Never auto-file a vulnerability

A suspected security issue goes through the private advisory form (`SECURITY.md`), never a public
issue. Tell the user in your next message instead of dispatching anything.

## When (all must hold)

1. **Evidenced.** A concrete symptom: error string, log line, reproducible condition, file and
   line. "This looks fragile" is not enough.
2. **Real harm on recurrence.** Wrong or skipped `check`/`fmt` results, a broken cache or shim,
   a failing release or CI path, silent data or config misreading. Not "could be nicer".
3. **Out of scope now.** Fixing it would expand or derail the current change. A one-line fix that
   touches nothing else: just fix it.
4. **Nothing left to investigate.** If confirming the root cause needs more digging, surface it to
   the user instead.

Otherwise mention it inline in one line, or fix it.

## Workflow

1. Do not draft the issue in the current context.
2. Dispatch a fresh subagent with the `Agent` tool: `general-purpose` (or `claude`), not `fork` (a
   fork inherits this conversation, which defeats the point).
3. Hand it a self-contained brief (it has no history): exact symptom, root cause with `file:line` if
   known, one sentence on why it is not fixed now, related PR/issue numbers, and a suggested title,
   Issue Type and `type::*` label as hints it may override.
4. Tell it to follow `.agents/skills/create-issue/SKILL.md` (duplicate check, title, Issue Type,
   labels, body, `gh issue create -R axnic/rtunk`).
5. On return, report one line, `Filed issue #N: <title>`, and resume the original task. Do not
   re-explain the bug or ask whether it was wanted.

## Brief template

```text
Read and follow .agents/skills/create-issue/SKILL.md in this repository to file a GitHub issue
(axnic/rtunk) for a bug found while working on <current task/PR>. Do not ask me anything: check for
duplicates, then create it directly. Existing labels only; do not create labels. Do not fix the
bug, and do not file it if it turns out to be a security vulnerability (report that back instead).

Symptom: <exact error text / log excerpt / command and output>
Root cause: <what is wrong, with file:line, or "unknown">
Why not fixed now: <one sentence>
Related: <PR/issue numbers>
Hints (override if wrong): title "<sentence-form title>", Issue Type <Bug|Task>, labels <type::*, priority::* if obvious>

Reply with one line only: the issue number and title, or "duplicate of #N" if one already exists.
```

## Non-goals

- Does not change when to merely propose an issue for lower-severity findings: say it in one line.
- Does not authorize fixing the bug outside the current task; filing is the entire action.
