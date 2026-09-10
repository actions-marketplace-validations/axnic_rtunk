package download_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xunleii/rtunk/pkg/trunk/config"
	"github.com/xunleii/rtunk/pkg/trunk/download"
)

func TestMatchEntry(t *testing.T) {
	entries := []config.DownloadEntry{
		{
			OS:  config.OSSpec{"linux": "linux"},
			CPU: config.OSSpec{"arm_64": "aarch64", "x86_64": "x86_64"},
			URL: "linux-url",
		},
		{
			OS:  config.OSSpec{"macos": "macos"},
			CPU: config.OSSpec{"x86_64": "x86_64"},
			URL: "macos-url",
		},
	}

	entry, osVal, cpuVal, ok := download.MatchEntry(entries, "linux", "arm64")
	require.True(t, ok)
	assert.Equal(t, "linux-url", entry.URL)
	assert.Equal(t, "linux", osVal)
	assert.Equal(t, "aarch64", cpuVal)

	_, _, _, ok = download.MatchEntry(entries, "windows", "amd64")
	assert.False(t, ok, "no windows entry declared")

	_, _, _, ok = download.MatchEntry(entries, "macos", "arm64")
	assert.False(t, ok, "macos entry only declares x86_64")
}
