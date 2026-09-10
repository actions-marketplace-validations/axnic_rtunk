package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// trunkYAML is the same fixture pkg/trunk/config's own tests resolve against: one local plugin
// source (testdata/pluginrepo) defining eslint/prettier/actionlint/shellcheck/node/commitlint,
// with lint [actionlint, prettier], runtimes [node], and actions [commitlint] enabled.
const trunkYAML = "../../pkg/trunk/config/testdata/trunk-with-plugins.yaml"

func run2(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var out, errOut bytes.Buffer
	err = run(args, &out, &errOut)
	return out.String(), errOut.String(), err
}

func TestConfigList(t *testing.T) {
	tests := []struct {
		args []string
		want []string
	}{
		// eslint/shellcheck's plugin.yaml only define a tool, not a lint: block (see
		// testdata/pluginrepo/linters/{eslint,shellcheck}/plugin.yaml) — actionlint/prettier are
		// the only two actual linter definitions in the fixture, and both happen to be enabled.
		{[]string{"config", "lint", "list"}, []string{"actionlint", "prettier"}},
		{[]string{"config", "lint", "list", "--enabled"}, []string{"actionlint", "prettier"}},
		// config.Resolve itself now trims to enabled+used (pkg/trunk/config's filterEnabled), so
		// eslint/shellcheck (unreferenced tools) and go-mod-tidy (a defined but never-enabled
		// action) are gone before the CLI even sees them — list and list --enabled agree here.
		{[]string{"config", "tools", "list"}, []string{"actionlint", "prettier"}},
		{[]string{"config", "tools", "list", "--enabled"}, []string{"actionlint", "prettier"}},
		{[]string{"config", "plugins", "list"}, []string{"trunk"}},
		{[]string{"config", "plugins", "list", "--enabled"}, []string{"trunk"}},
		{[]string{"config", "actions", "list"}, []string{"commitlint"}},
		{[]string{"config", "actions", "list", "--enabled"}, []string{"commitlint"}},
		{[]string{"config", "runtimes", "list"}, []string{"node"}},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			stdout, stderr, err := run2(t, append([]string{"--config", trunkYAML}, tt.args...)...)
			require.NoError(t, err, "stderr: %s", stderr)
			assert.Equal(t, tt.want, strings.Fields(stdout))
		})
	}
}

func TestConfigShow(t *testing.T) {
	stdout, stderr, err := run2(t, "--config", trunkYAML, "config", "lint", "show", "actionlint")
	require.NoError(t, err, "stderr: %s", stderr)
	assert.Contains(t, stdout, "name: actionlint")
	assert.Contains(t, stdout, "tools:")

	_, _, err = run2(t, "--config", trunkYAML, "config", "lint", "show", "does-not-exist")
	assert.Error(t, err)
}

func TestConfigShow_JSON(t *testing.T) {
	stdout, stderr, err := run2(t, "--config", trunkYAML, "config", "runtimes", "show", "node", "--output", "json")
	require.NoError(t, err, "stderr: %s", stderr)
	assert.Contains(t, stdout, `"type": "node"`)
}

func TestConfigPrint(t *testing.T) {
	stdout, stderr, err := run2(t, "--config", trunkYAML, "config", "print")
	require.NoError(t, err, "stderr: %s", stderr)
	assert.Contains(t, stdout, "actionlint")
	assert.Contains(t, stdout, "node")
	assert.Contains(t, stdout, "commitlint")
}

func TestFindTrunkYAML(t *testing.T) {
	// EvalSymlinks: on macOS, t.TempDir() lives under /var, a symlink to /private/var, and
	// os.Getwd() (which findTrunkYAML calls) returns the resolved physical path — normalize here
	// so the two sides of the comparison below agree.
	root, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".trunk"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".trunk", "trunk.yaml"), []byte("version: \"0.1\"\n"), 0o644))

	sub := filepath.Join(root, "a", "b")
	require.NoError(t, os.MkdirAll(sub, 0o755))

	cwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.Chdir(cwd)) })
	require.NoError(t, os.Chdir(sub))

	found, err := findTrunkYAML()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(root, ".trunk", "trunk.yaml"), found)
}

func TestFindTrunkYAML_NotFound(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.Chdir(cwd)) })
	require.NoError(t, os.Chdir(t.TempDir()))

	_, err = findTrunkYAML()
	assert.Error(t, err)
}

func TestUnknownCommand(t *testing.T) {
	_, _, err := run2(t, "bogus")
	assert.Error(t, err)
}
