package config_test

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axnic/rtunk/pkg/trunk/config"
)

// TestReadError_Unwrap and TestParseError_Unwrap: both typed errors must unwrap to the underlying
// os/yaml error, so callers can errors.Is/As against it (e.g. os.ErrNotExist).
func TestReadError_Unwrap(t *testing.T) {
	_, err := config.Resolve("testdata/does-not-exist.yaml", t.TempDir())
	assert.True(t, errors.Is(err, os.ErrNotExist))
}
