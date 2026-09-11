package download_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xunleii/rtunk/pkg/trunk/config"
	"github.com/xunleii/rtunk/pkg/trunk/download"
)

// makeTarGz writes a single-file tar.gz at path, wrapped in a top-level dir (to exercise
// strip_components), containing name with the given content and mode.
func makeTarGz(t *testing.T, path, topDir, name, content string, mode int64) {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name: topDir + "/" + name, Mode: mode, Size: int64(len(content)),
	}))
	_, err := tw.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	require.NoError(t, os.WriteFile(path, buf.Bytes(), 0o644))
}

func TestInstallDownload_TarGz_StripComponents(t *testing.T) {
	dir := t.TempDir()
	blob := filepath.Join(dir, "blob")
	makeTarGz(t, blob, "shellcheck-v0.11.0", "shellcheck", "#!/bin/sh\necho hi\n", 0o755)

	dest := filepath.Join(dir, "install")
	entry := config.DownloadEntry{StripComponents: 1}
	err := download.InstallDownload(blob, "https://example.com/shellcheck-v0.11.0.linux.x86_64.tar.gz", dest, entry)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dest, "shellcheck"))
	require.NoError(t, err)
	assert.Equal(t, "#!/bin/sh\necho hi\n", string(data))

	info, err := os.Stat(filepath.Join(dest, "shellcheck"))
	require.NoError(t, err)
	assert.NotZero(t, info.Mode()&0o111, "extracted file should keep its executable bit")
}

func TestInstallDownload_Executable(t *testing.T) {
	dir := t.TempDir()
	blob := filepath.Join(dir, "blob")
	require.NoError(t, os.WriteFile(blob, []byte("binary content"), 0o644))

	dest := filepath.Join(dir, "install")
	entry := config.DownloadEntry{Executable: true}
	err := download.InstallDownload(blob, "https://example.com/tool-linux-amd64", dest, entry)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dest, "tool-linux-amd64"))
	require.NoError(t, err)
	assert.Equal(t, "binary content", string(data))

	info, err := os.Stat(filepath.Join(dest, "tool-linux-amd64"))
	require.NoError(t, err)
	assert.NotZero(t, info.Mode()&0o111, "a bare Executable download must be made executable")
}

// TestInstallDownload_TarGz_PathTraversal pins down zip-slip protection (CWE-22): an entry whose
// name, after strip_components strips its leading "top/", is "../escaped.txt" must not be allowed
// to write outside dest -- filepath.Join(destDir, name) would otherwise happily escape it, one
// level up from install into dir (still under t.TempDir(), so a pre-fix failure self-cleans).
func TestInstallDownload_TarGz_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	blob := filepath.Join(dir, "blob")
	makeTarGz(t, blob, "top", "../escaped.txt", "pwned", 0o644)

	dest := filepath.Join(dir, "install")
	entry := config.DownloadEntry{StripComponents: 1}
	err := download.InstallDownload(blob, "https://example.com/archive.tar.gz", dest, entry)
	assert.Error(t, err)

	_, statErr := os.Stat(filepath.Join(dir, "escaped.txt"))
	assert.True(t, os.IsNotExist(statErr), "entry must not have been written outside destDir")
}

// makeZip writes a single-file zip archive at path, wrapped in topDir (to exercise
// strip_components), containing name with the given content.
func makeZip(t *testing.T, path, topDir, name, content string) {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create(topDir + "/" + name)
	require.NoError(t, err)
	_, err = w.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	require.NoError(t, os.WriteFile(path, buf.Bytes(), 0o644))
}

// TestInstallDownload_Zip_PathTraversal is TestInstallDownload_TarGz_PathTraversal's zip
// equivalent -- extractZip strips and joins paths the same way extractTar does.
func TestInstallDownload_Zip_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	blob := filepath.Join(dir, "blob")
	makeZip(t, blob, "top", "../escaped.txt", "pwned")

	dest := filepath.Join(dir, "install")
	entry := config.DownloadEntry{StripComponents: 1}
	err := download.InstallDownload(blob, "https://example.com/archive.zip", dest, entry)
	assert.Error(t, err)

	_, statErr := os.Stat(filepath.Join(dir, "escaped.txt"))
	assert.True(t, os.IsNotExist(statErr), "entry must not have been written outside destDir")
}

// TestInstallDownload_Concurrent_SameDest is Fix 3's concurrency regression test, done at
// InstallDownload's level (the choke point both fetchRuntimeRef and fetchToolRef funnel through)
// rather than through the full Download() pipeline: Download()'s internal goroutine scheduling
// (a semaphore-bounded WaitGroup) gives no hook to force two goroutines through the "not yet
// cached" check at the same instant, but two callers landing in InstallDownload for the same
// destDir at the same time is exactly the race the reviewer reproduced -- both would have just
// passed that check. Run with -race: pre-fix, InstallDownload extracts straight into destDir, so
// two concurrent callers interleave writes into the same tree (a real, sometimes-flagged race, and
// always a risk of a mixed/partial result even when the writes don't literally collide). Post-fix,
// each extracts into its own temp dir and only one wins the atomic rename into destDir -- no
// shared mutable state during extraction, so no race, and destDir ends up wholly one extraction's
// output, never an interleaving of both.
func TestInstallDownload_Concurrent_SameDest(t *testing.T) {
	dir := t.TempDir()
	blob := filepath.Join(dir, "blob")
	makeTarGz(t, blob, "shellcheck-1.0.0", "shellcheck", "#!/bin/sh\necho hi\n", 0o755)

	dest := filepath.Join(dir, "install")
	entry := config.DownloadEntry{StripComponents: 1}

	start := make(chan struct{})
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := range 2 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			errs[i] = download.InstallDownload(blob, "https://example.com/shellcheck.tar.gz", dest, entry)
		}(i)
	}
	close(start)
	wg.Wait()

	for _, err := range errs {
		assert.NoError(t, err, "the loser of the race must treat an already-completed dest as success, not fail")
	}

	data, err := os.ReadFile(filepath.Join(dest, "shellcheck"))
	require.NoError(t, err)
	assert.Equal(t, "#!/bin/sh\necho hi\n", string(data), "destDir must hold one complete extraction, not a partial/mixed write")
}

func TestInstallDownload_UnrecognizedFormat(t *testing.T) {
	dir := t.TempDir()
	blob := filepath.Join(dir, "blob")
	require.NoError(t, os.WriteFile(blob, []byte("?"), 0o644))
	err := download.InstallDownload(blob, "https://example.com/tool.unknown", filepath.Join(dir, "install"), config.DownloadEntry{})
	assert.Error(t, err)
}
