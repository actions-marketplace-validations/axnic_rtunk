package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xunleii/rtunk/pkg/trunk/download"
)

func TestPruneUnused(t *testing.T) {
	root := t.TempDir()
	keep := download.InstallDir(root, "tools", "actionlint", "1.0.0")
	stale := download.InstallDir(root, "tools", "eslint", "1.0.0")
	keepShim := download.ShimPath(root, "tools", "actionlint", "1.0.0", "actionlint")
	staleShim := download.ShimPath(root, "tools", "eslint", "1.0.0", "eslint")
	require.NoError(t, os.MkdirAll(keep, 0o755))
	require.NoError(t, os.MkdirAll(stale, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Dir(keepShim), 0o755))
	require.NoError(t, os.WriteFile(keepShim, nil, 0o644))
	require.NoError(t, os.MkdirAll(filepath.Dir(staleShim), 0o755))
	require.NoError(t, os.WriteFile(staleShim, nil, 0o644))

	keepPrefixes := map[string]bool{
		download.InstallsBase(root, "tools", "actionlint"):  true,
		filepath.Join(root, "shims", "tools", "actionlint"): true,
	}
	require.NoError(t, pruneUnused(root, keepPrefixes))

	assert.DirExists(t, keep)
	assert.NoDirExists(t, filepath.Dir(filepath.Dir(stale))) // installs/tools/eslint gone entirely
	assert.DirExists(t, filepath.Dir(filepath.Dir(keepShim)))
	assert.NoDirExists(t, filepath.Dir(filepath.Dir(staleShim)))
}

func TestCacheClean(t *testing.T) {
	cacheDir := t.TempDir()
	root, err := download.Root(cacheDir)
	require.NoError(t, err)
	marker := filepath.Join(root, "installs", "tools", "whatever", "1.0.0", "marker")
	require.NoError(t, os.MkdirAll(filepath.Dir(marker), 0o755))
	require.NoError(t, os.WriteFile(marker, nil, 0o644))

	_, stderr, err := run2(t, "--config", trunkYAML, "--cache-dir", cacheDir, "cache", "clean")
	require.NoError(t, err, "stderr: %s", stderr)
	assert.NoDirExists(t, root)
}

func TestCachePrune_KeepsEnabledUsed(t *testing.T) {
	cacheDir := t.TempDir()
	root, err := download.Root(cacheDir)
	require.NoError(t, err)

	// actionlint is enabled+used by trunkYAML's fixture; eslint is not (see cli_test.go's
	// TestConfigPrint/TestConfigPrint_All comments describing the same fixture).
	keep := download.InstallDir(root, "tools", "actionlint", "1.0.0")
	stale := download.InstallDir(root, "tools", "eslint", "1.0.0")
	require.NoError(t, os.MkdirAll(keep, 0o755))
	require.NoError(t, os.MkdirAll(stale, 0o755))

	// Seed a shim for each item too -- cachePruneCmd.Run's own keep-map construction (not just
	// pruneUnused's glob/removal logic, already covered by TestPruneUnused) must preserve a kept
	// item's shim as well as its install dir.
	keepShim := download.ShimPath(root, "tools", "actionlint", "1.0.0", "actionlint")
	staleShim := download.ShimPath(root, "tools", "eslint", "1.0.0", "eslint")
	require.NoError(t, os.MkdirAll(filepath.Dir(keepShim), 0o755))
	require.NoError(t, os.WriteFile(keepShim, nil, 0o644))
	require.NoError(t, os.MkdirAll(filepath.Dir(staleShim), 0o755))
	require.NoError(t, os.WriteFile(staleShim, nil, 0o644))

	_, stderr, err := run2(t, "--config", trunkYAML, "--cache-dir", cacheDir, "cache", "prune")
	require.NoError(t, err, "stderr: %s", stderr)
	assert.DirExists(t, keep)
	assert.NoDirExists(t, stale)
	assert.FileExists(t, keepShim)
	assert.NoFileExists(t, staleShim)
}
