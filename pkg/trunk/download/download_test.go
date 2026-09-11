package download_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"runtime"
	"sync/atomic"
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

// nodeToolConfig builds a runtime+package "tools" config: a "node" runtime fetched from srv (a
// tar.gz containing a stub bin/npm) and an "eslint" tool installed through it -- the shape
// TestDownload_ToolRuntimePackage_* tests exercise end to end.
func nodeToolConfig(nodeArchiveURL string) config.Config {
	return config.Config{
		Downloads: map[string]config.Download{
			"node": {Downloads: []config.DownloadEntry{{
				OS:  config.OSSpec{"linux": "linux", "macos": "macos", "windows": "windows"},
				CPU: config.OSSpec{"x86_64": "x86_64", "arm_64": "arm_64"},
				URL: nodeArchiveURL, StripComponents: 1,
			}}},
		},
		Tools: map[string]config.Tool{
			"eslint": {Name: "eslint", Runtime: "node", Package: "eslint", KnownGoodVersion: "8.10.0", Shims: []string{"eslint"}},
		},
		Runtimes: config.CategoryConfig[config.Runtime]{
			Definitions: map[string]config.Runtime{
				"node": {Type: "node", Download: "node", KnownGoodVersion: "18.0.0"},
			},
		},
	}
}

// TestDownload_ToolRuntimePackage_NodeModulesBin drives a runtime:+package: tool ("eslint" through
// a "node" runtime) through the real Download() end to end, with a stub npm that lays its
// installed binary out at node_modules/.bin/<name> the way real npm does. This is the layout
// installNodePackage's `npm install --prefix` actually produces -- shimSearchPaths must know to
// look there, or FindShimTarget always fails for every runtime+package tool (the bug this test
// pins down).
func TestDownload_ToolRuntimePackage_NodeModulesBin(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("npm stub is a POSIX shell script")
	}

	npmScript := `#!/bin/sh
mkdir -p "$3/node_modules/.bin"
cat > "$3/node_modules/.bin/eslint" <<'EOS'
#!/bin/sh
echo ran-eslint
EOS
chmod +x "$3/node_modules/.bin/eslint"
`
	archive := tarGzBytes(t, "node-18.0.0", "bin/npm", npmScript)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archive)
	}))
	defer srv.Close()

	cfg := nodeToolConfig(srv.URL + "/node.tar.gz")
	cacheDir := t.TempDir()
	events, err := download.Download(cfg, cacheDir, download.Ref{Category: "tools", ID: "eslint"})
	require.NoError(t, err)

	var phases []download.Phase
	for ev := range events {
		require.NoError(t, ev.Err, "event: %+v", ev)
		phases = append(phases, ev.Phase)
	}
	assert.Contains(t, phases, download.Done)

	root, err := download.Root(cacheDir)
	require.NoError(t, err)
	shimPath := download.ShimPath(root, "tools", "eslint", "8.10.0", "eslint")
	require.FileExists(t, shimPath)

	out, err := exec.Command(shimPath).CombinedOutput()
	require.NoError(t, err, "shim output: %s", out)
	assert.Equal(t, "ran-eslint\n", string(out))
}

// TestDownload_ToolRuntimePackage_RuntimeFetchFailure pins down that a runtime+package tool's
// Failed event carries the runtime's own real failure (a bad download), not the confusing
// downstream symptom of proceeding to InstallPackage anyway with no runtime on disk ("npm not
// found").
func TestDownload_ToolRuntimePackage_RuntimeFetchFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := nodeToolConfig(srv.URL + "/node.tar.gz")
	events, err := download.Download(cfg, t.TempDir(), download.Ref{Category: "tools", ID: "eslint"})
	require.NoError(t, err)

	var toolFailed *download.Event
	for ev := range events {
		if ev.Ref.Category == "tools" && ev.Ref.ID == "eslint" && ev.Phase == download.Failed {
			e := ev
			toolFailed = &e
		}
	}
	require.NotNil(t, toolFailed, "the tool ref itself must report Failed when its runtime fails to fetch")
	require.Error(t, toolFailed.Err)
	assert.NotContains(t, toolFailed.Err.Error(), "npm not found",
		"must report the runtime's real failure, not the downstream symptom of proceeding anyway")
	assert.ErrorContains(t, toolFailed.Err, "unexpected status", "must surface the runtime download's actual HTTP failure")
}

// TestDownload_Runtime_FailedInstallNotPoisoned pins down Fix 3: a failed/interrupted install
// must not leave installDir behind looking "done" -- dirNonEmpty(installDir) is the only signal
// fetchRuntimeRef has for "already cached", and the pre-fix InstallDownload created that directory
// as its very first action, before ever touching the blob. So a first attempt that fails partway
// (here: a corrupt archive) must not make a second attempt for the same ref report Cached; it must
// retry for real.
func TestDownload_Runtime_FailedInstallNotPoisoned(t *testing.T) {
	good := tarGzBytes(t, "tool-1.0.0", "shellcheck", "#!/bin/sh\n")
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			_, _ = w.Write([]byte("not a valid gzip stream"))
			return
		}
		_, _ = w.Write(good)
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
		Runtimes: config.CategoryConfig[config.Runtime]{
			Definitions: map[string]config.Runtime{
				"shellcheck": {Type: "shellcheck", Download: "shellcheck", KnownGoodVersion: "1.0.0", Shims: []string{"shellcheck"}},
			},
		},
	}
	cacheDir := t.TempDir()

	events, err := download.Download(cfg, cacheDir, download.Ref{Category: "runtimes", ID: "shellcheck"})
	require.NoError(t, err)
	var sawFailed bool
	for ev := range events {
		if ev.Phase == download.Failed {
			sawFailed = true
		}
	}
	require.True(t, sawFailed, "the corrupt archive must fail the first attempt")

	events, err = download.Download(cfg, cacheDir, download.Ref{Category: "runtimes", ID: "shellcheck"})
	require.NoError(t, err)
	var phases []download.Phase
	for ev := range events {
		require.NotEqual(t, download.Failed, ev.Phase, "event: %+v", ev)
		phases = append(phases, ev.Phase)
	}
	assert.NotContains(t, phases, download.Cached, "a failed install must not poison the cache as Cached")
	assert.Contains(t, phases, download.Done, "the retry must actually (re)install")
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
