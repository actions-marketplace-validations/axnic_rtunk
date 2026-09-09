// Package config parses and resolves trunk.yaml — the two-layer model described in
// ARCHITECTURE.md. Resolve runs three phases in sequence:
//
//  1. Lecture: read trunk.yaml, and every plugin.yaml under its local plugin sources' category
//     dirs, each into its own raw shape (as YAML naturally decodes it: lists, not maps).
//  2. Merge: fold every list into the Config's maps, keyed by each resource's own id/name. A
//     repeated key is never field-by-field merged — the later definition fully replaces the
//     earlier one — but every such collision is recorded as a *DuplicateError.
//  3. Validation: check that everything referenced by id (a tool's download/runtime, a linter's
//     tools, trunk.yaml's own enabled lists) actually exists among what was merged, recording a
//     *ReferenceError for anything dangling.
//
// Every recorded error (duplicates and dangling references) is returned together via
// errors.Join, so a single Resolve call reports every problem instead of one per run.
//
// Git plugin sources (uri/ref) are recorded but never fetched: that would be a `git clone`, i.e.
// network access, out of scope here (ROADMAP.md v0.2). Resolve treats such a config as
// incomplete and skips step 3's enabled-list checks for it — there is no way to know whether an
// enabled id is defined by a source this package cannot read.
package config

import "errors"

// Config is a fully resolved trunk.yaml: the repo's own enabled lists, plus every definition
// merged in from its local plugin sources, keyed by each resource's own id/name so existence and
// duplication can be checked in O(1) rather than by scanning a list.
type Config struct {
	Version string
	CLI     struct {
		Version string
	}
	Plugins struct {
		Sources map[string]PluginSource
	}

	Runtimes CategoryConfig[Runtime]
	Lint     CategoryConfig[Linter]
	Actions  CategoryConfig[Action]

	// Tools and Downloads have no trunk.yaml `enabled:` list of their own; they are referenced
	// by name from lint/runtime definitions instead (ARCHITECTURE.md `tools:`/`downloads:`).
	Tools     map[string]Tool
	Downloads map[string]Download

	// incomplete is set by Resolve when a plugin source couldn't be read (e.g. a git source,
	// which Resolve never fetches). Validate then skips the enabled-list checks: there is no way
	// to know whether an enabled id is defined by a source this package cannot read.
	incomplete bool
}

// Validate checks that everything referenced by id (a tool's download/runtime, a runtime's
// download, a linter's tools, and — unless cfg is incomplete — trunk.yaml's own enabled lists)
// actually exists among what was merged into cfg, returning every problem joined together via
// errors.Join.
func (cfg *Config) Validate() error {
	var errs []error
	if !cfg.incomplete {
		errs = append(errs, checkEnabled("runtime", cfg.Runtimes.Enabled, cfg.Runtimes.Definitions)...)
		errs = append(errs, checkEnabled("lint", cfg.Lint.Enabled, cfg.Lint.Definitions)...)
		errs = append(errs, checkEnabled("action", cfg.Actions.Enabled, cfg.Actions.Definitions)...)
	}
	errs = append(errs, validateReferences(cfg)...)
	return errors.Join(errs...)
}

// PluginSource is one entry of plugins.sources: either a git source ({id, uri, ref}) or a local
// filesystem source ({id, local}).
type PluginSource struct {
	ID    string `yaml:"id"`
	URI   string `yaml:"uri,omitempty"`
	Ref   string `yaml:"ref,omitempty"`
	Local string `yaml:"local,omitempty"`
}

// CategoryConfig is the {enabled: [...]} shape shared by runtimes/lint/actions in trunk.yaml,
// plus the merged Definitions resolved from plugin sources (never present in trunk.yaml itself),
// keyed by each definition's own id/name.
type CategoryConfig[T any] struct {
	Enabled     []string
	Definitions map[string]T
}
