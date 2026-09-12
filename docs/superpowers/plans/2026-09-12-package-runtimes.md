# Package-Based Fetch for python/go/ruby/rust/php Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend `pkg/trunk/download`'s `InstallPackage` from `node`-only to also support
`python`, `go`, `ruby`, `rust`, and `php` (explicitly excluding `java` per user request), so
`runtime:`+`package:` tools using those runtimes (28/13/4/1/4 tools respectively in the real
trunk-io plugin data) can actually be fetched.

**Architecture:** One new file per runtime in `pkg/trunk/download`, each mirroring
`runtime_node.go`'s existing shape exactly (locate the runtime's own bundled package-manager
binary, install into a scratch temp dir, `finalizeInstall` into place atomically), wired into
`InstallPackage`'s existing `switch rt.Type` in `runtime.go`. `php` is the one exception: its
runtime is already `SystemVersion: "required"` (never downloaded, decided in v0.2's Task 9), so
its package manager (`composer`) is looked up on the system `PATH` instead of inside a
`runtimeInstallDir` that will never exist.

**Tech Stack:** Go 1.27, stdlib `os/exec` only (no new dependencies).

**Spec:** `docs/superpowers/specs/2026-09-12-package-runtimes-design.md`

## Global Constraints

- Go 1.27 (go.mod). No new go.mod dependencies — stdlib `os/exec`/`os`/`path/filepath` only.
- `java` is explicitly out of scope for this plan.
- Reuse `finalizeInstall` (`pkg/trunk/download/extract.go`) unchanged in every new
  `install*Package` function — do not reimplement the scratch-dir/atomic-rename pattern.
- `shimSearchPaths` (`pkg/trunk/download/shim.go`) gets exactly one addition across this whole
  plan: `vendor/bin`, for `php`/composer (Task 5). python/go/ruby/rust must land their output
  under the existing `bin/` search path via the install command's own flags/env vars — no other
  `shimSearchPaths` changes.
- `BuildEnv`/`WriteEnvShim` (`pkg/trunk/download/env.go`) are unchanged by this plan — shim
  PATH/environment setup is already fully config-driven, not hardcoded per runtime.
- No changes to `pkg/trunk/config` or the plugin schema.
- Every new exported symbol needs a doc comment explaining why, not what.
- Every task's test must assert the exact resulting binary path via `FindShimTarget`, not just
  "the command exited zero" — this is the specific regression class (`node_modules/.bin`) that
  shipped once already in v0.2 and slipped past 14 task reviews.
- Commit convention: `.agents/skills/git-commit/SKILL.md`. This is all `pkg/trunk/download` work,
  scope `cache`. Sign every commit with `-S`; never `--no-gpg-sign`/`-c commit.gpgsign=false`/
  unsigned `-s` without the user's explicit, one-time consent for that specific commit.
- This repo's real commitlint rule `body-max-line-length: 80` applies to every commit body.

---

### Task 1: `python` package installs (pip)

**Files:**

- Create: `pkg/trunk/download/runtime_python.go`
- Create: `pkg/trunk/download/runtime_python_test.go`
- Modify: `pkg/trunk/download/runtime.go` (add the `"python"` case)
- Modify: `pkg/trunk/download/runtime_node_test.go` (`TestInstallPackage_UnsupportedRuntime` currently
  asserts `Type: "python"` errors — that stops being true after this task, so it must be repointed
  at a runtime that stays genuinely unsupported)

**Interfaces:**

- Consumes: `finalizeInstall(tmpDir, destDir string) error` (`extract.go`) — reuse unchanged.
- Produces: `installPythonPackage(runtimeInstallDir, pkgInstallDir, pkg, version string) error`,
  wired into `InstallPackage`'s switch.

- [ ] **Step 1: Write failing test**

```go
package download_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xunleii/rtunk/pkg/trunk/config"
	"github.com/xunleii/rtunk/pkg/trunk/download"
)

// fakePip writes a stub `pip` script into dir/bin that records its own argv to argvFile and,
// mimicking real `pip install --prefix <dir>`'s actual on-disk effect, creates a console-script
// executable at <dir>/bin/black -- runtime_python.go only needs to know it invoked pip correctly
// and that the result lands where FindShimTarget looks, not that pip itself works.
func fakePip(t *testing.T, dir, argvFile string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bin"), 0o755))
	script := `#!/bin/sh
echo "$@" > ` + argvFile + `
prefix=$3
mkdir -p "$prefix/bin"
echo '#!/bin/sh' > "$prefix/bin/black"
chmod +x "$prefix/bin/black"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bin", "pip"), []byte(script), 0o755))
}

