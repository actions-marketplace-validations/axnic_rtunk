package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xunleii/rtunk/pkg/trunk/download"
)

func TestExecCmd_UsesExistingShim(t *testing.T) {
	cacheDir := t.TempDir()
	root, err := download.Root(cacheDir)
	require.NoError(t, err)
	shimPath := download.ShimPath(root, "actions", "commitlint", "1.2.3", "commitlint")
	require.NoError(t, os.MkdirAll(filepath.Dir(shimPath), 0o755))
	require.NoError(t, os.WriteFile(shimPath, []byte("#!/bin/sh\necho ran $1\n"), 0o755))

	stdout, stderr, err := run2(t, "--config", trunkYAML, "--cache-dir", cacheDir, "exec", "actions", "commitlint", "--", "hello")
	require.NoError(t, err, "stderr: %s", stderr)
	assert.Contains(t, stdout, "ran hello")
}

// TestExecCmd_StreamsStdoutAndStderrSeparately guards against Kong's by-type DI collapsing both
// parameters onto the same bound io.Writer (both were declared as plain io.Writer, so Kong -- which
// resolves method parameters by static type only -- fed the single stdout binding to both). If that
// regresses, the shim's stderr line would leak into stdout instead of staying isolated.
func TestExecCmd_StreamsStdoutAndStderrSeparately(t *testing.T) {
	cacheDir := t.TempDir()
	root, err := download.Root(cacheDir)
	require.NoError(t, err)
	shimPath := download.ShimPath(root, "actions", "commitlint", "1.2.3", "commitlint")
	require.NoError(t, os.MkdirAll(filepath.Dir(shimPath), 0o755))
	require.NoError(t, os.WriteFile(shimPath, []byte("#!/bin/sh\necho on-stdout\necho on-stderr >&2\n"), 0o755))

	stdout, stderr, err := run2(t, "--config", trunkYAML, "--cache-dir", cacheDir, "exec", "actions", "commitlint")
	require.NoError(t, err, "stderr: %s", stderr)
	assert.Contains(t, stdout, "on-stdout")
	assert.NotContains(t, stdout, "on-stderr")
	assert.Contains(t, stderr, "on-stderr")
	assert.NotContains(t, stderr, "on-stdout")
}

// TestExecCmd_XAlias exercises the hidden `x` alias registered alongside `exec` in cli.go, so both
// names routing to the same execCmd type is actually verified rather than assumed.
func TestExecCmd_XAlias(t *testing.T) {
	cacheDir := t.TempDir()
	root, err := download.Root(cacheDir)
	require.NoError(t, err)
	shimPath := download.ShimPath(root, "actions", "commitlint", "1.2.3", "commitlint")
	require.NoError(t, os.MkdirAll(filepath.Dir(shimPath), 0o755))
	require.NoError(t, os.WriteFile(shimPath, []byte("#!/bin/sh\necho ran $1\n"), 0o755))

	stdout, stderr, err := run2(t, "--config", trunkYAML, "--cache-dir", cacheDir, "x", "actions", "commitlint", "--", "hello")
	require.NoError(t, err, "stderr: %s", stderr)
	assert.Contains(t, stdout, "ran hello")
}
