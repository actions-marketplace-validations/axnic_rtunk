# PR examples

Real merged PRs of `axnic/rtunk`; read one with `gh pr view <n> -R axnic/rtunk`. Titles are commit
headers because the squash merge keeps them.

## Small fix with a real root cause: #32

Title: `![ci]: Skip footer-leading-blank on GitHub squash-merge commits` (branch `ci/commitlint-squash-footer`)

Why it is a good model:

- `What this changes and why` opens with the failure (red `main` CI after #31, a wrapped line
  starting with `unimplemented:` parsed as a trailer), then the change (a local commitlint plugin
  rule), then a `Known limit`.
- `Related` points at the failing run, not just an issue number.
- `How this was tested` lists three manual scenarios, each with its expected outcome (passes
  before/after, still fails when it must).

Excerpt:

```markdown
## What this changes and why

The push-to-main commitlint step lints the last commit only. The squash commit of #31 embeds the PR
description, and a wrapped line starting with `unimplemented:` was parsed as a git trailer, so
`footer-leading-blank` failed and the CI of `main` went red.

## Related

Red `main` CI after #31 (run 37010219336).

## How this was tested

Manual steps, with `commitlint --config .commitlintrc.js`:

- [x] The squash message of #31 now passes (it failed before)
- [x] A direct commit with a footer glued to the body still fails on `squash-footer-leading-blank`
```

The `Known limit:` wording in the merged text is the pattern to avoid: write `**Known limit.**`.

## Feature with a design rationale: #27

Title: `+[check]: Add --format github and use it in the action` (branch `feat/github-output-format`)

Why it is a good model:

- Bold lead-ins (`**New format github**`, `**Name.**`, `**No auto-detection.**`) structure a long
  body without extra headings, and none of them is a `word:` line.
- Every non-obvious decision (name, no auto-detection, version handling) has its reason next to it,
  including the regression it causes and how it is flagged.
- `How this was tested` ticks the template commands, adds `actionlint`, names the unit tests, then
  shows a real `console` transcript of the end-to-end run.

## Docs-only pass: #30

Title: `@[docs]: Align docs with the github format and the AI trailer rule`

Why it is a good model:

- States up front that no Go code is touched, then one bullet per file with what was wrong and what
  it says now.
- Names what was checked and found already consistent (so a reviewer does not re-check it).
- Notes the verification used (the link/anchor check script of the `wiki-docs` skill, 0 broken).

## Title cheat sheet

| Bad                                | Why                                   | Good                                                     |
| ---------------------------------- | ------------------------------------- | -------------------------------------------------------- |
| `Add --format github`              | no type or scope, commitlint fails    | `+[check]: Add --format github and use it in the action` |
| `+[Check]: Add --format github`    | scope must be lower-case              | `+[check]: Add --format github`                          |
| `+[check]: add --format github.`   | subject case and trailing period      | `+[check]: Add --format github`                          |
| `feat(check): Add --format github` | Conventional Commits, not this repo's | `+[check]: Add --format github`                          |
| `![ci]: Fix`                       | says nothing                          | `![ci]: Skip footer-leading-blank on squash commits`     |
