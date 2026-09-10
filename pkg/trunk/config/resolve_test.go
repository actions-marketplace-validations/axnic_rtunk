package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xunleii/rtunk/pkg/trunk/config"
)

// TestResolve_InvalidSource: a plugins.sources entry has neither local: nor uri: set. Resolve
// must report an *InvalidSourceError rather than trying (and failing confusingly) to fetch it as
// a git source.
func TestResolve_InvalidSource(t *testing.T) {
	_, err := config.Resolve("testdata/trunk-invalid-source.yaml", t.TempDir())

	var invalidErr *config.InvalidSourceError
	require.ErrorAs(t, err, &invalidErr)
	assert.Equal(t, "neither", invalidErr.SourceID)
}

func TestResolve_MissingFile(t *testing.T) {
	_, err := config.Resolve("testdata/does-not-exist.yaml", t.TempDir())

	var readErr *config.ReadError
	require.ErrorAs(t, err, &readErr)
	assert.Equal(t, "testdata/does-not-exist.yaml", readErr.Path)
}

func TestResolve_InvalidYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trunk.yaml")
	require.NoError(t, os.WriteFile(path, []byte("version: [not-a-mapping\n"), 0o644))

	_, err := config.Resolve(path, t.TempDir())

	var parseErr *config.ParseError
	require.ErrorAs(t, err, &parseErr)
	assert.Equal(t, path, parseErr.Path)
}

// TestResolve_WithPluginRepo resolves a real trunk.yaml against a local plugin repository built
// from the excerpts documented in ARCHITECTURE.md (github.com/trunk-io/plugins), proving Resolve
// merges definitions from every category dir and that everything it enables/references checks
// out: this fixture is a fully valid, self-consistent config, so Resolve must return no error.
func TestResolve_WithPluginRepo(t *testing.T) {
	cfg, err := config.Resolve("testdata/trunk-with-plugins.yaml", t.TempDir())
	require.NoError(t, err)

	require.Contains(t, cfg.Downloads, "shellcheck")
	assert.Equal(t, config.OSSpec{"macos": "macos"}, cfg.Downloads["shellcheck"].Downloads[1].OS)
	assert.Equal(t, config.OSSpec{"linux": "linux"}, cfg.Downloads["shellcheck"].Downloads[0].OS)
	assert.Equal(t, config.OSSpec{"arm_64": "aarch64", "x86_64": "x86_64"}, cfg.Downloads["shellcheck"].Downloads[0].CPU)

	assert.Contains(t, cfg.Tools, "eslint")
	assert.Contains(t, cfg.Tools, "shellcheck")
	assert.Contains(t, cfg.Tools, "actionlint")
	assert.Contains(t, cfg.Tools, "prettier")

	require.Contains(t, cfg.Lint.Definitions, "actionlint")
	require.Contains(t, cfg.Lint.Definitions, "prettier")

	require.Contains(t, cfg.Actions.Definitions, "commitlint")
	require.Contains(t, cfg.Actions.Definitions, "go-mod-tidy")

	require.Contains(t, cfg.Runtimes.Definitions, "node")
	assert.Equal(t, config.ShimList{"node", "npm", "npx", "corepack"}, cfg.Runtimes.Definitions["node"].Shims)
}

// TestResolve_DuplicateResource: two plugin.yaml files under the same local source both define a
// tool named "foo". Resolve must report a *DuplicateError, and the later file (foo-b, sorted
// after foo-a) must win the overwrite.
func TestResolve_DuplicateResource(t *testing.T) {
	cfg, err := config.Resolve("testdata/trunk-duplicate.yaml", t.TempDir())

	var dupErr *config.DuplicateError
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
	cfg, err := config.Resolve("testdata/trunk-dangling.yaml", t.TempDir())
	require.NoError(t, err)

	err = cfg.Validate()

	var refErr *config.ReferenceError
	require.ErrorAs(t, err, &refErr)
	assert.Equal(t, "lint", refErr.Category)
	assert.Equal(t, "orphan", refErr.Key)
	assert.Equal(t, "tools", refErr.Field)
	assert.Equal(t, "nonexistent-tool", refErr.Reference)
}

// TestResolve_SourceNotFound: a `local:` plugin source pointing nowhere is a config error — unlike
// a git source, a local source is expected to already be present, never fetched.
func TestResolve_SourceNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trunk.yaml")
	require.NoError(t, os.WriteFile(path, []byte("plugins:\n  sources:\n    - id: gone\n      local: ./nowhere\n"), 0o644))

	_, err := config.Resolve(path, t.TempDir())

	var notFoundErr *config.SourceNotFoundError
	require.ErrorAs(t, err, &notFoundErr)
	assert.Equal(t, "gone", notFoundErr.SourceID)
}
