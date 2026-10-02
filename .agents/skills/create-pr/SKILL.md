---
name: create-pr
description: Use when the user wants to open, create or submit a pull request for the current work in the rtunk repository ("open a PR", "push this and create the PR", "make a pull request"). Runs the pre-flight (context, commit validation, `mise run ci`), pushes the branch, writes a PR title that is a valid commit header (squash merge keeps it) and a body that fills `.github/PULL_REQUEST_TEMPLATE.md` without tripping commitlint's footer-trailer parsing, creates it with `gh`, then arms a Monitor on the PR (comments, reviews, CI, merge). `main` is protected, so this is the only way work lands.
---

# Create a pull request

`main` is protected ("Changes must be made through a pull request"): a direct push is rejected and
every change lands through a squash-merged PR. The repo is `axnic/rtunk`; always pass
`-R axnic/rtunk` to `gh` instead of relying on the working directory. Merging is the user's call,
never done here unless explicitly asked.

## 1. Pre-flight

1. Snapshot (branch, push status, existing PR, commits, files, uncommitted changes), and use it for
   the whole draft instead of re-running `git log`/`git diff`:

   ```bash
   bash .agents/skills/create-pr/scripts/gather-context.sh        # base defaults to origin/main
   ```

2. Never push to or open a PR from `main`. If the commits sit on local `main`, push them to a new
   branch (`git push origin main:refs/heads/<branch>`) and leave local `main` for the user to realign.
   Uncommitted changes: stop, commits are made by the `git-commit` skill, not here.
3. Validate the commits (`mise run ci:commitlint -- --from <base> --to HEAD`, signatures,
   `Assisted-by:` trailer):

   ```bash
   bash .agents/skills/create-pr/scripts/validate_commits.sh
   ```

   `FAIL` blocks the PR (fix the commit, see `.agents/skills/git-commit/SKILL.md`). `WARN` is
   surfaced to the user as is.

4. Run the local gate and read the result:

   ```bash
   mise run ci        # lint, build, race tests with the coverage floor, release-script tests
   go vet ./...
   gofmt -l .         # must print nothing
   ```

   Tick a box in the PR body only for a command that actually ran and passed.

5. Style reference: `references/pr-examples.md`; recent merged PRs: `gh pr list -R axnic/rtunk --state merged -L 5`.

## 2. Branch and push

Branch names are `<area>/<short-kebab-description>` (`ci/fix-lint-findings`, `docs/wiki-quickstart`,
`feat/github-output-format`). Never `--force`; if the branch already exists on the remote with
other commits, ask the user.

```bash
git push -u origin HEAD
```

## 3. Title

The squash merge turns the PR title into the commit subject on `main`, so it is a valid commit
header, `type[scope]: Subject`, checked against `.commitlintrc.js`:

- `type`: one of `+ - ~ ! = ^ > < @ $ ? *`, breaking variants `+! ~! -!` only.
- `scope`: mandatory, lower-case, from the `scopes` list in `.commitlintrc.js` (`config`, `plugin`,
  `cache`, `check`, `engine`, `output`, `fmt`, `actions`, `upgrade`, `init`, `renovate`, `cli`,
  `deps`, `ci`, `docs`); several are comma-separated, no space (`engine,check`).
- `Subject`: sentence case, no trailing period, header at most 100 characters.
- One commit: reuse its header. Several: one header that sums them up, with the type of the
  dominant change.

Re-read `.commitlintrc.js` rather than trusting this list if it looks stale.

## 4. Body

Fill `.github/PULL_REQUEST_TEMPLATE.md` section by section (`What this changes and why`,
`Related`, `How this was tested`, `Checklist`); keep its headings.

- Why before what: the motivation, then the behavior change, then known limits.
- `Related`: the `ROADMAP.md` item or issue, with `Closes #N` when the PR fully resolves it.
- `How this was tested`: `[x]` only for what ran and passed; `[ ]` otherwise. Add manual steps
  (commands and observed output) when the change is not covered by the template's commands.
- Checklist: leave unchecked what is not true (signed commits, docs updated).
- **Footer-trailer pitfall.** The body becomes the squash commit's body, and commitlint on `main`
  parses any line shaped `word: text` near the end as a git trailer (`footer-leading-blank`,
  `footer-max-line-length` of 80). Never start a line with `word:` (`Note:`, `Known limit:`,
  `Fixes:`); put the label in bold or inside a sentence, and keep lines that wrap prose from
  starting with `word:`. The `squash-footer-leading-blank` rule exempts subjects ending `(#N)`, but
  a red `main` is exactly what #32 fixed, so do not rely on it.
- End the body with `🤖 Generated with [Claude Code](https://claude.com/claude-code)`.
- Check the file before creating: `grep -nE '^[A-Za-z-]+:( |$)' <body-file>` must print nothing
  (the `Related`-style lines `Closes #N` have no colon and pass).

Write the body to a file in the session scratchpad, never inline in the shell.

```bash
gh pr create -R axnic/rtunk --base main --head <branch> \
  --title "<header>" --body-file <scratchpad>/pr.md
```

The command prints the PR URL; the number is its last path segment. Labels are optional:
use only existing ones (`type::*`, `status::*`, `priority::*`); do not create labels.

## 5. Monitor

Right after the PR is created, arm the watcher with the `Monitor` tool:

```js
Monitor({
  command: "bash .agents/skills/create-pr/scripts/monitor-pr.sh <pr-number> 60",
  description: "PR #<pr-number>: comments, reviews, CI, merge",
  persistent: true,
});
```

`monitor-pr.sh` polls every 60 s, prints one line per event, and exits when the PR leaves `OPEN`:

- `[comment] <user>: <body>`: top-level conversation comment.
- `[review comment] <user> on <path>:<line>: <body>`: inline diff comment.
- `[review <STATE>] <user>: <body>`: submitted review (`APPROVED`, `CHANGES_REQUESTED`, or a
  `COMMENTED` review with a body).
- `[check] <name>: pass|fail|cancel|skipping`: each finished CI check, reported once.
- `[state] PR #<n> is now MERGED|CLOSED`: terminal; the script exits.

### Reacting to events

- `comment`, `review comment`, `review CHANGES_REQUESTED`: read the full thread with `gh`, then
  handle it with the `address-pr-review` skill (verify before changing anything, reply on the
  thread, resolve only once the fix is pushed). Do not reply on the user's behalf unless asked.
- `check <name>: fail`: read the log with `gh run view <run-id> -R axnic/rtunk --log-failed`, find
  the cause and tell the user; fix it only if it belongs to this PR's changes.
- `review APPROVED` or a plain `COMMENTED` with no ask: note it to the user, no action.
- `MERGED`: report it, then offer to `git switch main && git pull` and delete the local and remote
  branch (deletion needs confirmation). `CLOSED`: report it and ask whether to keep the branch.

### Follow-up commits

After pushing a new commit to a branch with an open PR: re-run steps 1.3 and 1.4, then ask the user
whether the PR body (and title) needs updating. Show the proposed diff of the body before applying
it with `gh pr edit`; never rewrite it silently.

## Rules

- Title is a valid commit header; body follows the template and has no `word:` lines.
- No force-push, no merge, no label creation, no PR from `main`.
- Branch names `<area>/<kebab>`; commits signed, `Assisted-by:` trailer when an AI helped (never
  `Co-Authored-By:` for the AI tool: see `CONTRIBUTING.md`).
- Do not wrap body prose to 80 columns: the 80-column limit applies to commit messages, but keep
  a line from starting with `word:`.
