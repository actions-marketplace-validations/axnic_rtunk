package config

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
