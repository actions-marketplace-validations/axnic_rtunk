package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unreachableSource never resolves (.invalid is reserved by RFC 2606 to never resolve), so any
// test using it that still succeeds proves the network was never actually touched.
var unreachableSource = PluginSource{ID: "unreachable", URI: "https://example.invalid/plugins", Ref: "v0.0.0"}

// TestFetchGitSource_CacheHit: a valid cache for src's uri+ref must be used as-is, with no git
// command run at all — proven here by pointing src at a uri that can never be reached.
func TestFetchGitSource_CacheHit(t *testing.T) {
	cacheDir := t.TempDir()
	want := sourceDefs{
		Downloads: map[string]Download{},
		Tools:     map[string]Tool{"foo": {Name: "foo", KnownGoodVersion: "1.0.0"}},
		Lint:      map[string]Linter{},
		Actions:   map[string]Action{},
		Runtimes:  map[string]Runtime{},
	}
	require.NoError(t, saveSourceCache(cacheFilePath(cacheDir, unreachableSource), want))

	defs, dupErrs, err := fetchGitSource(cacheDir, unreachableSource)

	require.NoError(t, err)
	assert.Empty(t, dupErrs)
	assert.Equal(t, want, defs)
}

// TestFetchGitSource_CorruptCache_Regenerates: a cache file that fails to decode must be dropped
// rather than trusted, and fetchGitSource must fall through to a real fetch — which, against an
// unreachable uri, surfaces as a *FetchError. Proves the stale-cache path without needing network
// to succeed, only to be attempted.
func TestFetchGitSource_CorruptCache_Regenerates(t *testing.T) {
	cacheDir := t.TempDir()
	cacheFile := cacheFilePath(cacheDir, unreachableSource)
	require.NoError(t, os.MkdirAll(filepath.Dir(cacheFile), 0o755))
	require.NoError(t, os.WriteFile(cacheFile, []byte("not valid json"), 0o644))

	_, _, err := fetchGitSource(cacheDir, unreachableSource)

	var fetchErr *FetchError
	require.ErrorAs(t, err, &fetchErr)
	assert.Equal(t, "unreachable", fetchErr.SourceID)
	_, statErr := os.Stat(cacheFile)
	assert.True(t, os.IsNotExist(statErr), "corrupt cache file must be removed, not left behind")
}
