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

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

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

// Download is a reusable, OS/CPU-templated download recipe (ARCHITECTURE.md `downloads:`).
type Download struct {
	Name      string          `yaml:"name"`
	Version   string          `yaml:"version"`
	Downloads []DownloadEntry `yaml:"downloads"`
}

// DownloadEntry is one os/cpu-specific variant of a Download recipe.
type DownloadEntry struct {
	OS              OSSpec `yaml:"os"`
	CPU             OSSpec `yaml:"cpu"`
	URL             string `yaml:"url"`
	StripComponents int    `yaml:"strip_components,omitempty"`
	Executable      bool   `yaml:"executable,omitempty"`
	Version         string `yaml:"version,omitempty"`
}

// OSSpec is an os/cpu selector: either a bare name ("macos"), or a map from trunk's own
// vocabulary (linux, macos, windows, x86_64, arm_64) to upstream's naming. A bare name decodes
// to a single key mapping to itself.
type OSSpec map[string]string

func (s *OSSpec) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		*s = OSSpec{node.Value: node.Value}
		return nil
	}
	var m map[string]string
	if err := node.Decode(&m); err != nil {
		return err
	}
	*s = m
	return nil
}

// Tool is a downloadable/runnable tool (ARCHITECTURE.md `tools:`), fetched either via a runtime's
// package manager (Runtime+Package) or a Download recipe. The two are mutually exclusive.
type Tool struct {
	Name             string   `yaml:"name"`
	Runtime          string   `yaml:"runtime,omitempty"`
	Package          string   `yaml:"package,omitempty"`
	Download         string   `yaml:"download,omitempty"`
	Shims            []string `yaml:"shims,omitempty"`
	KnownGoodVersion string   `yaml:"known_good_version,omitempty"`
}

// Linter is a linter definition (ARCHITECTURE.md `lint:`), tying files, tools, and commands
// together.
type Linter struct {
	Name               string          `yaml:"name"`
	Files              []string        `yaml:"files,omitempty"`
	Tools              []string        `yaml:"tools,omitempty"`
	MainTool           string          `yaml:"main_tool,omitempty"`
	Description        string          `yaml:"description,omitempty"`
	Commands           []Command       `yaml:"commands,omitempty"`
	DirectConfigs      []string        `yaml:"direct_configs,omitempty"`
	AffectsCache       []string        `yaml:"affects_cache,omitempty"`
	IssueURLFormat     string          `yaml:"issue_url_format,omitempty"`
	SuggestIf          string          `yaml:"suggest_if,omitempty"`
	KnownGoodVersion   string          `yaml:"known_good_version,omitempty"`
	KnownBadVersions   []string        `yaml:"known_bad_versions,omitempty"`
	SupportedPlatforms []string        `yaml:"supported_platforms,omitempty"`
	RunTimeout         string          `yaml:"run_timeout,omitempty"`
	CacheResults       *bool           `yaml:"cache_results,omitempty"`
	VersionCommand     *VersionCommand `yaml:"version_command,omitempty"`
}

// Command is one invocation of a Linter (ARCHITECTURE.md `commands[]`): a checker command by
// default, or a formatter when InPlace+Formatter are set.
type Command struct {
	Name           string  `yaml:"name"`
	Run            string  `yaml:"run"`
	Output         string  `yaml:"output,omitempty"`
	SuccessCodes   []int   `yaml:"success_codes,omitempty"`
	ErrorCodes     []int   `yaml:"error_codes,omitempty"`
	Batch          bool    `yaml:"batch,omitempty"`
	ReadOutputFrom string  `yaml:"read_output_from,omitempty"`
	SandboxType    string  `yaml:"sandbox_type,omitempty"`
	RunFrom        string  `yaml:"run_from,omitempty"`
	Version        string  `yaml:"version,omitempty"`
	InPlace        bool    `yaml:"in_place,omitempty"`
	Formatter      bool    `yaml:"formatter,omitempty"`
	Parser         *Parser `yaml:"parser,omitempty"`
}

// Parser converts a tool's native output into trunk's normalized shape, for tools with no native
// structured output mode.
type Parser struct {
	Runtime string `yaml:"runtime"`
	Run     string `yaml:"run"`
}

// VersionCommand detects an installed tool/runtime's version.
type VersionCommand struct {
	Run        string `yaml:"run"`
	ParseRegex string `yaml:"parse_regex"`
}

// Action is a git-hook or file-change-triggered automation (ARCHITECTURE.md `actions:`).
type Action struct {
	ID            string        `yaml:"id"`
	DisplayName   string        `yaml:"display_name,omitempty"`
	Description   string        `yaml:"description,omitempty"`
	Runtime       string        `yaml:"runtime,omitempty"`
	PackagesFile  string        `yaml:"packages_file,omitempty"`
	Run           string        `yaml:"run,omitempty"`
	Triggers      []Trigger     `yaml:"triggers,omitempty"`
	Interactive   Interactivity `yaml:"interactive,omitempty"`
	NotifyOnError bool          `yaml:"notify_on_error,omitempty"`
}

// Interactivity is Action.Interactive: bare `true`, or the literal string "optional".
type Interactivity string

