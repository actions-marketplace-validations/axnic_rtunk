package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// pluginCategories are the plugin repository subdirs that hold plugin.yaml files
// (ARCHITECTURE.md "Plugin repository layout"). repo-tools/ is deliberately excluded: it is
// internal tooling for the plugins repo itself, not consumed by trunk/rtunk.
var pluginCategories = []string{"linters", "actions", "tools", "runtimes"}

// trunkFile is the raw shape of a trunk.yaml file, as YAML naturally decodes it (lists, not
// maps) — the lecture phase's output, before Resolve merges it into a map-keyed Config.
type trunkFile struct {
	Version string `yaml:"version"`
	CLI     struct {
		Version string `yaml:"version"`
	} `yaml:"cli"`
	Plugins struct {
		Sources []PluginSource `yaml:"sources"`
	} `yaml:"plugins"`
	Runtimes struct {
		Enabled []string `yaml:"enabled"`
	} `yaml:"runtimes"`
	Lint struct {
		Enabled []string `yaml:"enabled"`
	} `yaml:"lint"`
	Actions struct {
		Enabled []string `yaml:"enabled"`
	} `yaml:"actions"`
}

// pluginFile is the raw shape of one category plugin.yaml: any mix of the section kinds below
// (ARCHITECTURE.md "a single file commonly mixes sections").
type pluginFile struct {
	Downloads []Download `yaml:"downloads"`
	Tools     struct {
		Definitions []Tool `yaml:"definitions"`
	} `yaml:"tools"`
	Lint struct {
		Definitions []Linter `yaml:"definitions"`
	} `yaml:"lint"`
	Actions struct {
		Definitions []Action `yaml:"definitions"`
	} `yaml:"actions"`
	Runtimes struct {
		Definitions []Runtime `yaml:"definitions"`
	} `yaml:"runtimes"`
}

// Resolve reads the trunk.yaml file at path, merges in every definition contributed by its local
// plugin sources, and validates the result. See the package doc for the three-phase pipeline.
// The returned Config is always populated with everything Resolve managed to read, even when it
// also returns an error — callers that only care about specific resources may still use it.
func Resolve(file string) (Config, error) {
	cfg := Config{
		Tools:     map[string]Tool{},
		Downloads: map[string]Download{},
	}
	cfg.Plugins.Sources = map[string]PluginSource{}
	cfg.Runtimes.Definitions = map[string]Runtime{}
	cfg.Lint.Definitions = map[string]Linter{}
	cfg.Actions.Definitions = map[string]Action{}

	// 1. Lecture
	tf, err := readTrunkFile(file)
	if err != nil {
		return cfg, err
	}
	cfg.Version = tf.Version
	cfg.CLI.Version = tf.CLI.Version
	cfg.Runtimes.Enabled = tf.Runtimes.Enabled
	cfg.Lint.Enabled = tf.Lint.Enabled
	cfg.Actions.Enabled = tf.Actions.Enabled

	var errs []error
	mergeKeyed(cfg.Plugins.Sources, tf.Plugins.Sources, func(s PluginSource) string { return s.ID }, "plugin source", &errs)

	// 2. Merge
	for _, src := range tf.Plugins.Sources {
		if src.Local == "" {
			cfg.incomplete = true // git source: needs a clone, out of scope here (ROADMAP.md v0.2)
			continue
		}
		dir := filepath.Join(filepath.Dir(file), src.Local)
		if info, statErr := os.Stat(dir); statErr != nil || !info.IsDir() {
			return cfg, &SourceNotFoundError{SourceID: src.ID, Path: dir}
		}
		if err := mergePluginRepo(&cfg, dir, &errs); err != nil {
			return cfg, err
		}
	}

	// 3. Validation
	errs = append(errs, cfg.Validate())

	return cfg, errors.Join(errs...)
}

func readTrunkFile(path string) (trunkFile, error) {
	var tf trunkFile
	data, err := os.ReadFile(path)
	if err != nil {
		return tf, &ReadError{Path: path, Err: err}
	}
	if err := yaml.Unmarshal(data, &tf); err != nil {
		return tf, &ParseError{Path: path, Err: err}
	}
	return tf, nil
}

