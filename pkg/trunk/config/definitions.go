package config

import "gopkg.in/yaml.v3"

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
