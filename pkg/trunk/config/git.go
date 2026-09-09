package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// fetchGitSource ensures src's pinned ref is checked out on disk under cacheDir, fetching it via
// the system git binary if it isn't already cached, and returns the checkout directory to merge
// definitions from. Caching is keyed by uri+ref (never mutates once fetched, since ref is always
// a tag or SHA per ARCHITECTURE.md): a cache hit skips the network entirely.
//
// ponytail: no TTL/prune — the cache only grows. Add eviction when `cache clean`/`cache prune`
// (ROADMAP v0.2) exist to drive it.
func fetchGitSource(cacheDir string, src PluginSource) (string, error) {
	if cacheDir == "" {
		dir, err := os.UserCacheDir()
		if err != nil {
			return "", &FetchError{SourceID: src.ID, URI: src.URI, Ref: src.Ref, Err: err}
		}
		cacheDir = filepath.Join(dir, "rtunk", "plugins")
	}

	sum := sha256.Sum256([]byte(src.URI + "@" + src.Ref))
	dest := filepath.Join(cacheDir, hex.EncodeToString(sum[:]))

	if info, err := os.Stat(dest); err == nil && info.IsDir() {
		return dest, nil // already fetched at this exact uri+ref
	}

	if err := os.MkdirAll(dest, 0o755); err != nil {
		return "", &FetchError{SourceID: src.ID, URI: src.URI, Ref: src.Ref, Err: err}
	}
	for _, args := range [][]string{
		{"init"},
		{"remote", "add", "origin", src.URI},
		{"fetch", "--depth", "1", "origin", src.Ref},
		{"checkout", "FETCH_HEAD"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dest
		if out, err := cmd.CombinedOutput(); err != nil {
			_ = os.RemoveAll(dest) // don't leave a half-fetched dir behind as a false cache hit
			return "", &FetchError{SourceID: src.ID, URI: src.URI, Ref: src.Ref, Err: fmt.Errorf("%v: %s", err, out)}
		}
	}

	return dest, nil
}
