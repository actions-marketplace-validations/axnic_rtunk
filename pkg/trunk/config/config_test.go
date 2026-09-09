package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestResolve(t *testing.T) {
	cfg, err := Resolve("testdata/trunk.yaml")
	require.NoError(t, err)

	assert.Equal(t, "0.1", cfg.Version)
	assert.Equal(t, "1.25.0", cfg.CLI.Version)
	assert.Equal(t, []PluginSource{
		{ID: "trunk", URI: "https://github.com/trunk-io/plugins", Ref: "v1.11.0"},
		{ID: "local-plugins", Local: "../plugins"},
	}, cfg.Plugins.Sources)
	assert.Equal(t, []string{"node@22.16.0", "python@3.14.4"}, cfg.Runtimes.Enabled)
	assert.Equal(t, []string{"checkov@3.3.16", "git-diff-check"}, cfg.Lint.Enabled)
	assert.Equal(t, []string{"commitlint", "trunk-check-pre-push"}, cfg.Actions.Enabled)

	// "../plugins" doesn't exist and "trunk" is a git source (no network): nothing to resolve.
	assert.Empty(t, cfg.Runtimes.Definitions)
	assert.Empty(t, cfg.Lint.Definitions)
	assert.Empty(t, cfg.Actions.Definitions)
	assert.Empty(t, cfg.Tools)
	assert.Empty(t, cfg.Downloads)
}

func TestResolve_MissingFile(t *testing.T) {
	_, err := Resolve("testdata/does-not-exist.yaml")
	assert.Error(t, err)
}

func TestResolve_InvalidYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trunk.yaml")
	require.NoError(t, os.WriteFile(path, []byte("version: [not-a-mapping\n"), 0o644))

	_, err := Resolve(path)
	assert.Error(t, err)
}

// TestResolve_WithPluginRepo resolves a real trunk.yaml against a local plugin repository built
// from the excerpts documented in ARCHITECTURE.md (github.com/trunk-io/plugins), proving Resolve
// merges definitions from every category dir (linters/, actions/, runtimes/) into the config.
func TestResolve_WithPluginRepo(t *testing.T) {
	cfg, err := Resolve("testdata/trunk-with-plugins.yaml")
	require.NoError(t, err)

	require.Len(t, cfg.Downloads, 1)
	assert.Equal(t, "shellcheck", cfg.Downloads[0].Name)
	assert.Equal(t, OSSpec{"macos": "macos"}, cfg.Downloads[0].Downloads[1].OS)
	assert.Equal(t, OSSpec{"linux": "linux"}, cfg.Downloads[0].Downloads[0].OS)
	assert.Equal(t, OSSpec{"arm_64": "aarch64", "x86_64": "x86_64"}, cfg.Downloads[0].Downloads[0].CPU)

	require.Len(t, cfg.Tools, 2)
	names := []string{cfg.Tools[0].Name, cfg.Tools[1].Name}
	assert.ElementsMatch(t, []string{"eslint", "shellcheck"}, names)

	require.Len(t, cfg.Lint.Definitions, 2)
	linterNames := []string{cfg.Lint.Definitions[0].Name, cfg.Lint.Definitions[1].Name}
	assert.ElementsMatch(t, []string{"actionlint", "prettier"}, linterNames)

	require.Len(t, cfg.Actions.Definitions, 2)
	actionIDs := []string{cfg.Actions.Definitions[0].ID, cfg.Actions.Definitions[1].ID}
	assert.ElementsMatch(t, []string{"commitlint", "go-mod-tidy"}, actionIDs)

	require.Len(t, cfg.Runtimes.Definitions, 1)
	assert.Equal(t, "node", cfg.Runtimes.Definitions[0].Type)
	assert.Equal(t, []string{"node", "npm", "npx", "corepack"}, cfg.Runtimes.Definitions[0].Shims)
}

// TestAction_Interactive checks the interactive field's two documented literal forms (bare
// `true`, or the string "optional") both decode without error.
func TestAction_Interactive(t *testing.T) {
	for _, snippet := range []string{"interactive: true", "interactive: optional"} {
		var a Action
		require.NoError(t, yaml.Unmarshal([]byte(snippet), &a))
		assert.NotEmpty(t, a.Interactive)
	}
}
