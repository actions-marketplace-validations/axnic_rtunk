package download_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xunleii/rtunk/pkg/trunk/config"
	"github.com/xunleii/rtunk/pkg/trunk/download"
)

func tarGzBytes(t *testing.T, topDir, name, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: topDir + "/" + name, Mode: 0o755, Size: int64(len(content))}))
	_, err := tw.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	return buf.Bytes()
}

func TestDownload_Runtime(t *testing.T) {
	archive := tarGzBytes(t, "tool-1.0.0", "shellcheck", "#!/bin/sh\n")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archive)
	}))
	defer srv.Close()

	cfg := config.Config{
		Downloads: map[string]config.Download{
			"shellcheck": {Downloads: []config.DownloadEntry{{
				OS:  config.OSSpec{"linux": "linux", "macos": "macos", "windows": "windows"},
				CPU: config.OSSpec{"x86_64": "x86_64", "arm_64": "arm_64"},
				URL: srv.URL + "/shellcheck.tar.gz", StripComponents: 1,
			}}},
		},
		Tools: map[string]config.Tool{},
		Runtimes: config.CategoryConfig[config.Runtime]{
			Definitions: map[string]config.Runtime{
				"shellcheck": {Type: "shellcheck", Download: "shellcheck", KnownGoodVersion: "1.0.0", Shims: []string{"shellcheck"}},
			},
		},
	}

	cacheDir := t.TempDir()
	events, err := download.Download(cfg, cacheDir, download.Ref{Category: "runtimes", ID: "shellcheck"})
	require.NoError(t, err)

	var phases []download.Phase
	for ev := range events {
		require.NoError(t, ev.Err, "event: %+v", ev)
		phases = append(phases, ev.Phase)
	}
	assert.Contains(t, phases, download.Started)
	assert.Contains(t, phases, download.Done)

	root, err := download.Root(cacheDir)
	require.NoError(t, err)
	shimPath := download.ShimPath(root, "runtimes", "shellcheck", "1.0.0", "shellcheck")
	assert.FileExists(t, shimPath)
}

func TestDownload_Runtime_AlreadyCached(t *testing.T) {
	cfg := config.Config{
		Downloads: map[string]config.Download{"shellcheck": {}},
		Runtimes: config.CategoryConfig[config.Runtime]{
			Definitions: map[string]config.Runtime{
				"shellcheck": {Type: "shellcheck", Download: "shellcheck", KnownGoodVersion: "1.0.0"},
			},
		},
	}
	cacheDir := t.TempDir()
	root, err := download.Root(cacheDir)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(download.InstallDir(root, "runtimes", "shellcheck", "1.0.0"), 0o755))

	events, err := download.Download(cfg, cacheDir, download.Ref{Category: "runtimes", ID: "shellcheck"})
	require.NoError(t, err)
	var phases []download.Phase
	for ev := range events {
		require.NoError(t, ev.Err)
		phases = append(phases, ev.Phase)
	}
	assert.Equal(t, []download.Phase{download.Cached}, phases, "an already-installed version must not be re-fetched")
}

func TestDownload_Runtime_SystemVersion(t *testing.T) {
	cfg := config.Config{
		Runtimes: config.CategoryConfig[config.Runtime]{
			Definitions: map[string]config.Runtime{"php": {Type: "php", SystemVersion: ">=8.0.0"}},
		},
	}
	events, err := download.Download(cfg, t.TempDir(), download.Ref{Category: "runtimes", ID: "php"})
	require.NoError(t, err)
	var phases []download.Phase
	for ev := range events {
		require.NoError(t, ev.Err)
		phases = append(phases, ev.Phase)
	}
	assert.Equal(t, []download.Phase{download.Cached}, phases, "a system_version runtime is never downloaded")
}

func TestDownload_UnknownRef(t *testing.T) {
	events, err := download.Download(config.Config{}, t.TempDir(), download.Ref{Category: "runtimes", ID: "nope"})
	require.NoError(t, err)
	ev := <-events
	assert.Equal(t, download.Failed, ev.Phase)
	assert.Error(t, ev.Err)
}

func TestDownload_LintRef_ExpandsToTools(t *testing.T) {
	archive := tarGzBytes(t, "tool-1.0.0", "actionlint", "#!/bin/sh\n")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archive)
	}))
	defer srv.Close()

	cfg := config.Config{
		Downloads: map[string]config.Download{
			"actionlint": {Downloads: []config.DownloadEntry{{
				OS: config.OSSpec{"linux": "linux", "macos": "macos", "windows": "windows"},
				CPU: config.OSSpec{"x86_64": "x86_64", "arm_64": "arm_64"},
				URL: srv.URL + "/actionlint.tar.gz", StripComponents: 1,
			}}},
		},
		Tools: map[string]config.Tool{
			"actionlint": {Name: "actionlint", Download: "actionlint", KnownGoodVersion: "1.0.0", Shims: []string{"actionlint"}},
		},
		Lint: config.LintConfig{
			CategoryConfig: config.CategoryConfig[config.Linter]{
				Definitions: map[string]config.Linter{"actionlint": {Name: "actionlint", Tools: []string{"actionlint"}}},
			},
		},
	}

	events, err := download.Download(cfg, t.TempDir(), download.Ref{Category: "lint", ID: "actionlint"})
	require.NoError(t, err)
	var sawToolDone bool
	for ev := range events {
		require.NoError(t, ev.Err, "event: %+v", ev)
		if ev.Ref.Category == "tools" && ev.Ref.ID == "actionlint" && ev.Phase == download.Done {
			sawToolDone = true
		}
	}
	assert.True(t, sawToolDone, "a lint ref must expand into fetching its underlying tool(s)")
}

func TestDownload_PluginsRef_AlwaysCached(t *testing.T) {
	cfg := config.Config{Plugins: struct {
		Sources map[string]config.PluginSource
	}{Sources: map[string]config.PluginSource{"trunk": {ID: "trunk"}}}}

	events, err := download.Download(cfg, t.TempDir(), download.Ref{Category: "plugins", ID: "trunk"})
	require.NoError(t, err)
	ev := <-events
	assert.Equal(t, download.Cached, ev.Phase, "resolving cfg already fetched every plugin source it references")
}