func TestInstallPackage_Python(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fakePip is a POSIX shell script")
	}
	runtimeDir := t.TempDir()
	argvFile := filepath.Join(t.TempDir(), "argv")
	fakePip(t, runtimeDir, argvFile)

	pkgDir := filepath.Join(t.TempDir(), "install")
	err := download.InstallPackage(config.Runtime{Type: "python"}, runtimeDir, pkgDir, "black", "24.0.0")
	require.NoError(t, err)

	argv, err := os.ReadFile(argvFile)
	require.NoError(t, err)
	fields := strings.Fields(string(argv))
	require.Len(t, fields, 4)
	assert.Equal(t, "install", fields[0])
	assert.Equal(t, "--prefix", fields[1])
	assert.NotEqual(t, pkgDir, fields[2], "pip must run against a scratch temp dir, not pkgDir directly")
	assert.Equal(t, filepath.Dir(pkgDir), filepath.Dir(fields[2]), "the scratch dir must be a sibling of pkgDir (same filesystem for the final rename)")
	assert.Equal(t, "black==24.0.0", fields[3])

	assert.DirExists(t, pkgDir, "a successful pip install must be renamed into pkgDir")
	target, err := download.FindShimTarget(pkgDir, "black")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(pkgDir, "bin", "black"), target)
}
```

Also update `runtime_node_test.go`'s existing `TestInstallPackage_UnsupportedRuntime`:

```go
func TestInstallPackage_UnsupportedRuntime(t *testing.T) {
	err := download.InstallPackage(config.Runtime{Type: "java"}, t.TempDir(), t.TempDir(), "checkstyle", "10.0.0")
	assert.ErrorContains(t, err, "java")
}
```

(It previously used `Type: "python"` — that stops being a valid "unsupported" example the moment
this task lands, so it must be repointed at `java`, which stays unsupported by this plan's own
Global Constraints.)

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/trunk/download/... -run TestInstallPackage_Python -v`
Expected: FAIL — `python` isn't a case in `InstallPackage`'s switch yet.

- [ ] **Step 3: Write minimal implementation**

```go
package download

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// installPythonPackage runs `pip install --prefix <scratch dir> pkg==version` using the pip
// shipped by the already-downloaded python runtime at runtimeInstallDir (never a system pip, per
// AGENTS.md "Reproducibility" -- no silent fallback to whatever happens to be on PATH). pip's
// --prefix scheme places console-script entry points at <prefix>/bin/<name>, matching
// shimSearchPaths' existing bin/ check with no further changes needed there.
func installPythonPackage(runtimeInstallDir, pkgInstallDir, pkg, version string) error {
	pip := filepath.Join(runtimeInstallDir, "bin", "pip")
	if _, err := os.Stat(pip); err != nil {
		return fmt.Errorf("download: pip not found at %s: %w", pip, err)
	}
	if err := os.MkdirAll(filepath.Dir(pkgInstallDir), 0o755); err != nil {
		return err
	}
	tmpDir, err := os.MkdirTemp(filepath.Dir(pkgInstallDir), ".tmp-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir) // no-op once finalizeInstall renames it into pkgInstallDir

	cmd := exec.Command(pip, "install", "--prefix", tmpDir, pkg+"=="+version)
	cmd.Env = append(os.Environ(), "PATH="+filepath.Join(runtimeInstallDir, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("download: pip install %s==%s: %w: %s", pkg, version, err, out)
	}
	return finalizeInstall(tmpDir, pkgInstallDir)
}
```

