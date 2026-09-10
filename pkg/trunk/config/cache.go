package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

// sourceDefs holds every definition contributed by a single plugin source, unmerged — the shape
// both the on-disk cache stores and mergeSourceInto folds into a Config.
type sourceDefs struct {
	// Environments and CommentFormats are global config, not per-id definitions: no map, just
	// concatenated across every plugin.yaml that contributes them (ARCHITECTURE.md "Built-in /
	// global config").
	Environments   []NamedEnvironment
	CommentFormats []CommentFormat

	Downloads map[string]Download
	Tools     map[string]Tool
	Lint      map[string]Linter
	Files     map[string]FileType
	Actions   map[string]Action
	Runtimes  map[string]Runtime
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
