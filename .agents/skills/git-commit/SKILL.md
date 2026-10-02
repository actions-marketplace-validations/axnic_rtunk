---
name: git-commit
description: Use when the user wants to commit, stage, or finalize work in the rtunk repository — creating a commit message, splitting staged changes into atomic commits, or checking a message against the project's commitlint rules.
---

# Git commit convention

## Why we commit this way

A commit is a historical record, not just a sync mechanism to move code from
one place to another. Future readers — human or AI — will use `git log` and
`git blame` to understand _why_ a line looks the way it does, long after the
diff itself stops being interesting. That only works if messages carry intent
instead of restating what the diff already shows.

Prefer small, imperfect, atomic commits over batching everything into one
large commit and squashing later. One logical change per commit: it makes
review, bisection, and revert all cheaper.

## Commit format

```text
type[scope]: Subject starting with uppercase

Body providing context and intent, max 80 chars per line

Assisted-by: <provider>:<model-id>
```

Breaking change variant — the `BREAKING CHANGE:` paragraph is mandatory:

```text
+![check]: Subject starting with uppercase

Body providing context and intent, max 80 chars per line

BREAKING CHANGE: what breaks and what the caller must do about it

Assisted-by: <provider>:<model-id>
```

## Types

| Symbol         | Name       | Use for                                                    |
| -------------- | ---------- | ---------------------------------------------------------- |
| `+`            | Add        | New feature, command, resource                             |
| `-`            | Remove     | Delete code, file, dead feature                            |
| `~`            | Improve    | Perf, config, behavioral improvement (non-bug)             |
| `!`            | Fix        | Repair a bug or broken behavior                            |
| `=`            | Refactor   | No behavior change (style, tests, DX, CI)                  |
| `^`            | Bump       | Dependency version upgrade or downgrade                    |
| `>`            | Move       | Rename or relocate resources                               |
| `<`            | Revert     | Undo a previous commit                                     |
| `@`            | Docs       | README, AGENTS/ROADMAP, comments                           |
| `$`            | Security   | Fix, policy, secret management                             |
| `?`            | Experiment | POC, investigation, research                               |
| `*`            | Wildcard   | Does not fit any other type                                |
| `+!` `~!` `-!` | Breaking   | Only Add, Improve, Remove can break backward compatibility |

Disambiguation rules:

- **`~` vs `!`** — did the previous behavior have a bug? Use `!` if you fixed
  something broken. Use `~` if the previous behavior was correct and you
  changed it to something better (faster, clearer, more configurable).
- **`=` vs `~`** — is the change observable by a user of rtunk (a person
  running the CLI, or reading its config schema)? Observable → `~`. Invisible
  (internal refactor, added tests, CI tweak) → `=`.
- **`^`** is always used for dependency bumps, even when done by hand rather
  than by an automated tool.
- CI/CD is not its own type. Use the type that describes what actually
  changed (usually `=` or `~`) with the `ci` scope.

## Scopes

Scope is mandatory and bracketed: `type[scope]: Subject`. rtunk is a single
Go CLI, not a monorepo, so the scope list stays flat rather than namespaced
like `project:*`/`catalog:*` would be.

| Scope      | Covers                                                                             |
| ---------- | ---------------------------------------------------------------------------------- |
| `config`   | `trunk.yaml`/`rtunk.yaml` parsing, schema, config resolution                       |
| `plugin`   | Plugin definitions, discovery, linter/runtime/tool resolution                      |
| `cache`    | Download, content-addressed cache, shims                                           |
| `check`    | Check command policy: which commands run (`Formatter: false`), read-only reporting |
| `engine`   | Shared job-queue engine: file matching, `RunFrom`/`SandboxType`, execution         |
| `output`   | Linter output-format parsers (SARIF, per-tool JSON schemas, `parse_regex`)         |
| `fmt`      | Formatters command                                                                 |
| `actions`  | Actions and git-hooks                                                              |
| `upgrade`  | Self-upgrade command                                                               |
| `init`     | `init`/`deinit` command                                                            |
| `renovate` | Renovate annotation generation (`rtunk renovate`)                                  |
| `cli`      | Top-level CLI wiring, flag compatibility, entrypoints                              |
| `deps`     | Go module or tool version bumps                                                    |
| `ci`       | `.github/` workflows, `.rtunk/` dogfood config, `mise.toml`                        |
| `docs`     | README, AGENTS.md, ROADMAP.md, ADRs                                                |

Decision tree: which files did the change touch?

1. Only `.github/`, `.rtunk/`, `mise.toml` → `ci`.
2. Only `go.mod`/`go.sum` (or a pinned tool version) with no code change →
   `deps`.
3. Only `*.md` prose (no code) → `docs`.
4. Otherwise, match the package/command area to the table above (e.g.
   `pkg/trunk/config` → `config`, `pkg/trunk/download` → `cache`,
   `pkg/trunk/check` → `check`, `pkg/trunk/engine` (including
   `pkg/trunk/engine/security`) → `engine`, `pkg/trunk/output` → `output`).
