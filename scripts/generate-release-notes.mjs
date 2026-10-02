#!/usr/bin/env node
/**
 * Deterministic release notes generator.
 *
 * Reads the structured commit log written by `git log` and the optional PR metadata written by the
 * release workflow's github-script step, and produces the full Markdown release notes in the same
 * layout as the LLM-polished version -- minus the written summary paragraph, which is left as a
 * placeholder for the model to fill (see .agents/skills/release-notes/SKILL.md). It is also the
 * fallback used verbatim when no model is available.
 *
 * Commits follow rtunk's symbol convention (`type[scope]: Subject`, see
 * .agents/skills/git-commit/SKILL.md and .commitlintrc.js).
 *
 * Usage:
 *   node scripts/generate-release-notes.mjs \
 *     --version 0.13.1 \
 *     --commits-file git-commits.txt \
 *     --prs-file pr-data.json \
 *     --repo axnic/rtunk --prev-tag v0.13.0 \
 *     --output release-notes-draft.md \
 *     --context-output llm-context.md
 *
 * --prs-file, --repo and --prev-tag are optional; without them the PR links, the author handles
 * and the "Full Changelog" link are left out. --context-output additionally writes the commit and
 * PR details the release workflow hands to the model next to the draft.
 */

import { readFileSync, writeFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { parseArgs } from "node:util";

export const SUMMARY_PLACEHOLDER =
  "<!-- SUMMARY_PLACEHOLDER: Replace this comment with a concise 2-4 sentence summary of the most important user-facing changes. -->";

/** Commit-type symbol -> marker shown in the changelog. */
const PREFIXES = {
  "+": "✦", // add
  "!": "✔", // fix
  "~": "⇧", // improve
  "=": "↻", // refactor
  "^": "⚙", // bump
  "-": "✖", // remove
  ">": "➜", // move
  "<": "↩", // revert
  "@": "¶", // docs
  $: "⛨", // security
  "?": "⚗", // experiment
  "*": "✱", // wildcard
};

/**
 * Parses the git log output (records separated by 0x1E, fields by 0x1F) produced by
 * `git log --format="%H%x1f%s%x1f%b%x1f%an%x1f%ae%x1e"`.
 */
export function parseCommits(raw) {
  return raw
    .split("\x1e")
    .filter((r) => r.trim())
    .map((r) => {
      const [sha = "", subject = "", body = "", authorName = "", authorEmail = ""] = r.split("\x1f");
      return {
        sha: sha.trim(),
        subject: subject.trim(),
        body: body.trim(),
        authorName: authorName.trim(),
        authorEmail: authorEmail.trim(),
      };
    })
    .filter((c) => c.sha.length === 40);
}

/**
 * Parses an rtunk commit subject, `type[scope]: Subject`, where type is one symbol (`+`, `!`, ...),
 * optionally followed by `!` for a breaking change (`+!`, `~!`, `-!`). The scope may list several
 * comma-separated scopes. A trailing ` (#123)` (GitHub's merge-commit suffix) is dropped, since the
 * PR link is rendered separately. Returns null when the subject does not follow the convention.
 */
export function parseSubject(subject) {
  const m = subject.match(/^([+\-~!=^><@$?*])(!?)\[([^\]]+)\]:\s*(.+?)(?:\s+\(#\d+\))?$/);
  if (!m) return null;
  return { type: m[1], breaking: m[2] === "!", scope: m[3], description: m[4] };
}

export function isBot(login) {
  return /\[bot\]$/i.test(login ?? "");
}

/** Builds one bullet of the Changes section; the description is always a code span. */
export function buildChangeLine({ prefix, scope, description, breaking }, pr) {
  // A description holding backticks needs a double-backtick span, padded with a space (CommonMark
  // strips it) so a backtick at either end can't merge with the delimiter.
  const [open, close] = description.includes("`") ? ["`` ", " ``"] : ["`", "`"];
  const scopePart = scope ? ` ❲${scope}❳:` : "";
  const flag = breaking ? " ⚠ BREAKING" : "";
  const prPart = pr
    ? ` ([#${pr.number}](${pr.url})${pr.login && !isBot(pr.login) ? ` by [@${pr.login}](https://github.com/${pr.login})` : ""})`
    : "";
  return `- ${open}${prefix}${flag}${scopePart} ${description}${close}${prPart}`;
}

function buildContributorsSection(contributors) {
  const lines = contributors.map((c) => {
    const handle = c.login ? `[@${c.login}](https://github.com/${c.login})` : c.name;
    const prLinks = c.prs.map((p) => `[#${p.number}](${p.url})`).join(", ");
    return `- ${handle}${prLinks ? ` (${prLinks})` : ""}`;
  });
  return ["### ◈ Contributors", "", "Thanks to all the contributors to this release:", "", ...lines].join("\n");
}

/**
 * Generates the deterministic release notes.
 *
 * @param {string} version Version without the leading "v"
 * @param {ReturnType<typeof parseCommits>} commits
 * @param {Map<string, {number: number, url: string, login: string|null}>} prBySha
 * @param {{repo?: string, prevTag?: string}} [opts]
 */
export function generateReleaseNotes(version, commits, prBySha, opts = {}) {
  if (commits.length === 0) return `## What's new in v${version}\n\nNo changes recorded.\n`;

  const changeLines = commits.map((c) => {
    const parsed = parseSubject(c.subject);
    const pr = prBySha.get(c.sha) ?? null;
    return buildChangeLine(
      parsed
        ? { prefix: PREFIXES[parsed.type], ...parsed }
        : { prefix: PREFIXES["*"], scope: null, description: c.subject, breaking: false },
      pr,
    );
  });

  // Contributors, keyed by GitHub login when known, else by git author name. Bots are not listed.
  const contribMap = new Map();
  for (const c of commits) {
    const pr = prBySha.get(c.sha) ?? null;
    if (isBot(pr?.login)) continue;
    const key = pr?.login ?? c.authorName;
    if (!contribMap.has(key)) contribMap.set(key, { login: pr?.login ?? null, name: c.authorName, prs: [] });
    const entry = contribMap.get(key);
    if (pr && !entry.prs.some((p) => p.number === pr.number)) entry.prs.push({ number: pr.number, url: pr.url });
  }

  const sections = [`## What's new in v${version}`, "", SUMMARY_PLACEHOLDER, "", "### ▸ Changes", "", ...changeLines];
  if (contribMap.size > 0) sections.push("", buildContributorsSection([...contribMap.values()]));
  if (opts.repo && opts.prevTag) {
    sections.push(
      "",
      `**Full Changelog**: https://github.com/${opts.repo}/compare/${opts.prevTag}...v${version}`,
    );
  }
  return sections.join("\n") + "\n";
}

/**
 * Renders the commit log, with the pull request that introduced each commit, as the plain-text
 * context handed to the model. Bodies are truncated: the model only needs the gist.
 */
export function buildLlmContext(commits, prBySha) {
  const clip = (text, max) => (text.length > max ? `${text.slice(0, max)}...` : text);
  return commits
    .map((c) => {
      const pr = prBySha.get(c.sha);
      return [
        `commit: ${c.sha.slice(0, 7)} ${c.subject}`,
        c.body && `body: ${clip(c.body, 600)}`,
        `author: ${pr?.login ? `@${pr.login}` : c.authorName}`,
        pr ? `pr: #${pr.number} ${pr.url}` : "pr: (none)",
        pr?.title && `pr title: ${pr.title}`,
        pr?.body && `pr description: ${clip(pr.body, 1200)}`,
      ]
        .filter(Boolean)
        .join("\n");
    })
    .join("\n---\n");
}

// ── CLI (only when executed directly, not when imported by the tests) ───────────────────────────

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const { values: args } = parseArgs({
    options: {
      version: { type: "string" },
      "commits-file": { type: "string" },
      "prs-file": { type: "string" },
      repo: { type: "string" },
      "prev-tag": { type: "string" },
      output: { type: "string" },
      "context-output": { type: "string" },
    },
  });
  if (!args.version || !args["commits-file"] || !args.output) {
    console.error(
      "Usage: node generate-release-notes.mjs --version <v> --commits-file <path> --output <path> [--prs-file <path>] [--repo <owner/repo> --prev-tag <tag>]",
    );
    process.exit(1);
  }

  const commits = parseCommits(readFileSync(args["commits-file"], "utf8"));
  const prBySha = new Map();
  if (args["prs-file"]) {
    try {
      for (const [sha, pr] of Object.entries(JSON.parse(readFileSync(args["prs-file"], "utf8")))) {
        if (pr) prBySha.set(sha, pr);
      }
    } catch {
      // PR data is best-effort: a missing or unparsable file only drops the PR links.
    }
  }

  writeFileSync(
    args.output,
    generateReleaseNotes(args.version, commits, prBySha, { repo: args.repo, prevTag: args["prev-tag"] }),
  );
  if (args["context-output"]) writeFileSync(args["context-output"], buildLlmContext(commits, prBySha));
  console.log(`Deterministic release notes written to ${args.output} (${commits.length} commit(s))`);
}
