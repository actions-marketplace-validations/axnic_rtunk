package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/xunleii/rtunk/pkg/trunk/config"
)

// TestAction_Interactive checks the interactive field's two documented literal forms (bare
// `true`, or the string "optional") both decode without error.
func TestAction_Interactive(t *testing.T) {
	for _, snippet := range []string{"interactive: true", "interactive: optional"} {
		var a config.Action
		require.NoError(t, yaml.Unmarshal([]byte(snippet), &a))
		assert.NotEmpty(t, a.Interactive)
	}
}

// TestOSSpec_UnmarshalYAML checks both documented forms (ARCHITECTURE.md): a bare name decoding
// to a single key mapping to itself, and a map from trunk's vocabulary to upstream's own naming —
// modeled on the real shellcheck/plugin.yaml download entries, which use one of each.
func TestOSSpec_UnmarshalYAML(t *testing.T) {
	var d config.Download
	require.NoError(t, yaml.Unmarshal([]byte(`
name: shellcheck
downloads:
  - os: { linux: linux }
    cpu: { arm_64: aarch64, x86_64: x86_64 }
  - os: macos
    cpu: x86_64
`), &d))

	assert.Equal(t, config.OSSpec{"linux": "linux"}, d.Downloads[0].OS)
	assert.Equal(t, config.OSSpec{"arm_64": "aarch64", "x86_64": "x86_64"}, d.Downloads[0].CPU)
	assert.Equal(t, config.OSSpec{"macos": "macos"}, d.Downloads[1].OS)
	assert.Equal(t, config.OSSpec{"x86_64": "x86_64"}, d.Downloads[1].CPU)
}
