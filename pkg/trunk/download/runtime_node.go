package download

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// installNodePackage runs `npm install --prefix <scratch dir> pkg@version` using the npm shipped
// by the already-downloaded node runtime at runtimeInstallDir (never a system npm, per AGENTS.md
// "Reproducibility" -- no silent fallback to whatever happens to be on PATH). Each tool gets its
// own pkgInstallDir, so its node_modules never shares or conflicts with another tool's.
//
// npm installs into a scratch temp directory (a sibling of pkgInstallDir, so the final os.Rename
// stays on one filesystem) and is only published to pkgInstallDir via finalizeInstall once it has
// actually succeeded (see Fix 3) -- otherwise a failed/killed `npm install` left pkgInstallDir
// existing (MkdirAll used to be this function's first action), and dirNonEmpty(pkgInstallDir)
// would wrongly treat that as a completed, cached install forever.
func installNodePackage(runtimeInstallDir, pkgInstallDir, pkg, version string) error {
	npm := filepath.Join(runtimeInstallDir, "bin", "npm")
	if _, err := os.Stat(npm); err != nil {
		return fmt.Errorf("download: npm not found at %s: %w", npm, err)
	}
	if err := os.MkdirAll(filepath.Dir(pkgInstallDir), 0o755); err != nil {
		return err
	}
	tmpDir, err := os.MkdirTemp(filepath.Dir(pkgInstallDir), ".tmp-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir) // no-op once finalizeInstall renames it into pkgInstallDir

	cmd := exec.Command(npm, "install", "--prefix", tmpDir, pkg+"@"+version)
	cmd.Env = append(os.Environ(), "PATH="+filepath.Join(runtimeInstallDir, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("download: npm install %s@%s: %w: %s", pkg, version, err, out)
	}
	return finalizeInstall(tmpDir, pkgInstallDir)
}