In `pkg/trunk/download/runtime.go`, add the case:

```go
	case "python":
		return installPythonPackage(runtimeInstallDir, pkgInstallDir, pkg, version)
```

(alongside the existing `case "node":`, before the `default:` branch).

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/trunk/download/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/trunk/download/runtime_python.go pkg/trunk/download/runtime_python_test.go \
  pkg/trunk/download/runtime.go pkg/trunk/download/runtime_node_test.go
git commit -S -m "+[cache]: Add python package installs via pip" \
  -m "Runs pip install --prefix <scratch> against the runtime's own bundled" \
  -m "pip, never a system one. pip's --prefix scheme already lands the" \
  -m "console-script at bin/<name>, matching the existing shim search path." \
  -m "Assisted-by: anthropic:claude-sonnet-5"
```

---

### Task 2: `go` package installs (go install)

**Files:**

- Create: `pkg/trunk/download/runtime_go.go`
- Create: `pkg/trunk/download/runtime_go_test.go`
- Modify: `pkg/trunk/download/runtime.go` (add the `"go"` case)

**Interfaces:**

- Consumes: `finalizeInstall` (unchanged, as in Task 1).
- Produces: `installGoPackage(runtimeInstallDir, pkgInstallDir, pkg, version string) error`.

- [ ] **Step 1: Write failing test**

```go
package download_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xunleii/rtunk/pkg/trunk/config"
	"github.com/xunleii/rtunk/pkg/trunk/download"
)

