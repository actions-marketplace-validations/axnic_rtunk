// Run with `node --test scripts/` (or `mise run test:scripts`); no dependency beyond Node.
import assert from "node:assert/strict";
import { test } from "node:test";

import {
  SUMMARY_PLACEHOLDER,
  buildChangeLine,
  buildLlmContext,
  generateReleaseNotes,
  isBot,
  parseCommits,
  parseSubject,
} from "./generate-release-notes.mjs";

const sha = (n) => String(n).padStart(40, "0");
const commit = (n, subject, author = "Alexandre") => ({
  sha: sha(n),
  subject,
  body: "",
  authorName: author,
  authorEmail: "",
});

test("parseSubject reads symbol, breaking marker, scopes and drops the PR suffix", () => {
  assert.deepEqual(parseSubject("+[ci]: Add the thing"), {
    type: "+",
    breaking: false,
    scope: "ci",
    description: "Add the thing",
  });
  assert.deepEqual(parseSubject("-![config,cli]: Remove the flag (#12)"), {
    type: "-",
    breaking: true,
    scope: "config,cli",
    description: "Remove the flag",
  });
  assert.equal(parseSubject("Merge pull request #3 from x/y"), null);
  assert.equal(parseSubject("feat(ci): conventional commits are not the rtunk convention"), null);
});

test("parseCommits splits records and fields and drops malformed records", () => {
  const raw = `${sha(1)}\x1f+[ci]: A\x1f\x1fAlexandre\x1fa@b.c\x1e${sha(2)}\x1f!\x1e` + `short\x1fx\x1e`;
  const commits = parseCommits(raw);
  assert.equal(commits.length, 2);
  assert.equal(commits[0].subject, "+[ci]: A");
  assert.equal(commits[0].authorEmail, "a@b.c");
});

test("buildChangeLine uses double backticks when the description has backticks", () => {
  const line = buildChangeLine(
    { prefix: "✔", scope: "cli", description: "Fix `rtunk init`", breaking: false },
    { number: 7, url: "https://x/pull/7", login: "alice" },
  );
  assert.equal(line, "- `` ✔ ❲cli❳: Fix `rtunk init` `` ([#7](https://x/pull/7) by [@alice](https://github.com/alice))");
});

test("generateReleaseNotes: sections, breaking flag, PR links, bots left out of contributors", () => {
  const commits = [
    commit(1, "+![cli]: Drop the old flag (#10)"),
    commit(2, "^[deps]: Bump github.com/x/y from 1 to 2", "Alexandre"),
    commit(3, "Merge branch 'main'"),
  ];
  const prs = new Map([
    [sha(1), { number: 10, url: "https://x/pull/10", login: "alice" }],
    [sha(2), { number: 11, url: "https://x/pull/11", login: "dependabot[bot]" }],
  ]);
  const out = generateReleaseNotes("0.13.1", commits, prs, { repo: "axnic/rtunk", prevTag: "v0.13.0" });

  assert.match(out, /^## What's new in v0\.13\.1/);
  assert.ok(out.includes(SUMMARY_PLACEHOLDER));
  assert.ok(out.includes("`✦ ⚠ BREAKING ❲cli❳: Drop the old flag` ([#10](https://x/pull/10) by [@alice]"));
  assert.ok(out.includes("`⚙ ❲deps❳: Bump github.com/x/y from 1 to 2` ([#11](https://x/pull/11))")); // no "by @bot" link
  assert.ok(out.includes("`✱ Merge branch 'main'`")); // non-conforming subject falls back to the wildcard marker
  assert.ok(out.includes("- [@alice](https://github.com/alice) ([#10](https://x/pull/10))"));
  assert.ok(!out.includes("dependabot[bot](")); // bots are not listed as contributors
  assert.ok(out.includes("compare/v0.13.0...v0.13.1"));
  assert.ok(isBot("dependabot[bot]") && !isBot("alice") && !isBot(null));
});

test("generateReleaseNotes with no commit says so", () => {
  assert.equal(generateReleaseNotes("1.0.0", [], new Map()), "## What's new in v1.0.0\n\nNo changes recorded.\n");
});

test("buildLlmContext joins commit and PR details and truncates long bodies", () => {
  const commits = [{ ...commit(1, "+[ci]: A"), body: "x".repeat(700) }, commit(2, "!?[bad]")];
  const prs = new Map([[sha(1), { number: 4, url: "https://x/pull/4", login: "bob", title: "Add A", body: "why" }]]);
  const out = buildLlmContext(commits, prs);
  assert.ok(out.includes("commit: 0000000 +[ci]: A"));
  assert.ok(out.includes(`body: ${"x".repeat(600)}...`));
  assert.ok(out.includes("author: @bob\npr: #4 https://x/pull/4\npr title: Add A\npr description: why"));
  assert.ok(out.includes("\n---\ncommit: 0000000 !?[bad]") && out.endsWith("pr: (none)"));
});
