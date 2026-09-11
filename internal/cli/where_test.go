package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xunleii/rtunk/pkg/trunk/download"
)

func TestWhereCmd_NotDownloaded(t *testing.T) {
	_, stderr, err := run2(t, "--config", trunkYAML, "--cache-dir", t.TempDir(), "where", "runtimes", "node")
	assert.Error(t, err, "stderr: %s", stderr)
}

func TestWhereCmd_Found(t *testing.T) {
	cacheDir := t.TempDir()
	root, err := download.Root(cacheDir)
	require.NoError(t, err)
	shimPath := download.ShimPath(root, "actions", "commitlint", "1.2.3", "commitlint")
	require.NoError(t, os.MkdirAll(filepath.Dir(shimPath), 0o755))
	require.NoError(t, os.WriteFile(shimPath, []byte("#!/bin/sh\n"), 0o755))

	// commitlint's actual version comes from the fixture's KnownGoodVersion / enabled pin; this
	// test seeds the exact version resolveConfig+ResolveVersion will land on -- see the "Note for
	// implementer" below.
	stdout, stderr, err := run2(t, "--config", trunkYAML, "--cache-dir", cacheDir, "where", "actions", "commitlint")
	require.NoError(t, err, "stderr: %s", stderr)
	assert.Contains(t, stdout, shimPath)
}
