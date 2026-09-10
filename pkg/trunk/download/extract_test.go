package download_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
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

func TestInstallDownload_UnrecognizedFormat(t *testing.T) {
	dir := t.TempDir()
	blob := filepath.Join(dir, "blob")
	require.NoError(t, os.WriteFile(blob, []byte("?"), 0o644))
	err := download.InstallDownload(blob, "https://example.com/tool.unknown", filepath.Join(dir, "install"), config.DownloadEntry{})
	assert.Error(t, err)
}
