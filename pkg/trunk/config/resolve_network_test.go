//go:build network

package config

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResolve_OfficialPluginsRepo actually clones github.com/trunk-io/plugins@v1.11.0 (testdata/
// trunk.yaml's only source) and resolves against it. Gated behind the "network" build tag so
// `go test ./...` never touches the network; run with `go test -tags network ./...`.
func TestResolve_OfficialPluginsRepo(t *testing.T) {
	cacheDir := t.TempDir()
	cfg, err := Resolve("testdata/trunk.yaml", cacheDir)
	require.NoError(t, err)

	assert.NotEmpty(t, cfg.Runtimes.Definitions)
	assert.NotEmpty(t, cfg.Lint.Definitions)
	assert.NotEmpty(t, cfg.Actions.Definitions)
	assert.NotEmpty(t, cfg.Tools)

	// The pinned tag must be genuinely self-consistent: everything trunk.yaml enables must
	// resolve against what the tag actually defines.
	assert.NoError(t, cfg.Validate())

	// The fetch must have left only a cache of the parsed definitions behind — no raw git
	// checkout — and a second Resolve against the same cacheDir must reuse it and get the exact
	// same result.
	entries, err := os.ReadDir(cacheDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.True(t, strings.HasSuffix(entries[0].Name(), ".json"))

	cfg2, err := Resolve("testdata/trunk.yaml", cacheDir)
	require.NoError(t, err)
	assert.Equal(t, cfg, cfg2)
}
