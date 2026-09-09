package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestAction_Interactive checks the interactive field's two documented literal forms (bare
// `true`, or the string "optional") both decode without error.
func TestAction_Interactive(t *testing.T) {
	for _, snippet := range []string{"interactive: true", "interactive: optional"} {
		var a Action
		require.NoError(t, yaml.Unmarshal([]byte(snippet), &a))
		assert.NotEmpty(t, a.Interactive)
	}
}
