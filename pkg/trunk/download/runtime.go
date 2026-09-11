package download

import (
	"fmt"

	"github.com/xunleii/rtunk/pkg/trunk/config"
)

// InstallPackage installs pkg@version through rt's own package manager into pkgInstallDir,
// dispatching to one file per runtime type (see runtime_node.go) so adding a runtime never
// touches another runtime's install logic. v0.2 implements node only -- trunk's own per-runtime
// install commands for anything else aren't part of the open plugin schema (see the spec's
// "Fetch mechanisms"), so an unimplemented runtime fails explicitly instead of guessing.
func InstallPackage(rt config.Runtime, runtimeInstallDir, pkgInstallDir, pkg, version string) error {
	switch rt.Type {
	case "node":
		return installNodePackage(runtimeInstallDir, pkgInstallDir, pkg, version)
	default:
		return fmt.Errorf("download: package-based fetch not yet supported for runtime %q", rt.Type)
	}
}