func (i *Interactivity) UnmarshalYAML(node *yaml.Node) error {
	*i = Interactivity(node.Value)
	return nil
}

// Trigger is one alternative way an Action can fire.
type Trigger struct {
	GitHooks []string  `yaml:"git_hooks,omitempty"`
	Files    []string  `yaml:"files,omitempty"`
	Schedule *Schedule `yaml:"schedule,omitempty"`
}

// Schedule is a periodic background trigger.
type Schedule struct {
	Interval string `yaml:"interval"`
	Delay    string `yaml:"delay,omitempty"`
}

// Runtime is a language runtime definition (ARCHITECTURE.md `runtimes:`).
type Runtime struct {
	Type               string             `yaml:"type"`
	Download           string             `yaml:"download,omitempty"`
	SystemVersion      string             `yaml:"system_version,omitempty"`
	KnownGoodVersion   string             `yaml:"known_good_version,omitempty"`
	Shims              []string           `yaml:"shims,omitempty"`
	VersionCommands    []VersionCommand   `yaml:"version_commands,omitempty"`
	RuntimeEnvironment []EnvironmentEntry `yaml:"runtime_environment,omitempty"`
	LinterEnvironment  []EnvironmentEntry `yaml:"linter_environment,omitempty"`
}

// EnvironmentEntry is one environment variable/PATH entry contributed by a Runtime.
type EnvironmentEntry struct {
	Name     string   `yaml:"name"`
	List     []string `yaml:"list,omitempty"`
	Value    string   `yaml:"value,omitempty"`
	Optional bool     `yaml:"optional,omitempty"`
}

// ReadError reports that trunk.yaml or a plugin.yaml could not be read from disk — e.g. it
// doesn't exist, or permissions deny access. Unwrap returns the underlying error (typically an
// *fs.PathError), so callers can e.g. errors.Is(err, os.ErrNotExist).
type ReadError struct {
	Path string
	Err  error
}

func (e *ReadError) Error() string { return fmt.Sprintf("config: read %s: %v", e.Path, e.Err) }
func (e *ReadError) Unwrap() error { return e.Err }

// ParseError reports that trunk.yaml or a plugin.yaml was read successfully but is not valid
// YAML (or doesn't match the expected shape). Unwrap returns the underlying yaml decode error.
type ParseError struct {
	Path string
	Err  error
}

func (e *ParseError) Error() string { return fmt.Sprintf("config: parse %s: %v", e.Path, e.Err) }
func (e *ParseError) Unwrap() error { return e.Err }

// SourceNotFoundError reports that a plugins.sources entry's `local:` path does not exist on
// disk. A local source is expected to already be present — unlike a git source (uri/ref), which
// Resolve deliberately never fetches (that would be network access), a missing local source is a
// configuration error, not something Resolve silently tolerates into an empty result.
type SourceNotFoundError struct {
	SourceID string
	Path     string
}

func (e *SourceNotFoundError) Error() string {
	return fmt.Sprintf("config: plugin source %q: local path %s does not exist", e.SourceID, e.Path)
}

// DuplicateError reports that a resource was defined more than once while merging config: two
// plugin.yaml files (or two entries in the same one) declaring the same download/tool/lint/
// action/runtime name, or trunk.yaml's plugins.sources repeating an id. It is never fatal on its
// own — the later definition always fully replaces the earlier one, never a field-by-field merge
// — but it is always reported, since a repeated name is almost always a mistake.
type DuplicateError struct {
	Category string // "plugin source", "download", "tool", "lint", "action", or "runtime"
	Key      string // the id/name that was defined more than once
}

func (e *DuplicateError) Error() string {
	return fmt.Sprintf("config: duplicate %s %q", e.Category, e.Key)
}

// ReferenceError reports that a definition names another resource, by id, that does not exist
// anywhere in the resolved config — e.g. a tool's `download:`/`runtime:` naming a download/
// runtime no plugin source defines, a linter's `tools:` naming an undefined tool, or trunk.yaml
// enabling a lint/action/runtime with no matching definition.
type ReferenceError struct {
	Category  string // category of the definition holding the reference: "tool", "runtime", "lint", "lint.enabled", ...
	Key       string // id/name of the definition holding the reference
	Field     string // the field holding the reference: "download", "runtime", "tools", "enabled"
	Reference string // the missing id/name
}

func (e *ReferenceError) Error() string {
	return fmt.Sprintf("config: %s %q: %s %q not found", e.Category, e.Key, e.Field, e.Reference)
}

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
	complete := true // whether every plugin source was actually resolved
	for _, src := range tf.Plugins.Sources {
		if src.Local == "" {
			complete = false // git source: needs a clone, out of scope here (ROADMAP.md v0.2)
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
	if complete {
		errs = append(errs, checkEnabled("runtime", cfg.Runtimes.Enabled, cfg.Runtimes.Definitions)...)
		errs = append(errs, checkEnabled("lint", cfg.Lint.Enabled, cfg.Lint.Definitions)...)
		errs = append(errs, checkEnabled("action", cfg.Actions.Enabled, cfg.Actions.Definitions)...)
	}
	errs = append(errs, validateReferences(&cfg)...)

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