// fakeGo writes a stub `go` script into dir/bin that records its own argv AND its GOBIN env var
// value to argvFile, then creates a fake binary inside $GOBIN -- mimicking real `go install`'s
// actual on-disk effect (a binary directly inside GOBIN, no further subdirectory).
func fakeGo(t *testing.T, dir, argvFile string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bin"), 0o755))
	script := `#!/bin/sh
echo "$@ GOBIN=$GOBIN" > ` + argvFile + `
mkdir -p "$GOBIN"
echo '#!/bin/sh' > "$GOBIN/gofumpt"
chmod +x "$GOBIN/gofumpt"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bin", "go"), []byte(script), 0o755))
}

func TestInstallPackage_Go(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fakeGo is a POSIX shell script")
	}
	runtimeDir := t.TempDir()
	argvFile := filepath.Join(t.TempDir(), "argv")
	fakeGo(t, runtimeDir, argvFile)

	pkgDir := filepath.Join(t.TempDir(), "install")
	err := download.InstallPackage(config.Runtime{Type: "go"}, runtimeDir, pkgDir, "mvdan.cc/gofumpt", "0.6.0")
	require.NoError(t, err)

	argv, err := os.ReadFile(argvFile)
	require.NoError(t, err)
	line := strings.TrimSpace(string(argv))
	fields := strings.Fields(line)
	require.GreaterOrEqual(t, len(fields), 3)
	assert.Equal(t, "install", fields[0])
	assert.Equal(t, "mvdan.cc/gofumpt@0.6.0", fields[1])
	assert.True(t, strings.HasPrefix(fields[2], "GOBIN="))
	gobin := strings.TrimPrefix(fields[2], "GOBIN=")
	assert.NotEqual(t, pkgDir, gobin, "go install must run against a scratch GOBIN, not pkgDir directly")
	assert.Equal(t, filepath.Dir(pkgDir), filepath.Dir(filepath.Dir(gobin)), "the scratch dir must be a sibling of pkgDir (same filesystem for the final rename)")

	assert.DirExists(t, pkgDir, "a successful go install must be renamed into pkgDir")
	target, err := download.FindShimTarget(pkgDir, "gofumpt")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(pkgDir, "bin", "gofumpt"), target)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/trunk/download/... -run TestInstallPackage_Go -v`
Expected: FAIL — `go` isn't a case in `InstallPackage`'s switch yet.

- [ ] **Step 3: Write minimal implementation**

```go
package download

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// installGoPackage runs `go install pkg@version` using the go toolchain shipped by the
// already-downloaded go runtime at runtimeInstallDir (never a system go, per AGENTS.md
// "Reproducibility"). GOBIN is pointed at a scratch bin/ dir so the built binary lands where
// shimSearchPaths already looks. GOTOOLCHAIN=local pins go to the toolchain version we just
// downloaded, instead of letting a `go` newer than the target module's go.mod silently fetch and
// use a completely different toolchain version -- defeating the point of a hermetic runtime.
func installGoPackage(runtimeInstallDir, pkgInstallDir, pkg, version string) error {
	goBin := filepath.Join(runtimeInstallDir, "bin", "go")
	if _, err := os.Stat(goBin); err != nil {
		return fmt.Errorf("download: go not found at %s: %w", goBin, err)
	}
	if err := os.MkdirAll(filepath.Dir(pkgInstallDir), 0o755); err != nil {
		return err
	}
	tmpDir, err := os.MkdirTemp(filepath.Dir(pkgInstallDir), ".tmp-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	binDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}

	cmd := exec.Command(goBin, "install", pkg+"@"+version)
	cmd.Env = append(os.Environ(),
		"PATH="+filepath.Join(runtimeInstallDir, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"),
		"GOBIN="+binDir,
		"GOPATH="+filepath.Join(tmpDir, "gopath"),
		"GOCACHE="+filepath.Join(tmpDir, "gocache"),
		"GOTOOLCHAIN=local",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("download: go install %s@%s: %w: %s", pkg, version, err, out)
	}
	return finalizeInstall(tmpDir, pkgInstallDir)
}
```

In `pkg/trunk/download/runtime.go`, add:

```go
	case "go":
		return installGoPackage(runtimeInstallDir, pkgInstallDir, pkg, version)
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/trunk/download/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/trunk/download/runtime_go.go pkg/trunk/download/runtime_go_test.go pkg/trunk/download/runtime.go
git commit -S -m "+[cache]: Add go package installs via go install" \
  -m "Runs go install against the runtime's own bundled go toolchain, never" \
  -m "a system one. GOBIN points at a scratch bin/ dir so the built binary" \
  -m "lands at the existing shim search path; GOTOOLCHAIN=local pins the" \
  -m "toolchain to the one rtunk downloaded, not whatever a target" \
  -m "module's go.mod would otherwise fetch." \
  -m "Assisted-by: anthropic:claude-sonnet-5"
```

---

### Task 3: `ruby` package installs (gem)

**Files:**

- Create: `pkg/trunk/download/runtime_ruby.go`
- Create: `pkg/trunk/download/runtime_ruby_test.go`
- Modify: `pkg/trunk/download/runtime.go` (add the `"ruby"` case)

**Interfaces:**

- Consumes: `finalizeInstall` (unchanged).
- Produces: `installRubyPackage(runtimeInstallDir, pkgInstallDir, pkg, version string) error`.

- [ ] **Step 1: Write failing test**

```go
package download_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xunleii/rtunk/pkg/trunk/config"
	"github.com/xunleii/rtunk/pkg/trunk/download"
)

// fakeGem writes a stub `gem` script into dir/bin that records its own argv to argvFile and
// creates a fake executable at the --bindir argument the way real `gem install --bindir` would.
func fakeGem(t *testing.T, dir, argvFile string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bin"), 0o755))
	script := `#!/bin/sh
echo "$@" > ` + argvFile + `
bindir=""
prev=""
for arg in "$@"; do
  if [ "$prev" = "--bindir" ]; then bindir=$arg; fi
  prev=$arg
done
mkdir -p "$bindir"
echo '#!/bin/sh' > "$bindir/rufo"
chmod +x "$bindir/rufo"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bin", "gem"), []byte(script), 0o755))
}

