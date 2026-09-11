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