5. Ambiguous, or genuinely spans more than one area → ask the user, never
   guess.

Multiple scopes are allowed, comma-separated (`type[scope1,scope2]:`), when
one atomic change truly spans areas — cap at 3. In a single-CLI repo this is
the exception, not the norm: most commits should carry exactly one scope.

## Subject

- Imperative mood ("Add", not "Added" or "Adds").
- Starts with an uppercase letter.
- No trailing period.
- Max 100 characters.

## Body

Optional for trivial changes (a one-line dependency bump, a typo fix),
mandatory otherwise. It must explain **why**, never restate **what** — the
diff already shows what changed.

- Motivation, trade-offs, or user impact — not a narration of the diff.
- Max 80 characters per line, sentence-case.
- Sentence-case applies to the body's very first character, not each line.
  `rtunk`/`trunk` are lowercase by convention, so if the first sentence would
  otherwise start with one of them, rephrase around it instead
  (e.g. "Pins go and trunk because..." not "trunk needs...").
- The "why" must come from the user's own words in conversation. Never infer
  it from the diff. If the user hasn't stated it, ask before writing the
  body.

## AI trailers

AI-assisted commits carry one trailer, `Assisted-by:`, in the footer (after `BREAKING CHANGE:`
when there is one):

```text
Assisted-by: anthropic:claude-sonnet-5.5
```

- `Assisted-by: <provider>:<model-id>` is the repository's disclosure convention: the human
  stays the author, the AI is a tool they direct. Use the model powering the session, with dots
  in version numbers, not hyphens (`claude-sonnet-5.5`, not `claude-sonnet-5-5`).
- Never `Co-authored-by:`/`Co-Authored-By:` for the AI tool: a tool a human directs is not a
  co-author. If the tooling adds that trailer, strip it (amend the message) and keep
  `Assisted-by:`.

Commits made without AI help carry no trailer.

## Signing and DCO: the AI must stay out of this

Commits are signed by the human's own key, configured in their git config (GPG or SSH). The AI's
command is always a plain `git commit`; signing happens on its own.

