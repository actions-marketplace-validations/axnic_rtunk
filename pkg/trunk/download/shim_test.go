package download_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xunleii/rtunk/pkg/trunk/download"
)

func TestFindShimTarget(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "bin"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bin", "node"), []byte(""), 0o755))

	got, err := download.FindShimTarget(dir, "node")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "bin", "node"), got)

	_, err = download.FindShimTarget(dir, "missing")
	assert.Error(t, err)
}

func TestWriteShim_Exec(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "real-tool")
	require.NoError(t, os.WriteFile(target, []byte("#!/bin/sh\necho ran-$1\n"), 0o755))

	shimPath := filepath.Join(dir, "shim")
	require.NoError(t, download.WriteShim(shimPath, target))

	out, err := exec.Command(shimPath, "ok").CombinedOutput()
	require.NoError(t, err)
	assert.Equal(t, "ran-ok\n", string(out))
}

func TestWriteEnvShim_Exec(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "real-tool")
	require.NoError(t, os.WriteFile(target, []byte("#!/bin/sh\necho $MY_VAR\n"), 0o755))

	shimPath := filepath.Join(dir, "shim")
	require.NoError(t, download.WriteEnvShim(shimPath, target, []string{"MY_VAR=hello"}))

	out, err := exec.Command(shimPath).CombinedOutput()
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(out))
}