// mergePluginRepo walks every plugin.yaml under dir's category subdirs and merges its
// definitions into cfg, recording any duplicate keys into *errs.
func mergePluginRepo(cfg *Config, dir string, errs *[]error) error {
	for _, category := range pluginCategories {
		matches, err := filepath.Glob(filepath.Join(dir, category, "*", "plugin.yaml"))
		if err != nil {
			return err // malformed glob pattern; unreachable with our fixed patterns
		}
		for _, path := range matches {
			data, err := os.ReadFile(path)
			if err != nil {
				return &ReadError{Path: path, Err: err}
			}
			var pf pluginFile
			if err := yaml.Unmarshal(data, &pf); err != nil {
				return &ParseError{Path: path, Err: err}
			}

			mergeKeyed(cfg.Downloads, pf.Downloads, func(d Download) string { return d.Name }, "download", errs)
			mergeKeyed(cfg.Tools, pf.Tools.Definitions, func(t Tool) string { return t.Name }, "tool", errs)
			mergeKeyed(cfg.Lint.Definitions, pf.Lint.Definitions, func(l Linter) string { return l.Name }, "lint", errs)
			mergeKeyed(cfg.Actions.Definitions, pf.Actions.Definitions, func(a Action) string { return a.ID }, "action", errs)
			mergeKeyed(cfg.Runtimes.Definitions, pf.Runtimes.Definitions, func(r Runtime) string { return r.Type }, "runtime", errs)
		}
	}
	return nil
}

// mergeKeyed inserts each item into dst, keyed by key(item). A key already present in dst (from
// an earlier file or source) is fully overwritten — never merged field-by-field — and the
// collision is appended to *errs as a *DuplicateError.
func mergeKeyed[T any](dst map[string]T, items []T, key func(T) string, category string, errs *[]error) {
	for _, item := range items {
		k := key(item)
		if _, exists := dst[k]; exists {
			*errs = append(*errs, &DuplicateError{Category: category, Key: k})
		}
		dst[k] = item
	}
}

// checkEnabled reports a *ReferenceError for every entry of enabled (a trunk.yaml `enabled:`
// list, each optionally pinned as `id@version`) whose id has no matching key in defs.
func checkEnabled[T any](category string, enabled []string, defs map[string]T) []error {
	var errs []error
	for _, e := range enabled {
		id, _, _ := strings.Cut(e, "@")
		if _, ok := defs[id]; !ok {
			errs = append(errs, &ReferenceError{Category: category, Key: e, Field: "enabled", Reference: id})
		}
	}
	return errs
}

// validateReferences reports a *ReferenceError for every by-id reference (a tool's
// download/runtime, a runtime's download, a linter's tools) that doesn't resolve to a key
// actually present in cfg.
func validateReferences(cfg *Config) []error {
	var errs []error

	for name, t := range cfg.Tools {
		if t.Runtime != "" {
			if _, ok := cfg.Runtimes.Definitions[t.Runtime]; !ok {
				errs = append(errs, &ReferenceError{Category: "tool", Key: name, Field: "runtime", Reference: t.Runtime})
			}
		}
		if t.Download != "" {
			if _, ok := cfg.Downloads[t.Download]; !ok {
				errs = append(errs, &ReferenceError{Category: "tool", Key: name, Field: "download", Reference: t.Download})
			}
		}
	}

	for typ, r := range cfg.Runtimes.Definitions {
		if r.Download != "" {
			if _, ok := cfg.Downloads[r.Download]; !ok {
				errs = append(errs, &ReferenceError{Category: "runtime", Key: typ, Field: "download", Reference: r.Download})
			}
		}
	}

	for name, l := range cfg.Lint.Definitions {
		for _, toolName := range l.Tools {
			if _, ok := cfg.Tools[toolName]; !ok {
				errs = append(errs, &ReferenceError{Category: "lint", Key: name, Field: "tools", Reference: toolName})
			}
		}
	}

	return errs
}
