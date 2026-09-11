package download

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// installNodePackage runs `npm install --prefix pkgInstallDir pkg@version` using the npm shipped
// by the already-downloaded node runtime at runtimeInstallDir (never a system npm, per AGENTS.md
// "Reproducibility" -- no silent fallback to whatever happens to be on PATH). Each tool gets its
// own pkgInstallDir, so its node_modules never shares or conflicts with another tool's.
func installNodePackage(runtimeInstallDir, pkgInstallDir, pkg, version string) error {
	npm := filepath.Join(runtimeInstallDir, "bin", "npm")
	if _, err := os.Stat(npm); err != nil {
		return fmt.Errorf("download: npm not found at %s: %w", npm, err)
	}
	if err := os.MkdirAll(pkgInstallDir, 0o755); err != nil {
		return err
	}
	cmd := exec.Command(npm, "install", "--prefix", pkgInstallDir, pkg+"@"+version)
	cmd.Env = append(os.Environ(), "PATH="+filepath.Join(runtimeInstallDir, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("download: npm install %s@%s: %w: %s", pkg, version, err, out)
	}
	return nil
}
