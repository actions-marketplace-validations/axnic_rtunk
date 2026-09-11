package download_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xunleii/rtunk/pkg/trunk/config"
	"github.com/xunleii/rtunk/pkg/trunk/download"
)

// fakeNpm writes a stub `npm` script into dir/bin that records its own argv to argvFile instead
// of touching the real network -- runtime_node.go only needs to know it invoked npm correctly,
// not that npm itself works.
func fakeNpm(t *testing.T, dir, argvFile string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bin"), 0o755))
	script := "#!/bin/sh\necho \"$@\" > " + argvFile + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bin", "npm"), []byte(script), 0o755))
}

func TestInstallPackage_Node(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fakeNpm is a POSIX shell script")
	}
	runtimeDir := t.TempDir()
	argvFile := filepath.Join(t.TempDir(), "argv")
	fakeNpm(t, runtimeDir, argvFile)

	pkgDir := t.TempDir()
	err := download.InstallPackage(config.Runtime{Type: "node"}, runtimeDir, pkgDir, "eslint", "8.10.0")
	require.NoError(t, err)

	argv, err := os.ReadFile(argvFile)
	require.NoError(t, err)
	assert.Equal(t, "install --prefix "+pkgDir+" eslint@8.10.0\n", string(argv))
}

func TestInstallPackage_UnsupportedRuntime(t *testing.T) {
	err := download.InstallPackage(config.Runtime{Type: "python"}, t.TempDir(), t.TempDir(), "black", "24.0.0")
	assert.ErrorContains(t, err, "python")
}
