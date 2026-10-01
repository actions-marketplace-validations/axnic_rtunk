---
name: wiki-docs
description: Use when writing, restructuring, reviewing or regenerating rtunk's documentation in docs/ (the source of the GitHub Wiki) — documenting a new command, flag or config key, adding or splitting a page, fixing navigation, or rebuilding the whole docs set from scratch.
---

# Wiki documentation

`docs/` is the single source of rtunk's GitHub Wiki: plain Markdown that reads correctly in the
repository browser and is published unchanged in meaning to the wiki by CI. This skill holds
everything needed to write it — the distilled rules ([references/style-guide.md](references/style-guide.md)),
the page skeletons ([references/page-templates.md](references/page-templates.md)), and the page
inventory below. Do not re-study the reference wikis; the style guide is their extract.

## Hard constraints (wiki rendering)

These come from how GitHub Wiki and the sync action (`Andrew-Chen-Wang/github-wiki-action`,
`path: docs`, preprocessing on — see `.github/workflows/push,workflow_dispatch.wiki.yaml`, which publishes on every push to
`main` touching `docs/`) behave. Breaking one produces a broken page on the wiki even when
the repo view looks fine.

- **Flat directory.** Every page is a file directly in `docs/`; no subdirectories. Wiki page names
  are a flat namespace, and subfolder behavior is not reliable.
- **Filename = page title.** `Title-Case-With-Hyphens.md`; hyphens become spaces on the wiki
  (`Checking-Code.md` → "Checking Code"). Names are unique and stable — renaming breaks inbound
  wiki links.
- **`docs/README.md` is Home.** The action renames it to `Home.md`. `_Sidebar.md` and `_Footer.md`
  sit in `docs/` and are wiki chrome, not pages.
- **One `# Title` per page**, matching the filename with spaces. Heading anchors are GitHub slugs
  (lowercase, spaces → `-`, punctuation dropped); keep headings unique within a page.
- **Links between pages:** relative with extension, `[Checking code](Checking-Code.md#exit-codes)`.
  Never `[[Wiki Links]]` (the repo view does not render them).
- **Links to repo files:** relative from `docs/`, `[AGENTS.md](../AGENTS.md)`,
  `[engine](../pkg/run/engine)`; the action rewrites them to `blob/` URLs.
- **Images** in `docs/images/`, referenced as `images/name.svg`. Prefer Mermaid over images.
- **No TOC markup** (the wiki has none); a quick-lookup table at the top of long references
  replaces it.

## Page inventory

Every page, its sidebar group, its type (template in references/page-templates.md), and where its
facts come from. The **sources of truth** win over any existing prose — prose drifts, the binary
does not.

| File                          | Group           | Type             | Sources of truth                                                                |
| ----------------------------- | --------------- | ---------------- | ------------------------------------------------------------------------------- |
| `README.md` (Home)            | —               | Home             | root `README.md`, `AGENTS.md`                                                   |
| `Installation.md`             | Getting started | Guide            | `go.mod`, `cmd/rtunk`                                                           |
| `Quickstart.md`               | Getting started | Guide (tutorial) | `rtunk init`, `linters enable`, `check` run in a scratch repo                   |
| `Migrating-From-Trunk.md`     | Getting started | Guide            | `CLI-Design.md`, `Decision-Log.md`                                              |
| `Checking-Code.md`            | Using rtunk     | Guide            | `rtunk check run --help`, `CLI-Design.md` (file selection, exit codes, `--fix`) |
| `Formatting-Code.md`          | Using rtunk     | Guide            | `rtunk fmt --help`, `CLI-Design.md` (`fmt`)                                     |
| `Ignoring-Issues.md`          | Using rtunk     | Guide            | `pkg/ignore` and its tests                                                      |
| `Managing-Linters.md`         | Using rtunk     | Guide            | `rtunk linters`, `plugins print`, `config print` help                           |
| `Actions-And-Git-Hooks.md`    | Using rtunk     | Guide            | `rtunk actions`, `run`, `git-hooks` help                                        |
| `Keeping-Tools-Up-To-Date.md` | Using rtunk     | Guide            | `rtunk renovate --help`, `CLI-Design.md` (Renovate)                             |
| `Cache-And-Logs.md`           | Using rtunk     | Guide            | `rtunk cache`, `logs`, `toolbox link` help                                      |
| `Command-Reference.md`        | Reference       | Reference        | `rtunk help --all`, `rtunk <cmd> --help`                                        |
| `Configuration-Reference.md`  | Reference       | Reference        | `pkg/trunk/config` types and loader                                             |
| `FAQ.md`                      | Reference       | FAQ              | the pages above; recurring issues                                               |
| `Architecture.md`             | Internals       | Internals        | `pkg/`, `internal/` layout                                                      |
| `Plugin-Model.md`             | Internals       | Internals        | `pkg/trunk/config`                                                              |
| `Plugin-Sources.md`           | Internals       | Internals        | plugin source resolution code                                                   |
| `Cache-Architecture.md`       | Internals       | Internals        | `pkg/cache`                                                                     |
| `Run-Flows.md`                | Internals       | Internals        | `pkg/run/engine`                                                                |
| `CLI-Design.md`               | Internals       | Design spec      | `internal/cli`, `ROADMAP.md`                                                    |
| `Terminal-UX-Design.md`       | Internals       | Design spec      | `internal/cli/render`                                                           |
| `Decision-Log.md`             | Internals       | Decision log     | maintainers' decisions (append-only)                                            |