- Never `-S` (it can pick another identity or key than the human's configured one).
- Never `-s`/`--signoff` or a `Signed-off-by:` trailer: the DCO sign-off is the human's
  attestation that they may submit the change, which an AI cannot give.
- Never `--no-gpg-sign` to get past a signing failure, and never `--no-verify` to skip hooks.
  If signing or a hook fails, stop and tell the user.

## Commitlint rules (canonical reference)

This section mirrors `.commitlintrc.js`. **Keep it in sync whenever that file
changes — this section IS the reference for anyone reading the skill instead
of the config.**

Header pattern: `^(\S+?)\[([^\]]+)\]:\s(.+)$` (breaking: `^([+~-]!)\[([^\]]+)\]:\s(.+)$`).

| Rule                          | Level | Value                                                                                                                                 |
| ----------------------------- | ----- | ------------------------------------------------------------------------------------------------------------------------------------- |
| `header-max-length`           | error | 100                                                                                                                                   |
| `header-full-stop`            | error | never `.`                                                                                                                             |
| `header-trim`                 | error | always                                                                                                                                |
| `header-case`                 | off   | symbols have no case                                                                                                                  |
| `type-empty`                  | error | never empty                                                                                                                           |
| `type-enum`                   | error | see Types table                                                                                                                       |
| `scope-empty`                 | error | never empty                                                                                                                           |
| `scope-case`                  | error | lower-case                                                                                                                            |
| `scope-enum`                  | warn  | see Scopes table (warn only — multi-scope commits won't match a single enum entry)                                                    |
| `subject-empty`               | error | never empty                                                                                                                           |
| `subject-case`                | error | sentence-case                                                                                                                         |
| `subject-full-stop`           | error | never `.`                                                                                                                             |
| `subject-max-length`          | error | 100                                                                                                                                   |
| `body-case`                   | error | sentence-case                                                                                                                         |
| `body-max-line-length`        | error | 80                                                                                                                                    |
| `footer-leading-blank`        | off   | built-in disabled, replaced by `squash-footer-leading-blank`                                                                          |
| `squash-footer-leading-blank` | error | always: blank line before the footer, skipped when the subject ends with `(#N)` (GitHub squash-merge: the body is the PR description) |
| `footer-max-line-length`      | error | 80                                                                                                                                    |
| `signed-off-by`               | off   | no DCO sign-off required (and never added by the AI)                                                                                  |

Prompt configuration (for interactive commit tooling, if wired up later):
`allowBreakingChanges: ["+!", "~!", "-!"]`, `allowCustomScopes: false`,
`allowEmptyScopes: false`, `enableMultipleScopes: true`,
`scopeEnumSeparator: ","`, `skipQuestions: ["body", "footerPrefix", "footer"]`,
`upperCaseSubject: true`, `useCommitSignGPG: true`, `useEmoji: false`.

### Keeping this skill in sync

`.commitlintrc.js` is authoritative; when it changes, update this skill to match:

1. **Read `.commitlintrc.js`** and identify what changed (types, scopes, rules, local plugin
   rules, parser patterns, or prompt config).
2. **Update the matching section** here:
   - New type: Types table and the `type-enum` row's wording.
   - New scope: Scopes table and the decision tree.
   - Rule change: the rules table (including the "If commitlint fails" row for it).
   - Parser or prompt change: the header patterns and the prompt configuration paragraph.
3. **Keep this skill self-contained**: the config's header comment points here as the readable
   reference, so do not turn this section into "go check the config".
4. **Verify**: re-read both files side by side and confirm every rule, level, value, type, scope
   and pattern matches exactly. Commit with `type[docs]:` or `type[ci]:` depending on what changed.

### If commitlint fails

| Rule                          | Likely cause                                              | Fix                                                                        |
| ----------------------------- | --------------------------------------------------------- | -------------------------------------------------------------------------- |
| `type-enum`                   | Symbol missing or misspelled                              | Use one of the 12 base symbols, or a `+!`/`~!`/`-!` breaking variant       |
| `scope-empty`                 | No `[scope]` bracket                                      | Add a bracketed scope from the Scopes table                                |
| `header-max-length`           | Subject too long                                          | Trim to ≤100 chars total, move detail to the body                          |
| `subject-full-stop`           | Trailing period on subject                                | Remove it                                                                  |
| `body-max-line-length`        | Body line >80 chars                                       | Rewrap                                                                     |
| `body-case`                   | Body starts with a lowercase word (often `rtunk`/`trunk`) | Rephrase the opening so the first character is uppercase                   |
| `squash-footer-leading-blank` | No blank line before `BREAKING CHANGE:`/`Assisted-by:`    | Add a blank line before the footer (not enforced on `(#N)` squash commits) |

## Workflow

1. Survey the workspace:
   `git status`, `git diff --cached --name-only`, `git diff --name-only`,
   `git log --oneline --no-merges -10`.
2. If staged changes span multiple scopes and aren't one atomic change,
   split them into separate commits instead of forcing a multi-scope header.
3. Select the type using the disambiguation rules above.
4. Determine the scope (or scopes, comma-separated, capped at 3) using the
   decision tree above.
5. Draft the subject: imperative, uppercase start, no trailing period,
   ≤100 chars.
6. Write the body: ask the user for the "why" if it hasn't already come up
   in conversation. Skip the body only for genuinely trivial changes.
7. Stage the files, then run a plain `git commit` whose message comes from a quoted heredoc
   (`git commit -m "$(cat <<'EOF' ... EOF)"`), so multi-line bodies and trailers stay intact.
   Include the `Assisted-by:` trailer (see "AI trailers") when the AI helped, and strip any
   `Co-Authored-By:` the tooling adds.
8. Never add `-S`, `-s`/`--signoff`, `--no-gpg-sign` or `--no-verify` (see "Signing and DCO").

## Examples

**Good — simple add:**

```text
+[output]: Add SARIF output normalization for gitleaks

Trunk-compatible tooling expects SARIF; without it, downstream
consumers (editors, CI annotators) can't parse gitleaks findings
the same way they parse every other linter's output.

Assisted-by: anthropic:claude-sonnet-5.5
```

**Good — dependency bump, no body needed:**

```text
^[deps]: Bump golang.org/x/tools to v0.27.0
```

**Good — breaking change:**

```text
-![config]: Drop support for trunk.yaml v0.0 schema

v0.0 lacked a version field, which made every later schema
migration ambiguous to detect. Every real-world config already
declares a version, so keeping v0.0 support only hid config bugs.

BREAKING CHANGE: configs without a `version` field are now
rejected at load time instead of falling back to v0.0 defaults.

Assisted-by: anthropic:claude-sonnet-5.5
```

**Bad — no type/scope, restates the diff:**

```text
Updated config.go to add new validation function
```

Missing `type[scope]:` entirely, and the body (if any) would just repeat
what the diff already shows instead of explaining why validation was added.

**Bad — Co-Authored-By for the tool, forbidden signing flags:**

```text
![check]: Fix nil pointer in report renderer

Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>
```

committed with `git commit -s -S`. The AI tool is not a co-author: the trailer must be
`Assisted-by: <provider>:<model-id>` only (and it is missing here), and the command must never
carry `-s` or `-S`.

**Bad — invalid type:**

```text
!![check]: Fix and majorly change the reporting pipeline
```

`!!` isn't a valid symbol — only `+`, `~`, `-` can take the `!` breaking
suffix, and a fix (`!`) can't itself be marked breaking.

## References

- [Trailers for AI-assisted commits — All Things Open](https://allthingsopen.org/articles/ai-assisted-commits)
- [Linux kernel documentation on coding assistants](https://docs.kernel.org/process/ai.html)
