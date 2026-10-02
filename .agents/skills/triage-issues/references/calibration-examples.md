# Calibration examples

Five reference cases for the priority procedure, built from rtunk's own history and roadmap.
C1 and C2 come from real repository history; C3 to C5 are constructed scenarios. They are proposed calibrations, not maintainer-rated data (the repo has no rated backlog yet):
when the maintainer overrides one, edit it here so the next pass inherits the correction.

## C1: Broken now, relied on: `high`

Case: `main` CI is red after a squash merge because commitlint parsed a wrapped body line
(`unimplemented: ...`) as a trailer and failed `footer-leading-blank` (the situation fixed by #32).

- Rung 2: currently failing, and every later PR and release depends on a green `main`.
- Effort and danger do not enter: the fix was a small local commitlint rule, still `high`.
- Type: `Bug` / `type::bug`.

## C2: Large feature, no current reason: `low`

Case: "Add `rtunk.lock` to verify downloaded tools" (`ROADMAP.md` v1.1, post-v1.0).

- Rung 6: scheduled for a later milestone, nothing blocked on it, no user asking. The trust-on-first-use
  gap is real (`AGENTS.md`), but that is why it is on the roadmap, not why it moves ahead of
  current work.
- It moves to `medium` when v1.0 ships and work on it starts, or immediately if a downloaded-tool
  compromise is reported (that report goes through the private advisory first).
- Type: `Feature` / `type::feature`.

## C3: Broken, nobody uses it: `low`

Case: a documented flag that errors, but the only mention is a deprecated section nobody has
referenced since a command was renamed.

- Rung 2 exception: not relied on by anyone. Propose removing the flag and its doc instead of fixing it.
- Contrast with C1: same kind of breakage, opposite priority, decided only by who is affected today.
- Type: `Bug` if the fix is the proposal, `Task` / `type::chore` if the removal is.

## C4: Security gap, priority depends on the path: `medium` versus `high`

Case A: a diagnostic log line prints an absolute home path. A gap, no exploitation, off the
download/verify/release path: `medium` (rung 3). Type `Task` / `type::security`.

Case B: the release workflow signs artifacts with a step that can be skipped silently on failure
(`SECURITY.md` promises cosign signatures and SLSA attestations): same "gap, nothing exploiting it",
but on the release pipeline, so `high`. Type `Bug` / `type::security` (a flaw being fixed).

If either report describes an exploitable vulnerability rather than a gap, it leaves the public
tracker (rung 1).

## C5: Docs: `medium` versus `low`

Case A: `docs/Command-Reference.md` documents a flag default that the code does not implement
(checked against the flag definition). Users act on it: `medium`. Type `Task` / `type::docs`.

Case B: heading wording, a missing Oxford comma, an inconsistent table header: `low`.

Always verify the discrepancy in the code before rating a docs issue `medium`; an unverified claim
is `low` until checked.
