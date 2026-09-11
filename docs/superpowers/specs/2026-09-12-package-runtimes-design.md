# Package-based fetch for python, go, ruby, rust, php

Design for extending `pkg/trunk/download`'s `InstallPackage` (today: `node` only, per
`docs/superpowers/specs/2026-09-10-v0.2-download-design.md`'s non-goals) to the remaining
runtime types actually used by `runtime:`+`package:` tools in the real trunk-io plugin ecosystem,
except `java` (explicitly excluded by user request -- 0 tools in the current plugin snapshot use
`java` this way, so there is nothing to validate against yet).

## Goals

- `InstallPackage` (`pkg/trunk/download/runtime.go`) dispatches to a working implementation for
  `python`, `go`, `ruby`, `rust`, and `php`, one new file per runtime (`runtime_python.go`,
  `runtime_go.go`, `runtime_ruby.go`, `runtime_rust.go`, `runtime_php.go`), mirroring
  `runtime_node.go`'s existing shape exactly: locate the runtime's own package manager binary,
  install into a scratch temp dir, `finalizeInstall` (Fix 3's atomic rename) into place.
- Reuse `finalizeInstall`, `shimSearchPaths`/`FindShimTarget`, and `BuildEnv` unchanged wherever
  possible -- this is an extension of an established, already-reviewed pattern, not new
  architecture.
- Real counts from the cached trunk-io plugin data (162 tools total): `python` 28 tools,
  `go` 13, `ruby` 4, `php` 4, `rust` 1 (`ripgrep`). `java` 0.

## Non-goals (this batch)

- `java` (explicitly excluded by user request).
- cgo-dependent Go tools (anything requiring a C compiler on PATH during `go install`) -- pure-Go
  module installs only, matching what `go install pkg@version` supports without extra toolchain
  setup.
- Composer dependency resolution beyond a single pinned package (no lockfile merging, no
  project-level `composer.json` semantics) -- one tool, one version, in isolation, same scope as
  every other runtime here.
- Any change to `pkg/trunk/config` or the plugin schema itself -- this only teaches
  `InstallPackage` to handle types the schema already declares.

## Per-runtime install mechanics

Every runtime except `php` bundles its own package manager inside the runtime's own downloaded
install dir, exactly like `node` bundles `npm` -- confirmed against the user's real cached
installs (`node`: `bin/npm`; `python`: `bin/pip`). The install command is chosen, per runtime, to
land its output directly under `<pkgInstallDir>/bin/<name>` -- the search path
`shimSearchPaths` already checks -- so **no change to `shimSearchPaths` is needed for
python/go/ruby/rust**, only for `php` (see below).

| Runtime | Binary (in `runtimeInstallDir`)      | Install command                                                                                                      | Output lands at                                          |
| ------- | ------------------------------------ | -------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------- |
| python  | `bin/pip`                            | `pip install --prefix <tmp> pkg==version`                                                                            | `<tmp>/bin/<name>` (POSIX `--prefix` scheme)             |
| go      | `bin/go`                             | `go install pkg@version` with `GOBIN=<tmp>/bin`, `GOPATH=<tmp>/gopath`, `GOCACHE=<tmp>/gocache`, `GOTOOLCHAIN=local` | `<tmp>/bin/<name>` (`GOBIN`)                             |
| ruby    | `bin/gem`                            | `gem install --no-document --install-dir <tmp> --bindir <tmp>/bin pkg -v version`                                    | `<tmp>/bin/<name>` (explicit `--bindir`)                 |
| rust    | `bin/cargo`                          | `cargo install --root <tmp> --version version pkg`                                                                   | `<tmp>/bin/<name>` (cargo's own `--root` convention)     |
| php     | system `composer` (see ruling below) | `composer require --working-dir=<tmp> --no-interaction pkg:version`                                                  | `<tmp>/vendor/bin/<name>` (composer's default `bin-dir`) |

`GOTOOLCHAIN=local` matters specifically: without it, `go install` can silently fetch and use a
_different_ Go toolchain version than the one rtunk just downloaded and pinned, defeating the
whole point of a hermetic runtime install.

Each command needs network access to its package registry (PyPI, the Go module proxy, RubyGems,
crates.io, Packagist) during install -- expected and consistent with `npm install` already
needing npmjs.org; nothing here is fetched at `rtunk exec` time, only at `rtunk download`/first
use, same as every other runtime.

Shim environment (PATH prepends for the runtime's own `bin/`, etc.) is already fully generic and
config-driven via `BuildEnv`/`WriteEnvShim` (`pkg/trunk/download/env.go`) -- each plugin's own
`runtime_environment`/`linter_environment` YAML entries declare what a shim needs on PATH, and
none of that is hardcoded per-runtime in Go today. **No changes needed here for any of the five
runtimes.**

## Ruling: `php`'s already-established `SystemVersion: "required"` exception

**Finding:** `php`'s `Runtime` definition has `SystemVersion: "required"` (confirmed in the real
cached plugin data), which `fetchRuntimeRef` (`pkg/trunk/download/download.go:127`) already
handles as a deliberate, pre-existing exception (built and reviewed in Task 9, v0.2): a
`system_version` runtime is always reported `Cached` and _never downloaded_ -- `installDir` for
`php` is never created, so there is no bundled `bin/composer` to find the way there's a bundled
`bin/npm`/`bin/pip`/etc.

**Ruling:** `runtime_php.go`'s `installPhpPackage` looks up `composer` via `exec.LookPath`
(the system `PATH`), not inside a `runtimeInstallDir` that will never exist for this runtime.
This does depart from `runtime_node.go`'s own doc-comment principle ("never a system npm, per
AGENTS.md Reproducibility -- no silent fallback to whatever happens to be on PATH") -- but `php`
itself _already_ is that exact fallback, by a decision this batch of work does not revisit.
Requiring a hermetically-downloaded Composer next to a system-provided PHP would be a stricter,
inconsistent middle ground: half-hermetic. Using the system's `composer`, right next to the
system's already-required `php`, is the one option consistent with `php`'s existing status quo.

**Cost if wrong:** if a machine has `php` on `PATH` but no `composer` (a real, plausible gap --
they're often installed separately), the 4 php-based tools (`paratest`, `php-cs-fixer`,
`phpstan`, `phpunit`) fail with a clear "composer not found on PATH" error rather than a silent
guess -- same failure shape `InstallPackage`'s existing `default:` branch already uses for
genuinely unimplemented runtimes. Recoverable by installing Composer; not data-loss, not a
reproducibility regression beyond what `php`'s own `SystemVersion: "required"` already accepted
in Task 9.

## Shim resolution: one new search path, for composer only

`composer require`'s default `bin-dir` is `vendor/bin`, not `bin/` -- add
`filepath.Join(installDir, "vendor", "bin", name)` to `shimSearchPaths`
(`pkg/trunk/download/shim.go`), the same way Fix 1 added `node_modules/.bin` for npm. This is the
only `shimSearchPaths` change in this batch; python/go/ruby/rust's chosen flags already funnel
into the existing `bin/` search path with zero code change there.

## Testing approach (carrying forward the lesson from the node_modules/.bin gap)

The final v0.2 branch review found a real production bug (`runtime:`+`package:` node fetch was
dead end-to-end) that 14 task-level reviews missed, because no test drove a fake package-manager
stub with the _exact_ real-world directory layout the actual tool produces. Each of this batch's
5 tasks must include, from its first commit (not bolted on later):

1. A unit test with a fake/stub package-manager script (mirroring `installNodePackage`'s existing
   test pattern) verifying the install command's arguments, environment, and working directory.
2. **A test asserting the resulting binary lands at the exact path `FindShimTarget` will look
   for** -- this is the specific class of bug that shipped once already; every task's test must
   assert on the concrete resulting path, not just "the command exited zero."
3. Where practical, a test using `finalizeInstall`'s existing atomic-rename/cache-poisoning
   regression pattern (Fix 3) -- a failed package-manager invocation must not leave
   `pkgInstallDir` looking cached. Since every one of these new functions is structured exactly
   like `installNodePackage` (temp dir + `finalizeInstall`), this should come for free from
   reusing that helper correctly, but each task must still prove it with a test, not assume it.

## File layout

- Create: `pkg/trunk/download/runtime_python.go`, `runtime_python_test.go`
- Create: `pkg/trunk/download/runtime_go.go`, `runtime_go_test.go`
- Create: `pkg/trunk/download/runtime_ruby.go`, `runtime_ruby_test.go`
- Create: `pkg/trunk/download/runtime_rust.go`, `runtime_rust_test.go`
- Create: `pkg/trunk/download/runtime_php.go`, `runtime_php_test.go`
- Modify: `pkg/trunk/download/runtime.go` (extend the `switch rt.Type` dispatch)
- Modify: `pkg/trunk/download/shim.go` (add the `vendor/bin` search path)

## Self-review

**Placeholder scan:** no TBD/TODO -- every command, flag, and env var above is a concrete,
checkable claim (verifiable against each package manager's real, stable CLI, and against this
TDD process's own tests).

**Internal consistency:** all five follow the same shape (locate binary → scratch temp dir →
package-manager-specific install command → `finalizeInstall`) except `php`'s binary-lookup
mechanism, which is explicitly and separately ruled above rather than silently inconsistent.

**Scope:** bounded to five files plus two small edits to existing files; no plugin-schema or
config changes; `java` explicitly out per user instruction.