func TestInstallPackage_Ruby(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fakeGem is a POSIX shell script")
	}
	runtimeDir := t.TempDir()
	argvFile := filepath.Join(t.TempDir(), "argv")
	fakeGem(t, runtimeDir, argvFile)

	pkgDir := filepath.Join(t.TempDir(), "install")
	err := download.InstallPackage(config.Runtime{Type: "ruby"}, runtimeDir, pkgDir, "rufo", "0.15.0")
	require.NoError(t, err)

	argv, err := os.ReadFile(argvFile)
	require.NoError(t, err)
	fields := strings.Fields(string(argv))
	assert.Contains(t, fields, "--install-dir")
	assert.Contains(t, fields, "--bindir")
	assert.Contains(t, fields, "rufo")
	assert.Contains(t, fields, "-v")
	assert.Contains(t, fields, "0.15.0")

	assert.DirExists(t, pkgDir, "a successful gem install must be renamed into pkgDir")
	target, err := download.FindShimTarget(pkgDir, "rufo")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(pkgDir, "bin", "rufo"), target)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/trunk/download/... -run TestInstallPackage_Ruby -v`
Expected: FAIL — `ruby` isn't a case in `InstallPackage`'s switch yet.

- [ ] **Step 3: Write minimal implementation**

```go
package download

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// installRubyPackage runs `gem install --install-dir <scratch> --bindir <scratch>/bin pkg -v
// version` using the gem shipped by the already-downloaded ruby runtime at runtimeInstallDir
// (never a system gem, per AGENTS.md "Reproducibility"). --bindir explicitly controls where the
// executable lands, matching shimSearchPaths' existing bin/ check with no further changes.
func installRubyPackage(runtimeInstallDir, pkgInstallDir, pkg, version string) error {
	gem := filepath.Join(runtimeInstallDir, "bin", "gem")
	if _, err := os.Stat(gem); err != nil {
		return fmt.Errorf("download: gem not found at %s: %w", gem, err)
	}
	if err := os.MkdirAll(filepath.Dir(pkgInstallDir), 0o755); err != nil {
		return err
	}
	tmpDir, err := os.MkdirTemp(filepath.Dir(pkgInstallDir), ".tmp-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command(gem, "install", "--no-document",
		"--install-dir", tmpDir, "--bindir", filepath.Join(tmpDir, "bin"),
		pkg, "-v", version)
	cmd.Env = append(os.Environ(), "PATH="+filepath.Join(runtimeInstallDir, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("download: gem install %s -v %s: %w: %s", pkg, version, err, out)
	}
	return finalizeInstall(tmpDir, pkgInstallDir)
}
```

In `pkg/trunk/download/runtime.go`, add:

```go
	case "ruby":
		return installRubyPackage(runtimeInstallDir, pkgInstallDir, pkg, version)
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/trunk/download/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/trunk/download/runtime_ruby.go pkg/trunk/download/runtime_ruby_test.go pkg/trunk/download/runtime.go
git commit -S -m "+[cache]: Add ruby package installs via gem" \
  -m "Runs gem install --bindir against the runtime's own bundled gem," \
  -m "never a system one. --bindir explicitly controls where the" \
  -m "executable lands, matching the existing shim search path." \
  -m "Assisted-by: anthropic:claude-sonnet-5"
```

---

### Task 4: `rust` package installs (cargo)

**Files:**

- Create: `pkg/trunk/download/runtime_rust.go`
- Create: `pkg/trunk/download/runtime_rust_test.go`
- Modify: `pkg/trunk/download/runtime.go` (add the `"rust"` case)

**Interfaces:**

- Consumes: `finalizeInstall` (unchanged).
- Produces: `installRustPackage(runtimeInstallDir, pkgInstallDir, pkg, version string) error`.

- [ ] **Step 1: Write failing test**

```go
package download_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xunleii/rtunk/pkg/trunk/config"
	"github.com/xunleii/rtunk/pkg/trunk/download"
)

