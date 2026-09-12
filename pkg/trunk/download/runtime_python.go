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
