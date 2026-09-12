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
// CARGO_TARGET_DIR is pointed at a scratch build directory so cargo's build lock and ephemera
// stay isolated from any inherited CARGO_TARGET_DIR, preventing lock contention with other cargo
// processes on the host (see Task 2's GOCACHE pattern for the equivalent Go isolation).
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
	cmd.Env = append(os.Environ(),
		"PATH="+filepath.Join(runtimeInstallDir, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"),
		"CARGO_TARGET_DIR="+filepath.Join(tmpDir, "target"),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("download: cargo install %s --version %s: %w: %s", pkg, version, err, out)
	}
	return finalizeInstall(tmpDir, pkgInstallDir)
}
