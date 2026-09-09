package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	cfg, err := Resolve("testdata/trunk.yaml")
	require.NoError(t, err)

	assert.Equal(t, "0.1", cfg.Version)
	assert.Equal(t, "1.25.0", cfg.CLI.Version)
	assert.Equal(t, map[string]PluginSource{
		"trunk": {ID: "trunk", URI: "https://github.com/trunk-io/plugins", Ref: "v1.11.0"},
	}, cfg.Plugins.Sources)
	assert.Equal(t, []string{"node@22.16.0", "python@3.14.4"}, cfg.Runtimes.Enabled)
	assert.Equal(t, []string{"checkov@3.3.16", "git-diff-check"}, cfg.Lint.Enabled)
	assert.Equal(t, []string{"commitlint", "trunk-check-pre-push"}, cfg.Actions.Enabled)

	// "trunk" is a git source (no network, never fetched): nothing to merge, and — since
	// resolution is incomplete — the enabled lists aren't checked against (empty) Definitions.
	assert.Empty(t, cfg.Runtimes.Definitions)
	assert.Empty(t, cfg.Lint.Definitions)
	assert.Empty(t, cfg.Actions.Definitions)
	assert.Empty(t, cfg.Tools)
	assert.Empty(t, cfg.Downloads)
}

func TestResolve_MissingFile(t *testing.T) {
	_, err := Resolve("testdata/does-not-exist.yaml")

	var readErr *ReadError
	require.ErrorAs(t, err, &readErr)
	assert.Equal(t, "testdata/does-not-exist.yaml", readErr.Path)
}

func TestResolve_InvalidYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trunk.yaml")
	require.NoError(t, os.WriteFile(path, []byte("version: [not-a-mapping\n"), 0o644))

	_, err := Resolve(path)

	var parseErr *ParseError
	require.ErrorAs(t, err, &parseErr)
	assert.Equal(t, path, parseErr.Path)
}

// TestResolve_WithPluginRepo resolves a real trunk.yaml against a local plugin repository built
// from the excerpts documented in ARCHITECTURE.md (github.com/trunk-io/plugins), proving Resolve
// merges definitions from every category dir and that everything it enables/references checks
// out: this fixture is a fully valid, self-consistent config, so Resolve must return no error.
func TestResolve_WithPluginRepo(t *testing.T) {
	cfg, err := Resolve("testdata/trunk-with-plugins.yaml")
	require.NoError(t, err)

	require.Contains(t, cfg.Downloads, "shellcheck")
	assert.Equal(t, OSSpec{"macos": "macos"}, cfg.Downloads["shellcheck"].Downloads[1].OS)
	assert.Equal(t, OSSpec{"linux": "linux"}, cfg.Downloads["shellcheck"].Downloads[0].OS)
	assert.Equal(t, OSSpec{"arm_64": "aarch64", "x86_64": "x86_64"}, cfg.Downloads["shellcheck"].Downloads[0].CPU)

	assert.Contains(t, cfg.Tools, "eslint")
	assert.Contains(t, cfg.Tools, "shellcheck")
	assert.Contains(t, cfg.Tools, "actionlint")
	assert.Contains(t, cfg.Tools, "prettier")

	require.Contains(t, cfg.Lint.Definitions, "actionlint")
	require.Contains(t, cfg.Lint.Definitions, "prettier")

	require.Contains(t, cfg.Actions.Definitions, "commitlint")
	require.Contains(t, cfg.Actions.Definitions, "go-mod-tidy")

	require.Contains(t, cfg.Runtimes.Definitions, "node")
	assert.Equal(t, []string{"node", "npm", "npx", "corepack"}, cfg.Runtimes.Definitions["node"].Shims)
}

// TestResolve_DuplicateResource: two plugin.yaml files under the same local source both define a
// tool named "foo". Resolve must report a *DuplicateError, and the later file (foo-b, sorted
// after foo-a) must win the overwrite.
func TestResolve_DuplicateResource(t *testing.T) {
	cfg, err := Resolve("testdata/trunk-duplicate.yaml")

	var dupErr *DuplicateError
	require.ErrorAs(t, err, &dupErr)
	assert.Equal(t, "tool", dupErr.Category)
	assert.Equal(t, "foo", dupErr.Key)

	require.Contains(t, cfg.Tools, "foo")
	assert.Equal(t, "2.0.0", cfg.Tools["foo"].KnownGoodVersion, "later source must win the whole key, not merge fields")
}

// TestResolve_DanglingReference: a linter names a tool that no plugin.yaml defines. Resolve
// itself reads it in without complaint; Validate must report a *ReferenceError identifying
// exactly what's missing.
func TestResolve_DanglingReference(t *testing.T) {
	cfg, err := Resolve("testdata/trunk-dangling.yaml")
	require.NoError(t, err)

	err = cfg.Validate()

	var refErr *ReferenceError
	require.ErrorAs(t, err, &refErr)
	assert.Equal(t, "lint", refErr.Category)
	assert.Equal(t, "orphan", refErr.Key)
	assert.Equal(t, "tools", refErr.Field)
	assert.Equal(t, "nonexistent-tool", refErr.Reference)
}

// TestResolve_SourceNotFound: a `local:` plugin source pointing nowhere is a config error, unlike
// a git source (uri/ref), which Resolve intentionally never fetches.
func TestResolve_SourceNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trunk.yaml")
	require.NoError(t, os.WriteFile(path, []byte("plugins:\n  sources:\n    - id: gone\n      local: ./nowhere\n"), 0o644))

	_, err := Resolve(path)

	var notFoundErr *SourceNotFoundError
	require.ErrorAs(t, err, &notFoundErr)
	assert.Equal(t, "gone", notFoundErr.SourceID)
}
