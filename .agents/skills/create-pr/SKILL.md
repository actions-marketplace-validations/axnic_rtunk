---
name: create-pr
description: Use when the user wants to open a pull request for the current work in the rtunk repository — pushing the branch, filling the PR template, creating the PR with gh, and then monitoring it for review comments, CI results and merge. `main` is protected (changes must go through a PR), so this is the only way work lands.
---

# Create a pull request

`main` is protected by a repository rule ("Changes must be made through a pull request"), so a
direct `git push origin main` is rejected. Every change lands through a PR. The repo lives at
`axnic/rtunk`; always pass `-R <owner/repo>` to `gh`
instead of relying on the working directory.

## 1. Preconditions

- Never push to or open a PR from `main`. If the commits sit on local `main`, push them to a new
  branch (`git push origin main:refs/heads/<branch>`) and leave local `main` for the user to realign.
- Every commit follows the commit convention (`type[scope]: Subject`, GPG-signed, `Assisted-by:`
  trailer). Commits are made by the `git-commit` skill / `git-commit-assistant` agent, not here.
- Run the local gate and read the result before opening anything:

  ```bash
  mise run ci        # golangci-lint + build + tests with the coverage floor
  go vet ./...
  gofmt -l .         # must print nothing
  ```

  Tick a box in the PR body only for a command that actually ran and passed.

## 2. Branch and push

Branch names are `<area>/<short-kebab-description>` (`ci/fix-lint-findings`, `docs/wiki-quickstart`).
Never use `--force`; if the branch already exists on the remote, ask the user.

```bash
git push -u origin HEAD          # when already on the feature branch
```

## 3. Write the PR

- **Title**: a valid commit header, `type[scope]: Subject` (CI lints the commits, and the title is
  what a squash merge keeps). One commit: reuse its header. Several: a header that sums them up.
- **Body**: fill `.github/PULL_REQUEST_TEMPLATE.md` section by section — what changes and *why*,
  related `ROADMAP.md` item or issue, how it was tested, checklist. Leave unchecked what was not
  run. End the body with `🤖 Generated with [Claude Code](https://claude.com/claude-code)`.
- Write the body to a file in the session scratchpad, never inline in the shell.

```bash
gh pr create -R axnic/rtunk --base main --head <branch> --title "<header>" --body-file <scratchpad>/pr.md
```

`gh pr create` prints the PR URL; the number is its last path segment. Do not merge the PR — that
is the user's call unless they explicitly ask.

## 4. Monitor the PR

Immediately after creating the PR, arm a `Monitor` (long watch: `timeout_ms: 1800000`, the maximum;
description like `PR #<n>: comments, CI, merge`). Each stdout line is one event. Replace `PR` and
`REPO`:

```bash
PR=<number>; REPO=axnic/rtunk
since=$(date -u +%Y-%m-%dT%H:%M:%SZ); prev=""
while true; do
  now=$(date -u +%Y-%m-%dT%H:%M:%SZ)
  {
    gh api "repos/$REPO/issues/$PR/comments?since=$since" | jq -r \
      '.[] | "comment @\(.user.login): \(.body | gsub("\n";" ") | .[0:300])"'
    gh api "repos/$REPO/pulls/$PR/comments?since=$since" | jq -r \
      '.[] | "review-comment @\(.user.login) \(.path):\(.line // .original_line): \(.body | gsub("\n";" ") | .[0:300])"'
    gh api "repos/$REPO/pulls/$PR/reviews" | jq -r --arg s "$since" \
      '.[] | select(.submitted_at > $s) | "review \(.state) @\(.user.login): \(.body | gsub("\n";" ") | .[0:300])"'
  } 2>/dev/null || true
  cur=$(gh pr checks "$PR" -R "$REPO" --json name,bucket \
    --jq '.[] | select(.bucket!="pending") | "check \(.name): \(.bucket)"' 2>/dev/null | sort)
  comm -13 <(echo "$prev") <(echo "$cur")
  prev=$cur
  case "$(gh pr view "$PR" -R "$REPO" --json state -q .state 2>/dev/null)" in
    MERGED) echo "PR #$PR MERGED"; exit 0 ;;
    CLOSED) echo "PR #$PR CLOSED without merge"; exit 0 ;;
  esac
  since=$now
  sleep 30
done
```

It covers every terminal state: new comments, review comments and reviews, each check result
(`pass`, `fail`, `cancel`, `skipping`), and the PR being merged or closed (the loop then exits).

### Reacting to events

- `comment` / `review-comment` / `review CHANGES_REQUESTED`: read the full thread with `gh`, then
  follow the `superpowers:receiving-code-review` skill — verify before changing anything, and do
  not reply on the user's behalf unless asked. Surface the comment to the user.
- `check <name>: fail`: read the log with `gh run view <run-id> -R <repo> --log-failed`, find the
  cause, and tell the user; fix it only if it belongs to this PR's changes.
- `PR #<n> MERGED`: report it, then offer to `git switch main && git pull` and delete the local
  and remote branch. `CLOSED`: report it and stop.
- Monitor expired with the PR still open: re-arm it with the same script (`since` restarts at the
  expiry time, so nothing already reported is replayed).
