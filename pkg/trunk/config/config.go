// Package config parses and resolves trunk.yaml — the two-layer model described in
// ARCHITECTURE.md. Resolve reads the repo's own trunk.yaml, then merges in every definition
// (linters, tools, actions, runtimes, downloads) contributed by its local plugin sources.
// Git plugin sources (uri/ref) are recorded but not resolved: fetching them is a `git clone`,
// which is network access and belongs to a later milestone (ROADMAP.md v0.2).
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is a fully resolved trunk.yaml: the repo's own enabled lists, plus every definition
// merged in from its local plugin sources.
type Config struct {
	Version string `yaml:"version"`
	CLI     struct {
		Version string `yaml:"version"`
	} `yaml:"cli"`
	Plugins struct {
		Sources []PluginSource `yaml:"sources"`
	} `yaml:"plugins"`

	Runtimes CategoryConfig[Runtime] `yaml:"runtimes"`
	Lint     CategoryConfig[Linter]  `yaml:"lint"`
	Actions  CategoryConfig[Action]  `yaml:"actions"`

	// Tools and Downloads have no trunk.yaml `enabled:` list of their own; they are referenced
	// by name from lint/runtime definitions instead (ARCHITECTURE.md `tools:`/`downloads:`).
	Tools     []Tool
	Downloads []Download
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
// plus the merged Definitions resolved from plugin sources (never present in trunk.yaml itself).
type CategoryConfig[T any] struct {
	Enabled     []string `yaml:"enabled"`
	Definitions []T      `yaml:"-"`
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

// pluginCategories are the plugin repository subdirs that hold plugin.yaml files
// (ARCHITECTURE.md "Plugin repository layout"). repo-tools/ is deliberately excluded: it is
// internal tooling for the plugins repo itself, not consumed by trunk/rtunk.
var pluginCategories = []string{"linters", "actions", "tools", "runtimes"}

// pluginFile is the shape of one category plugin.yaml: any mix of the section kinds below
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

// Resolve reads the trunk.yaml file at path, then merges in every definition contributed by its
// local plugin sources.
func Resolve(file string) (Config, error) {
	var cfg Config

	data, err := os.ReadFile(file)
	if err != nil {
		return cfg, fmt.Errorf("config: read %s: %w", file, err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("config: parse %s: %w", file, err)
	}

	for _, src := range cfg.Plugins.Sources {
		if src.Local == "" {
			continue // git source: needs a clone, out of scope here (ROADMAP.md v0.2)
		}
		dir := filepath.Join(filepath.Dir(file), src.Local)
		if err := mergePluginRepo(&cfg, dir); err != nil {
			return cfg, fmt.Errorf("config: resolve plugin source %q: %w", src.ID, err)
		}
	}

	return cfg, nil
}

// mergePluginRepo walks every plugin.yaml under dir's category subdirs and merges its
// definitions into cfg.
func mergePluginRepo(cfg *Config, dir string) error {
	for _, category := range pluginCategories {
		matches, err := filepath.Glob(filepath.Join(dir, category, "*", "plugin.yaml"))
		if err != nil {
			return err
		}
		for _, path := range matches {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			var pf pluginFile
			if err := yaml.Unmarshal(data, &pf); err != nil {
				return fmt.Errorf("parse %s: %w", path, err)
			}
			cfg.Downloads = append(cfg.Downloads, pf.Downloads...)
			cfg.Tools = append(cfg.Tools, pf.Tools.Definitions...)
			cfg.Lint.Definitions = append(cfg.Lint.Definitions, pf.Lint.Definitions...)
			cfg.Actions.Definitions = append(cfg.Actions.Definitions, pf.Actions.Definitions...)
			cfg.Runtimes.Definitions = append(cfg.Runtimes.Definitions, pf.Runtimes.Definitions...)
		}
	}
	return nil
}
