# Release notes examples

Reference for maintainers of the summary step; the model never receives this file (the workflow
sends `SKILL.md` only). Each example shows the commit log the model gets (format of
`buildLlmContext` in `scripts/generate-release-notes.mjs`, trimmed) and the one paragraph it must
return. The workflow puts that paragraph in place of `SUMMARY_PLACEHOLDER`; the changes list,
links and contributors come from the draft. Inputs are abridged from real rtunk releases.

## Example 1: single CI fix (v0.13.2)

### Input

```text
commit: 3572614 ![ci]: Only ask the model for the release notes summary paragraph (#16)
body: The model used to rewrite the whole notes; rejected answers silently fell back to the draft.
author: @xunleii
pr: #16 https://github.com/axnic/rtunk/pull/16
```

### Output

```text
This release only touches the release tooling: the notes workflow now asks the model for the summary paragraph alone, so the changes list and contributors always come from the deterministic draft. There is no change to rtunk itself.
```

## Example 2: breaking release (v0.14.0)

### Input

```text
commit: a1b2c3d ~![cli]: Move Go module path to github.com/axnic/rtunk (#24)
commit: b2c3d4e +[ci]: Add a reusable GitHub Action that runs rtunk check (#20)
commit: c3d4e5f $[ci]: Sign release artifacts and generate SLSA attestations with cosign (#17)
commit: d4e5f6a =[ci]: Run the test job on Linux and macOS (#22)
commit: e5f6a7b @[docs]: Fix statements that contradict the released state (#23)
```

### Output

```text
Breaking: the Go module path is now `github.com/axnic/rtunk`, so update your imports and your `go install` command. This release also adds a reusable GitHub Action that runs `rtunk check`, and its release artifacts are now signed with cosign and carry SLSA attestations. CI now tests on Linux and macOS.
```

## Example 3: new feature and a fix (v0.15.0)

### Input

```text
commit: f6a7b8c +[check]: Add --format github and use it in the action (#27)
body: The action post-processed SARIF to get annotations and a job summary.
commit: a7b8c9d +[engine]: Honor lint.ignore in file selection (#31)
commit: b8c9d0e @[docs]: Run a documentation pass after v0.14.0 (#28)
commit: c9d0e1f =[ci]: Ignore .mise/ local state except the tracked lock directories (#26)
```

### Output

```text
`rtunk check --format github` now emits native GitHub annotations and a job summary, so the GitHub Action no longer post-processes its output. File selection also honors `lint.ignore` from `.trunk/trunk.yaml`, so ignored files stop being linted. The rest is documentation and CI housekeeping.
```

## Example 4: thin log (dependencies only)

### Input

```text
commit: d0e1f2a ^[deps]: Bump golang.org/x/tools to v0.27.0 (#32)
commit: e1f2a3b ^[deps]: Bump actions/checkout to v7.0.1 (#33)
```

### Output

```text
This release only updates dependencies: `golang.org/x/tools` and the `actions/checkout` pin. There are no user-facing changes.
```
