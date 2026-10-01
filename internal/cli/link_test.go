package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func isSymlink(t *testing.T, path string) bool {
	t.Helper()
	info, err := os.Lstat(path)
	require.NoError(t, err)
	return info.Mode()&os.ModeSymlink != 0
}

func TestLinkCmd_CreatesLogsAndPerToolSymlinks(t *testing.T) {
	cfgPath, repoRoot := writeToolLinterFixture(t, []string{"fixture"})

	_, stderr, err := run2(t, "--config", cfgPath, "toolbox", "link")
	require.NoError(t, err, "stderr: %s", stderr)

	rtunkDir := filepath.Join(repoRoot, ".rtunk")
	assert.True(t, isSymlink(t, filepath.Join(rtunkDir, "logs")))
	// .rtunk/tools/<id> links directly to the tool's shim -- not a single symlink to the whole
	// (cross-repo-shared) shims tree -- so `.rtunk/tools/fixture` must exist.
	assert.True(t, isSymlink(t, filepath.Join(rtunkDir, "tools", "fixture")))

	gitignore, err := os.ReadFile(filepath.Join(rtunkDir, ".gitignore"))
	require.NoError(t, err)
	assert.Contains(t, string(gitignore), "logs")
	assert.Contains(t, string(gitignore), "tools")
	assert.Contains(t, string(gitignore), "plugins")
	assert.Contains(t, string(gitignore), "rtunk.local.yaml\n")
}

func TestWriteLinkGitignore_CompletesAnExistingFileWithoutRewritingIt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")
	require.NoError(t, os.WriteFile(path, []byte("logs\nmine\ntools"), 0o644)) // no trailing newline

	require.NoError(t, writeLinkGitignore(dir))
	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "logs\nmine\ntools\nplugins\nuser_trunk.yaml\nuser.yaml\nrtunk.local.yaml\n", string(got))

	require.NoError(t, writeLinkGitignore(dir))
	again, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, string(got), string(again), "idempotent")
}

// TestLinkCmd_LocalPluginSource_NoPluginsSymlink: a `local:` plugin source is already a real
// directory in the repo -- link must not create a dangling .rtunk/plugins/<id> symlink for it.
func TestLinkCmd_LocalPluginSource_NoPluginsSymlink(t *testing.T) {
	cfgPath, repoRoot := writeToolLinterFixture(t, nil)

	_, stderr, err := run2(t, "--config", cfgPath, "toolbox", "link")
	require.NoError(t, err, "stderr: %s", stderr)

	entries, err := os.ReadDir(filepath.Join(repoRoot, ".rtunk", "plugins"))
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestLinkCmd_Idempotent(t *testing.T) {
	cfgPath, repoRoot := writeToolLinterFixture(t, nil)

	_, stderr, err := run2(t, "--config", cfgPath, "toolbox", "link")
	require.NoError(t, err, "stderr: %s", stderr)
	target, err := os.Readlink(filepath.Join(repoRoot, ".rtunk", "logs"))
	require.NoError(t, err)

	_, stderr, err = run2(t, "--config", cfgPath, "toolbox", "link")
	require.NoError(t, err, "stderr: %s", stderr)
	target2, err := os.Readlink(filepath.Join(repoRoot, ".rtunk", "logs"))
	require.NoError(t, err)

	assert.Equal(t, target, target2)
}

// TestInitCmd_AlsoLinksDotRtunk: a fresh `rtunk init` should leave .rtunk ready to use, not just
// the scaffolded rtunk.yaml -- this repo's own link logic runs as part of init, not a separate
// manual step the user has to remember.
func TestInitCmd_AlsoLinksDotRtunk(t *testing.T) {
	repoRoot, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	require.NoError(t, exec.Command("git", "-C", repoRoot, "init", "-q").Run())

	cwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.Chdir(cwd)) })
	require.NoError(t, os.Chdir(repoRoot))

	stdout, stderr, err := run2(t, "init")
	require.NoError(t, err, "stdout: %s stderr: %s", stdout, stderr)

	assert.True(t, isSymlink(t, filepath.Join(repoRoot, ".rtunk", "logs")))
	// A fresh init scaffold enables no linters, so there's nothing to put under tools/ yet -- the
	// directory itself existing is all init can promise.
	assert.DirExists(t, filepath.Join(repoRoot, ".rtunk", "tools"))
}
