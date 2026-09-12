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
