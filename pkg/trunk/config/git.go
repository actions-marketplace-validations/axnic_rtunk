package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// sourceDefs holds every definition contributed by a single plugin source, unmerged — the shape
// both the on-disk cache stores (as JSON, not gob: gob elides pointer/slice fields that are the
// zero value, which would silently turn e.g. a linter's explicit `cache_results: false` into a
// nil "unset" after a cache round-trip) and mergeSourceInto folds into a Config.
type sourceDefs struct {
	Downloads map[string]Download
	Tools     map[string]Tool
	Lint      map[string]Linter
	Actions   map[string]Action
	Runtimes  map[string]Runtime
}

// fetchGitSource returns everything a git plugin source (uri/ref) contributes, preferring a
// cache of the already-parsed definitions at cacheDir/<sha256(uri+ref)>.json over touching the
// network at all. Caching is keyed by uri+ref (never mutates once fetched, since ref is always a
// tag or SHA per ARCHITECTURE.md), so a cache hit is safe to reuse indefinitely.
//
// On a cache miss — or a cache file that fails to decode, e.g. corrupted or from an incompatible
// rtunk version — it's dropped and regenerated: src's ref is cloned into a throwaway temp dir
// (removed once parsing finishes, never persisted itself), parsed, and the result written back to
// the cache for next time. dupErrs carries any *DuplicateError found while parsing (only possible
// on a cold fetch: a cache hit returns the already-deduplicated result, so there's nothing left to
// report).
//
// ponytail: no TTL/prune — the cache only grows. Add eviction when `cache clean`/`cache prune`
// (ROADMAP v0.2) exist to drive it.
func fetchGitSource(cacheDir string, src PluginSource) (defs sourceDefs, dupErrs []error, err error) {
	if cacheDir == "" {
		dir, err := os.UserCacheDir()
		if err != nil {
			return sourceDefs{}, nil, &FetchError{SourceID: src.ID, URI: src.URI, Ref: src.Ref, Err: err}
		}
		cacheDir = filepath.Join(dir, "rtunk", "plugins")
	}

	cacheFile := cacheFilePath(cacheDir, src)

	if defs, err := loadSourceCache(cacheFile); err == nil {
		return defs, nil, nil
	}
	_ = os.Remove(cacheFile) // missing is fine; corrupt/stale is dropped so it regenerates below

	tmpDir, err := os.MkdirTemp("", "rtunk-plugin-*")
	if err != nil {
		return sourceDefs{}, nil, &FetchError{SourceID: src.ID, URI: src.URI, Ref: src.Ref, Err: err}
	}
	defer os.RemoveAll(tmpDir) // only the parsed result is cached, never the checkout itself

	for _, args := range [][]string{
		{"init"},
		{"remote", "add", "origin", src.URI},
		{"fetch", "--depth", "1", "origin", src.Ref},
		{"checkout", "FETCH_HEAD"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			return sourceDefs{}, nil, &FetchError{SourceID: src.ID, URI: src.URI, Ref: src.Ref, Err: fmt.Errorf("%v: %s", err, out)}
		}
	}

	defs, dupErrs, err = parseSourceDir(tmpDir)
	if err != nil {
		return sourceDefs{}, nil, err
	}

	if err := saveSourceCache(cacheFile, defs); err != nil {
		return sourceDefs{}, nil, &FetchError{SourceID: src.ID, URI: src.URI, Ref: src.Ref, Err: err}
	}

	return defs, dupErrs, nil
}

// cacheFilePath returns where src's cache lives: keyed by uri+ref, since a pinned ref never
// changes content.
func cacheFilePath(cacheDir string, src PluginSource) string {
	sum := sha256.Sum256([]byte(src.URI + "@" + src.Ref))
	return filepath.Join(cacheDir, hex.EncodeToString(sum[:])+".json")
}

func loadSourceCache(path string) (sourceDefs, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return sourceDefs{}, err
	}

	var defs sourceDefs
	if err := json.Unmarshal(data, &defs); err != nil {
		return sourceDefs{}, err
	}
	return defs, nil
}

// saveSourceCache writes defs atomically (temp file + rename) so a crash mid-write never leaves a
// half-written file behind to be mistaken for a valid cache hit.
func saveSourceCache(path string, defs sourceDefs) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.Marshal(defs)
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // no-op once renamed below

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
