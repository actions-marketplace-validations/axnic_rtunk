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