// fakeCargo writes a stub `cargo` script into dir/bin that records its own argv to argvFile and
// creates a fake binary directly under the --root argument's bin/ subdirectory, the way real
// `cargo install --root` actually lays its output out.
func fakeCargo(t *testing.T, dir, argvFile string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bin"), 0o755))
	script := `#!/bin/sh
echo "$@" > ` + argvFile + `
root=""
prev=""
for arg in "$@"; do
  if [ "$prev" = "--root" ]; then root=$arg; fi
  prev=$arg
done
mkdir -p "$root/bin"
echo '#!/bin/sh' > "$root/bin/ripgrep"
chmod +x "$root/bin/ripgrep"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bin", "cargo"), []byte(script), 0o755))
}

func TestInstallPackage_Rust(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fakeCargo is a POSIX shell script")
	}
	runtimeDir := t.TempDir()
	argvFile := filepath.Join(t.TempDir(), "argv")
	fakeCargo(t, runtimeDir, argvFile)

	pkgDir := filepath.Join(t.TempDir(), "install")
	err := download.InstallPackage(config.Runtime{Type: "rust"}, runtimeDir, pkgDir, "ripgrep", "14.1.0")
	require.NoError(t, err)

	argv, err := os.ReadFile(argvFile)
	require.NoError(t, err)
	fields := strings.Fields(string(argv))
	assert.Contains(t, fields, "--root")
	assert.Contains(t, fields, "--version")
	assert.Contains(t, fields, "14.1.0")
	assert.Contains(t, fields, "ripgrep")

	assert.DirExists(t, pkgDir, "a successful cargo install must be renamed into pkgDir")
	target, err := download.FindShimTarget(pkgDir, "ripgrep")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(pkgDir, "bin", "ripgrep"), target)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/trunk/download/... -run TestInstallPackage_Rust -v`
Expected: FAIL — `rust` isn't a case in `InstallPackage`'s switch yet.

- [ ] **Step 3: Write minimal implementation**

```go
package download

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// installRustPackage runs `cargo install --root <scratch> --version version pkg` using the cargo
// shipped by the already-downloaded rust runtime at runtimeInstallDir (never a system cargo, per
// AGENTS.md "Reproducibility"). cargo's own --root convention places binaries at
// <root>/bin/<name>, matching shimSearchPaths' existing bin/ check with no further changes.
func installRustPackage(runtimeInstallDir, pkgInstallDir, pkg, version string) error {
	cargo := filepath.Join(runtimeInstallDir, "bin", "cargo")
	if _, err := os.Stat(cargo); err != nil {
		return fmt.Errorf("download: cargo not found at %s: %w", cargo, err)
	}
	if err := os.MkdirAll(filepath.Dir(pkgInstallDir), 0o755); err != nil {
		return err
	}
	tmpDir, err := os.MkdirTemp(filepath.Dir(pkgInstallDir), ".tmp-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command(cargo, "install", "--root", tmpDir, "--version", version, pkg)
	cmd.Env = append(os.Environ(), "PATH="+filepath.Join(runtimeInstallDir, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("download: cargo install %s --version %s: %w: %s", pkg, version, err, out)
	}
	return finalizeInstall(tmpDir, pkgInstallDir)
}
```

In `pkg/trunk/download/runtime.go`, add:

```go
	case "rust":
		return installRustPackage(runtimeInstallDir, pkgInstallDir, pkg, version)
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/trunk/download/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/trunk/download/runtime_rust.go pkg/trunk/download/runtime_rust_test.go pkg/trunk/download/runtime.go
git commit -S -m "+[cache]: Add rust package installs via cargo" \
  -m "Runs cargo install --root against the runtime's own bundled cargo," \
  -m "never a system one. cargo's own --root convention already lands" \
  -m "binaries at bin/<name>, matching the existing shim search path." \
  -m "Assisted-by: anthropic:claude-sonnet-5"
```

---

### Task 5: `php` package installs (composer, system PATH)

**Files:**

- Create: `pkg/trunk/download/runtime_php.go`
- Create: `pkg/trunk/download/runtime_php_test.go`
- Modify: `pkg/trunk/download/runtime.go` (add the `"php"` case)
- Modify: `pkg/trunk/download/shim.go` (add the `vendor/bin` search path)

**Interfaces:**

- Consumes: `finalizeInstall` (unchanged).
- Produces: `installPhpPackage(pkgInstallDir, pkg, version string) error` — note this does NOT
  take `runtimeInstallDir`: `php`'s runtime is `SystemVersion: "required"` (v0.2's Task 9,
  `fetchRuntimeRef`, `pkg/trunk/download/download.go:127`) and is never downloaded, so there is no
  `runtimeInstallDir` to look inside. `composer` is found via `exec.LookPath` on the system
  `PATH` instead — see the design doc's ruling on this exception before touching this task.

