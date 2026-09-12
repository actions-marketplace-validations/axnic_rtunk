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
