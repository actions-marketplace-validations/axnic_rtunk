package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axnic/rtunk/pkg/cache/download"
	"github.com/axnic/rtunk/pkg/trunk/config"
)

func TestCacheClean(t *testing.T) {
	cacheDir := t.TempDir()
	root, err := download.Root(cacheDir)
	require.NoError(t, err)
	sharedRoot := filepath.Dir(root)

	for _, marker := range []string{
		filepath.Join(root, "installs", "tools", "whatever", "1.0.0", "marker"),
		filepath.Join(sharedRoot, "plugins", "checkouts", "abc", "marker"),
		filepath.Join(sharedRoot, "logs", "marker"),
	} {
		require.NoError(t, os.MkdirAll(filepath.Dir(marker), 0o755))
		require.NoError(t, os.WriteFile(marker, nil, 0o644))
	}
	unrelated := filepath.Join(cacheDir, "unrelated-file.txt")
	require.NoError(t, os.WriteFile(unrelated, nil, 0o644))

	stdout, stderr, err := run2(t, "--config", trunkYAML, "--cache-dir", cacheDir, "cache", "clean")
	require.NoError(t, err, "stderr: %s", stderr)
	assert.Equal(t, "removed downloads (0 B)\nremoved plugins (0 B)\nremoved logs (0 B)\n", stdout)
	assert.NoDirExists(t, root)
	assert.NoDirExists(t, filepath.Join(sharedRoot, "plugins"))
	assert.NoDirExists(t, filepath.Join(sharedRoot, "logs"))
	assert.FileExists(t, unrelated)

	stdout, _, err = run2(t, "--config", trunkYAML, "--cache-dir", cacheDir, "cache", "clean")
	require.NoError(t, err)
	assert.Equal(t, "The cache is already empty.\n", stdout)
}

func TestCleanModel_RowsTurnIntoChecks(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "f"), make([]byte, 2048), 0o644))
	rows := []*cleanRow{{name: "downloads", path: dir}}
	m := newCleanModel(rows)
	assert.Contains(t, m.View().Content, " downloads\n", "spinner while removing")

	size, err := removeMeasured(dir)
	_, cmd := m.Update(cleanDoneMsg{row: rows[0], size: size, err: err})
	assert.NotNil(t, cmd, "last row done: quit")
	assert.Contains(t, m.View().Content, "✔")
	assert.Contains(t, m.View().Content, "downloads  2.0 KB freed")
	assert.NoDirExists(t, dir)
}

func TestCachePrune_DropsEntryForGoneRepo(t *testing.T) {
	cacheDir := t.TempDir()
	require.NoError(t, download.RecordUsage(cacheDir, filepath.Join(t.TempDir(), "gone"), config.Config{
		Tools: map[string]config.Tool{"eslint": {KnownGoodVersion: "1.0.0"}},
	}))
	root, err := download.Root(cacheDir)
	require.NoError(t, err)
	installDir := download.InstallDir(root, "tools", "eslint", "1.0.0")
	require.NoError(t, os.MkdirAll(installDir, 0o755))

	_, stderr, err := run2(t, "--config", trunkYAML, "--cache-dir", cacheDir, "cache", "prune")
	require.NoError(t, err, "stderr: %s", stderr)
	assert.NoDirExists(t, installDir)
}

func TestCachePrune_KeepsEntryForLiveRepo(t *testing.T) {
	cacheDir := t.TempDir()
	liveRepo := t.TempDir()
	require.NoError(t, download.RecordUsage(cacheDir, liveRepo, config.Config{
		Tools: map[string]config.Tool{"eslint": {KnownGoodVersion: "1.0.0"}},
	}))
	root, err := download.Root(cacheDir)
	require.NoError(t, err)
	installDir := download.InstallDir(root, "tools", "eslint", "1.0.0")
	require.NoError(t, os.MkdirAll(installDir, 0o755))

	_, stderr, err := run2(t, "--config", trunkYAML, "--cache-dir", cacheDir, "cache", "prune")
	require.NoError(t, err, "stderr: %s", stderr)
	assert.DirExists(t, installDir)
}
