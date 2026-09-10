//go:build network

package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xunleii/rtunk/pkg/trunk/config"
)

// TestResolve_OfficialPluginsRepo actually clones github.com/trunk-io/plugins@v1.11.0 (testdata/
// trunk.yaml's only source) and resolves against it. Gated behind the "network" build tag so
// `go test ./...` never touches the network; run with `go test -tags network ./...`.
func TestResolve_OfficialPluginsRepo(t *testing.T) {
	cfg, err := config.Resolve("testdata/trunk.yaml", t.TempDir())
	require.NoError(t, err)

	assert.NotEmpty(t, cfg.Runtimes.Definitions)
	assert.NotEmpty(t, cfg.Lint.Definitions)
	assert.NotEmpty(t, cfg.Actions.Definitions)
	assert.NotEmpty(t, cfg.Tools)

	// The pinned tag must be genuinely self-consistent: everything trunk.yaml enables must
	// resolve against what the tag actually defines.
	assert.NoError(t, cfg.Validate())
}
