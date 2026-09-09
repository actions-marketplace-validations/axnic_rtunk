package config

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestReadError_Unwrap and TestParseError_Unwrap: both typed errors must unwrap to the underlying
// os/yaml error, so callers can errors.Is/As against it (e.g. os.ErrNotExist).
func TestReadError_Unwrap(t *testing.T) {
	_, err := Resolve("testdata/does-not-exist.yaml")
	assert.True(t, errors.Is(err, os.ErrNotExist))
}
