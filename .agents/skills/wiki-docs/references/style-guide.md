# Style guide — what makes a wiki excellent

Distilled from a study of three GitHub wikis held up as models — [Guava](https://github.com/google/guava/wiki),
[Sniffnet](https://github.com/GyulyVGC/sniffnet/wiki), [Gondwana](https://github.com/Isthimius/Gondwana/wiki)
— and the guides listed in [awesome-github-wiki](https://github.com/MyHoneyBadger/awesome-github-wiki)
(freeCodeCamp, Carl de Souza, Nimble, Sparkbox, Ably, BugHerd) plus GitHub's own wiki docs. Study
done 2026-10-01. Each rule names where it was observed, so it can be weighed instead of obeyed
blindly. Do not re-research these sources to write rtunk docs; this file is the extract.

## Information architecture

1. **Home is a hub, not a page of content.** Home states what the project is in one sentence, then
   routes the reader: a few titled groups, each with a one-line lead and its links. No tutorial
   prose on Home. (Sniffnet: 28 pages in four groups plus FAQ; Gondwana: one-sentence definition,
   then sections; Guava: category blocks.)
2. **Group by reader problem, not by code module.** Section leads say what the reader is trying to
   do ("scenarios where values are expensive to compute", "Set up the analysis"), not which
   package implements it. (Guava, Sniffnet.)
3. **Order groups by the reader's journey.** Getting started → everyday use → reference →
   internals/advanced, with FAQ apart. (Sniffnet's progressive disclosure, Gondwana's six-group
   sidebar from Introduction to Advanced Topics, Nimble's Home/Getting Started/Architecture
   skeleton.)
4. **Shallow hierarchy.** Home → page, at most one grouping level in the sidebar. No intermediate
   "category" pages. (All three wikis.)
5. **One topic per page.** A page answers one question or covers one feature; split rather than
   append. (freeCodeCamp, Sniffnet.)
6. **Names signal page type.** A consistent naming pattern lets a reader predict what a page is
   (Guava's `XxxExplained`, `ReleaseNN`). In rtunk: verb phrases for guides (`Checking-Code`),
   `-Reference` for references, nouns/`-Design` for internals.
7. **A curated sidebar is the main navigation.** GitHub's default page list is flat and
   alphabetical; a hand-written `_Sidebar.md` that groups pages is what makes a wiki navigable.
   (Carl de Souza, Nimble; every studied wiki has one.)

## Page anatomy

8. **Short version first.** Open every page with one or two sentences — or the minimal command —
   that satisfies a reader who reads nothing else; depth follows. (Gondwana's "short version";
   Guava defines the concept in plain terms before any API.)
9. **Motivation before mechanics.** Say why a feature exists, or what the reader gains, before how
   to drive it; state constraints and trade-offs explicitly. (Guava's design-rationale sections,
   Sniffnet's benefit-first intros.)
10. **Example before explanation.** Show a working command or config with real values and its real
    output, then explain it. (Sniffnet shows a full TOML theme before setup steps; Gondwana never
    describes an API without a usage example.)
11. **Tables for anything enumerable.** Flags, config keys, exit codes, is/is-not comparisons, task
    → command lookups go in tables, not prose. (Guava API tables, Gondwana quick-lookup tables,
    Sniffnet shortcut table.)
12. **Admonitions for the exception, not the rule.** `> [!NOTE]`, `> [!TIP]`, `> [!WARNING]` for
    platform caveats, gotchas and format hints, so the main prose stays the happy path. (Sniffnet.)
    Use at most two or three per page.
13. **Diagrams replace prose for flows and structure.** Mermaid renders on both repo and wiki.
    (Gondwana's lifecycle and pipeline diagrams; GitHub docs confirm Mermaid support.)
14. **End with "Where to go next".** Two to four links to the logical next pages. Every page is
    linked from the sidebar and links onward; no dead ends. (Gondwana; Sniffnet's bottom
    navigation.)
15. **Mention related topics inline.** Link the first mention of a concept that has its own page.
    (Guava, Gondwana.)

## Writing

16. **Active voice, present tense, second person** for guides ("run `rtunk check`"), plain
    declarative statements for reference ("`--fix` applies autofixes"). (Guava; Nimble.)
17. **Short sentences; one idea each.** Explain jargon inline at first use. (Guava, Sniffnet.)
18. **Write for a developer who does not know the project.** Assume they know git and a terminal,
    not rtunk or trunk internals. (Nimble.)
19. **State current, verified behavior.** Document what the shipped binary does. Planned work is a
    one-line pointer to ROADMAP.md, never described as if it exists.
20. **Every fact lives in one place.** Guides explain a workflow and link to the reference for the
    full flag/key table; they do not copy it. Copies drift.
21. **No marketing, no emoji, no filler** ("simply", "just", "powerful", "easy"). Match the repo's
    existing tone: dense and precise. Sniffnet's emoji on Home is not carried over.

## Maintenance

22. **Docs are reviewed like code.** The source lives in the main repo under `docs/` and is
    published to the wiki by CI; nobody edits the wiki in the web UI, because the wiki has no pull
    requests. (Sparkbox, Ably, BugHerd, freeCodeCamp.)
23. **Update docs in the same change as the behavior.** A new flag or config key ships with its
    reference entry. (Nimble: write docs when the feature lands.)