- [ ] **Step 1: Write failing test**

```go
package download_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xunleii/rtunk/pkg/trunk/config"
	"github.com/xunleii/rtunk/pkg/trunk/download"
)

// fakeComposer writes a stub `composer` script into dir that records its own argv to argvFile
// and creates a fake executable under vendor/bin inside the --working-dir argument, the way real
// `composer require`'s default bin-dir actually lays its output out.
func fakeComposer(t *testing.T, dir, argvFile string) {
	t.Helper()
	script := `#!/bin/sh
echo "$@" > ` + argvFile + `
wd=""
prev=""
for arg in "$@"; do
  case $arg in
    --working-dir=*) wd=${arg#--working-dir=} ;;
  esac
  prev=$arg
done
mkdir -p "$wd/vendor/bin"
echo '#!/bin/sh' > "$wd/vendor/bin/php-cs-fixer"
chmod +x "$wd/vendor/bin/php-cs-fixer"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "composer"), []byte(script), 0o755))
}

func TestInstallPackage_Php(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fakeComposer is a POSIX shell script")
	}
	fakeComposerDir := t.TempDir()
	argvFile := filepath.Join(t.TempDir(), "argv")
	fakeComposer(t, fakeComposerDir, argvFile)
	t.Setenv("PATH", fakeComposerDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	pkgDir := filepath.Join(t.TempDir(), "install")
	// runtimeInstallDir is irrelevant for php -- pass a path that doesn't exist to prove it's
	// never touched.
	err := download.InstallPackage(config.Runtime{Type: "php"}, "/does/not/exist", pkgDir, "friendsofphp/php-cs-fixer", "3.40.0")
	require.NoError(t, err)

	argv, err := os.ReadFile(argvFile)
	require.NoError(t, err)
	fields := strings.Fields(string(argv))
	assert.Contains(t, fields, "require")
	assert.Contains(t, fields, "--no-interaction")
	assert.Contains(t, fields, "friendsofphp/php-cs-fixer:3.40.0")

	assert.DirExists(t, pkgDir, "a successful composer require must be renamed into pkgDir")
	target, err := download.FindShimTarget(pkgDir, "php-cs-fixer")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(pkgDir, "vendor", "bin", "php-cs-fixer"), target)
}