Audience split: Getting started, Using rtunk and Reference are for **users** (no package names, no
implementation status). Internals is for **contributors** and may cite code; `CLI-Design.md`,
`Terminal-UX-Design.md` and `Decision-Log.md` are authoritative design records referenced from
`AGENTS.md` and code comments — edit their content only to reflect a decided change, never to
"tidy" a decision away.

## Writing rules (summary)

The full list, with the reasoning and where each rule was observed, is in the style guide. The
ones most often broken:

1. Short version first: the opening lines alone must serve a reader who stops there.
2. Example before explanation: real command, real output (run the binary; do not invent output).
3. One fact, one place: guides link to `Command-Reference.md` / `Configuration-Reference.md` for
   full flag and key tables instead of copying them.
4. Document shipped behavior only; planned work is a one-line pointer to `ROADMAP.md`.
5. Tables for anything enumerable; admonitions (`> [!NOTE]`, `> [!TIP]`, `> [!WARNING]`) only for
   exceptions, at most three per page.
6. Every page ends with `## Where to go next` (2–4 links) and appears in `_Sidebar.md`.
7. Active voice, present tense, no filler words or emoji; prose wrapped at 100 columns.

## Workflows

**Documenting a behavior change** (new flag, key, command): update the reference entry first, then
any guide whose workflow changed, then FAQ if it answered the old behavior. Same commit as the code.

**Adding a page:** pick the type, copy its template, name the file per the constraints, add it to
the inventory above, to `_Sidebar.md`, and to the relevant group on Home; link it from at least one
existing page's "Where to go next".

**Regenerating the docs set from scratch:** for each inventory row, rebuild the page from its
sources of truth using its template; build `README.md`, `_Sidebar.md`, `_Footer.md` last from the
inventory. Pages are independent, so they can be written in parallel, one writer per group.

## Verification

Before calling docs work done:

- Every command and output block was produced by the current binary (`go build ./cmd/rtunk`).
- Every relative link and `#anchor` in `docs/` resolves (check script below).
- Every inventory page is in `_Sidebar.md`; no file in `docs/` is missing from the inventory.
- Markdown passes the repo's formatters (`prettier`, `markdownlint` via `rtunk fmt` / `rtunk check`).

````bash
# relative-link + anchor check for docs/
python3 - <<'EOF'
import re, pathlib
docs = pathlib.Path("docs")
slug = lambda h: re.sub(r"[^\w\- ]", "", h.strip().lower()).replace(" ", "-")
strip = lambda p: re.sub(r"```.*?```", "", p.read_text(), flags=re.S)
anchors = {p.name: {slug(m) for m in re.findall(r"^#+ (.+)$", strip(p), re.M)} for p in docs.glob("*.md")}
bad = 0
for p in docs.glob("*.md"):
    text = strip(p)
    for target in re.findall(r"\]\(([^)\s]+)\)", text):
        if re.match(r"[a-z]+:", target): continue
        path, _, frag = target.partition("#")
        f = p.name if not path else path
        if path and not (docs / path).exists(): print(f"{p.name}: missing {target}"); bad += 1
        elif frag and f in anchors and frag not in anchors[f]: print(f"{p.name}: bad anchor {target}"); bad += 1
print("broken:", bad)
EOF
````
