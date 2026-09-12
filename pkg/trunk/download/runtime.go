package download

import (
	"fmt"

	"github.com/xunleii/rtunk/pkg/trunk/config"
)

// InstallPackage installs pkg@version through rt's own package manager into pkgInstallDir,
// dispatching to one file per runtime type (see runtime_node.go) so adding a runtime never
// touches another runtime's install logic. Implemented for node, python, go, ruby, rust, and php;
// java (and any future type) fails explicitly rather than guessing, since trunk's own per-runtime
// install commands for anything else aren't part of the open plugin schema (see the spec's
// "Fetch mechanisms").
func InstallPackage(rt config.Runtime, runtimeInstallDir, pkgInstallDir, pkg, version string) error {
	switch rt.Type {
	case "node":
		return installNodePackage(runtimeInstallDir, pkgInstallDir, pkg, version)
	case "python":
		return installPythonPackage(runtimeInstallDir, pkgInstallDir, pkg, version)
	case "go":
		return installGoPackage(runtimeInstallDir, pkgInstallDir, pkg, version)
	case "ruby":
		return installRubyPackage(runtimeInstallDir, pkgInstallDir, pkg, version)
	case "rust":
		return installRustPackage(runtimeInstallDir, pkgInstallDir, pkg, version)
	case "php":
		return installPhpPackage(pkgInstallDir, pkg, version)
	default:
		return fmt.Errorf("download: package-based fetch not yet supported for runtime %q", rt.Type)
	}
}

// ExtraToolEnv returns environment a runtime+package tool's shim needs beyond what the plugin's
// own runtime_environment/linter_environment config provides. Most runtimes need nothing extra
// (their install layout already matches what the plugin config assumes); python is the one
// exception -- pip's --prefix install has no venv/pyvenv.cfg, so nothing else gives the shim's
// python interpreter a way to find the installed package (see pythonSitePackages).
func ExtraToolEnv(rt config.Runtime, installDir string) ([]string, error) {
	switch rt.Type {
	case "python":
		sitePackages, err := pythonSitePackages(installDir)
		if err != nil {
			return nil, err
		}
		return []string{"PYTHONPATH=" + sitePackages}, nil
	default:
		return nil, nil
	}
}