func TestInstallPackage_Php_ComposerNotOnPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // empty dir, guaranteed no composer
	err := download.InstallPackage(config.Runtime{Type: "php"}, "/does/not/exist", filepath.Join(t.TempDir(), "install"), "pkg", "1.0.0")
	assert.ErrorContains(t, err, "composer")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/trunk/download/... -run TestInstallPackage_Php -v`
Expected: FAIL — `php` isn't a case in `InstallPackage`'s switch yet, and `vendor/bin` isn't in
`shimSearchPaths` yet.

- [ ] **Step 3: Write minimal implementation**

```go
package download

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// installPhpPackage runs `composer require --working-dir <scratch> --no-interaction pkg:version`
// using composer found on the system PATH. Unlike every other runtime in this file set, php's own
// runtime is never downloaded -- Runtime.SystemVersion == "required" (v0.2's Task 9,
// fetchRuntimeRef), so runtimeInstallDir never exists for php and there is no bundled composer to
// find there. Using the system's composer is a deliberate, documented exception (see
// docs/superpowers/specs/2026-09-12-package-runtimes-design.md's ruling): php already relies on
// the system for its own interpreter, so requiring a hermetically-downloaded Composer next to a
// system-provided PHP would be a stricter, inconsistent half-hermetic middle ground.
func installPhpPackage(pkgInstallDir, pkg, version string) error {
	composer, err := exec.LookPath("composer")
	if err != nil {
		return fmt.Errorf("download: composer not found on PATH: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(pkgInstallDir), 0o755); err != nil {
		return err
	}
	tmpDir, err := os.MkdirTemp(filepath.Dir(pkgInstallDir), ".tmp-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command(composer, "require", "--working-dir="+tmpDir, "--no-interaction", pkg+":"+version)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("download: composer require %s:%s: %w: %s", pkg, version, err, out)
	}
	return finalizeInstall(tmpDir, pkgInstallDir)
}
```

In `pkg/trunk/download/runtime.go`, add:

```go
	case "php":
		return installPhpPackage(pkgInstallDir, pkg, version)
```

In `pkg/trunk/download/shim.go`'s `shimSearchPaths`, add one more entry:

```go
		// composer require --working-dir <installDir> (installPhpPackage) lays its bins out
		// here (composer's own default bin-dir), not at <installDir>/bin.
		filepath.Join(installDir, "vendor", "bin", name),
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/trunk/download/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/trunk/download/runtime_php.go pkg/trunk/download/runtime_php_test.go \
  pkg/trunk/download/runtime.go pkg/trunk/download/shim.go
git commit -S -m "+[cache]: Add php package installs via system composer" \
  -m "php's runtime is already system-required (never downloaded, v0.2" \
  -m "Task 9), so composer is found on the system PATH too rather than" \
  -m "inventing a half-hermetic middle ground. Adds vendor/bin to the shim" \
  -m "search path, composer's own default bin-dir." \
  -m "Assisted-by: anthropic:claude-sonnet-5"
```

---

## Self-Review

**Spec coverage:** every runtime named in the spec (python, go, ruby, rust, php) has a task;
`java` is correctly absent per the spec's non-goals; the one `shimSearchPaths` addition (Task 5,
`vendor/bin`) matches the spec's "one new search path, for composer only" section exactly;
python/go/ruby/rust's tasks correctly make no `shimSearchPaths` change.

**Placeholder scan:** no TBD/TODO; every task's implementation is complete, runnable code, not a
sketch.

**Type consistency:** every `install*Package` function's signature matches how `runtime.go`'s
switch calls it (php's genuinely differs — no `runtimeInstallDir` param — and that's called out
explicitly in Task 5's Interfaces section, not a silent inconsistency). `finalizeInstall`'s
signature (`tmpDir, destDir string) error`) is used identically across all five.

**Ordering:** Task 1 must run before Tasks 2-5 since it's the one that repoints
`TestInstallPackage_UnsupportedRuntime` away from `"python"` — if a later task ran first and
implemented `"python"` speculatively that test would silently start asserting nothing meaningful.
Tasks 2-5 have no ordering dependency on each other beyond all sharing `runtime.go`'s switch
statement (each adds one case; sequential dispatch, never parallel, avoids any conflict there).

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-09-12-package-runtimes.md`.
Executing via **Subagent-Driven** (superpowers:subagent-driven-development): fresh subagent per
task, task review after each, whole-branch review at the end — same process used for v0.2.
