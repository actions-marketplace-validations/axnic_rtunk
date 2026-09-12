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

// fakeGo writes a stub `go` script into dir/bin that records its own argv AND its GOBIN env var
// value to argvFile, then creates a fake binary inside $GOBIN -- mimicking real `go install`'s
// actual on-disk effect (a binary directly inside GOBIN, no further subdirectory).
func fakeGo(t *testing.T, dir, argvFile string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bin"), 0o755))
	script := `#!/bin/sh
echo "$@ GOBIN=$GOBIN" > ` + argvFile + `
mkdir -p "$GOBIN"
echo '#!/bin/sh' > "$GOBIN/gofumpt"
chmod +x "$GOBIN/gofumpt"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bin", "go"), []byte(script), 0o755))
}

func TestInstallPackage_Go(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fakeGo is a POSIX shell script")
	}
	runtimeDir := t.TempDir()
	argvFile := filepath.Join(t.TempDir(), "argv")
	fakeGo(t, runtimeDir, argvFile)

	pkgDir := filepath.Join(t.TempDir(), "install")
	err := download.InstallPackage(config.Runtime{Type: "go"}, runtimeDir, pkgDir, "mvdan.cc/gofumpt", "0.6.0")
	require.NoError(t, err)

	argv, err := os.ReadFile(argvFile)
	require.NoError(t, err)
	line := strings.TrimSpace(string(argv))
	fields := strings.Fields(line)
	require.GreaterOrEqual(t, len(fields), 3)
	assert.Equal(t, "install", fields[0])
	assert.Equal(t, "mvdan.cc/gofumpt@0.6.0", fields[1])
	assert.True(t, strings.HasPrefix(fields[2], "GOBIN="))
	gobin := strings.TrimPrefix(fields[2], "GOBIN=")
	assert.NotEqual(t, pkgDir, gobin, "go install must run against a scratch GOBIN, not pkgDir directly")
	assert.Equal(t, filepath.Dir(pkgDir), filepath.Dir(filepath.Dir(gobin)), "the scratch dir must be a sibling of pkgDir (same filesystem for the final rename)")

	assert.DirExists(t, pkgDir, "a successful go install must be renamed into pkgDir")
	target, err := download.FindShimTarget(pkgDir, "gofumpt")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(pkgDir, "bin", "gofumpt"), target)
}
