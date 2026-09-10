// Command rtunk is the CLI entry point. See ROADMAP.md for the staged command set; this file
// currently implements v0.1 ("read and query an existing trunk configuration") only.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/xunleii/rtunk/pkg/trunk/config"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, "rtunk:", err)
		os.Exit(1)
	}
}

const topUsage = `usage: rtunk [--config path] [--cache-dir path] <command>

commands:
  config {plugins,lint,actions,tools,runtimes} list [--enabled]
  config {plugins,lint,actions,tools,runtimes} show <id> [--output yaml|json]
  config print [--output yaml|json] [--all]
`

func run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("rtunk", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { fmt.Fprint(stderr, topUsage) }
	configPath := fs.String("config", "", "path to trunk.yaml (default: nearest .trunk/trunk.yaml)")
	cacheDir := fs.String("cache-dir", os.Getenv("RTUNK_CACHE_DIR"), "plugin cache directory (default: OS cache dir)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	args = fs.Args()
	if len(args) == 0 {
		fs.Usage()
		return fmt.Errorf("no command given")
	}

	switch args[0] {
	case "config":
		return runConfig(*configPath, *cacheDir, args[1:], stdout, stderr)
	default:
		fs.Usage()
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runConfig(configPath, cacheDir string, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: rtunk config {plugins,lint,actions,tools,runtimes} {list,show} | rtunk config print")
	}

	if args[0] == "print" {
		return runConfigPrint(configPath, cacheDir, args[1:], stdout, stderr)
	}

	cat, ok := categories[args[0]]
	if !ok {
		return fmt.Errorf("unknown config category %q (want one of plugins, lint, actions, tools, runtimes)", args[0])
	}
	if len(args) < 2 {
		return fmt.Errorf("usage: rtunk config %s {list,show}", args[0])
	}

	switch args[1] {
	case "list":
		return runConfigList(configPath, cacheDir, cat, args[2:], stdout, stderr)
	case "show":
		return runConfigShow(configPath, cacheDir, cat, args[2:], stdout, stderr)
	default:
		return fmt.Errorf("unknown subcommand %q (want list or show)", args[1])
	}
}

func runConfigList(configPath, cacheDir string, cat category, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("rtunk config list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	enabledOnly := fs.Bool("enabled", false, "only list elements that are actually turned on")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := resolveConfig(configPath, cacheDir, false)
	if err != nil {
		return err
	}

	ids := cat.ids(cfg)
	sort.Strings(ids)
	if *enabledOnly {
		enabled := cat.enabled(cfg)
		filtered := ids[:0]
		for _, id := range ids {
			if _, ok := enabled[id]; ok {
				filtered = append(filtered, id)
			}
		}
		ids = filtered
	}

	for _, id := range ids {
		fmt.Fprintln(stdout, id)
	}
	return nil
}

func runConfigShow(configPath, cacheDir string, cat category, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("rtunk config show", flag.ContinueOnError)
	fs.SetOutput(stderr)
	output := fs.String("output", "yaml", "output format: yaml|json")
	// flag.Parse stops at the first non-flag token, so `show <id> --output json` (flag after the
	// positional id, as ROADMAP.md's `show <id> [--output ...]` reads) would otherwise leave
	// --output unparsed; move recognized flags to the front first.
	if err := fs.Parse(reorderFlags(args, map[string]bool{"output": true})); err != nil {
		return err
	}
	rest := fs.Args()
	if len(rest) != 1 {
		return fmt.Errorf("usage: rtunk config <category> show <id> [--output yaml|json]")
	}

	cfg, err := resolveConfig(configPath, cacheDir, false)
	if err != nil {
		return err
	}

	v, ok := cat.show(cfg, rest[0])
	if !ok {
		return fmt.Errorf("%q not found", rest[0])
	}
	return printValue(stdout, v, *output)
}

func runConfigPrint(configPath, cacheDir string, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("rtunk config print", flag.ContinueOnError)
	fs.SetOutput(stderr)
	output := fs.String("output", "yaml", "output format: yaml|json")
	all := fs.Bool("all", false, "print the full merged plugin catalog instead of only enabled+used")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := resolveConfig(configPath, cacheDir, *all)
	if err != nil {
		return err
	}
	return printValue(stdout, cfg, *output)
}

// resolveConfig finds (unless configPath is already set) and resolves the trunk.yaml in effect.
// all selects ResolveAll (the full merged catalog) over Resolve (enabled+used only); only `config
// print --all` sets it — list/show always operate on the trimmed config.
func resolveConfig(configPath, cacheDir string, all bool) (config.Config, error) {
	if configPath == "" {
		found, err := findTrunkYAML()
		if err != nil {
			return config.Config{}, err
		}
		configPath = found
	}
	if all {
		return config.ResolveAll(configPath, cacheDir)
	}
	return config.Resolve(configPath, cacheDir)
}

// findTrunkYAML walks up from the working directory looking for .trunk/trunk.yaml, the same way
// git locates .git — the nearest match wins. rtunk's own .rtunk/rtunk.yaml is not read yet (see
// pkg/trunk/config's package doc).
func findTrunkYAML() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	start := dir
	for {
		candidate := filepath.Join(dir, ".trunk", "trunk.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .trunk/trunk.yaml found (searched from %s upward); use --config to specify one", start)
		}
		dir = parent
	}
}

// printValue marshals v as YAML (every definition's field tags, e.g. Linter/Tool/Runtime, are
// already YAML tags matching plugin.yaml's own vocabulary) and, for --output json, round-trips
// that through yaml.Unmarshal into a generic any — yaml.v3 decodes mappings into
// map[string]interface{}, so the result re-marshals as JSON with the same field names, no
// separate set of json tags to keep in sync.
func printValue(w io.Writer, v any, format string) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}

	switch format {
	case "", "yaml":
		_, err := w.Write(data)
		return err
	case "json":
		var generic any
		if err := yaml.Unmarshal(data, &generic); err != nil {
			return err
		}
		js, err := json.MarshalIndent(generic, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w, string(js))
		return err
	default:
		return fmt.Errorf("invalid --output %q: must be yaml or json", format)
	}
}

// category adapts one of trunk.yaml's five config categories (plugins/lint/actions/tools/
// runtimes) to the shared list/show behavior: ids lists everything defined, enabled narrows that
// to what's actually turned on, show looks up one definition by id.
type category struct {
	ids     func(config.Config) []string
	enabled func(config.Config) map[string]struct{}
	show    func(cfg config.Config, id string) (any, bool)
}

var categories = map[string]category{
	"plugins": {
		ids: func(cfg config.Config) []string { return keys(cfg.Plugins.Sources) },
		// Every entry in plugins.sources is unconditionally in effect — trunk.yaml has no notion
		// of a defined-but-disabled source — so --enabled is a no-op here: everything defined is
		// also enabled.
		enabled: func(cfg config.Config) map[string]struct{} { return set(keys(cfg.Plugins.Sources)) },
		show: func(cfg config.Config, id string) (any, bool) {
			v, ok := cfg.Plugins.Sources[id]
			return v, ok
		},
	},
	"lint": {
		ids:     func(cfg config.Config) []string { return keys(cfg.Lint.Definitions) },
		enabled: func(cfg config.Config) map[string]struct{} { return enabledIDs(cfg.Lint.Enabled) },
		show: func(cfg config.Config, id string) (any, bool) {
			v, ok := cfg.Lint.Definitions[id]
			return v, ok
		},
	},
	"actions": {
		ids:     func(cfg config.Config) []string { return keys(cfg.Actions.Definitions) },
		enabled: func(cfg config.Config) map[string]struct{} { return enabledIDs(cfg.Actions.Enabled) },
		show: func(cfg config.Config, id string) (any, bool) {
			v, ok := cfg.Actions.Definitions[id]
			return v, ok
		},
	},
	"runtimes": {
		ids:     func(cfg config.Config) []string { return keys(cfg.Runtimes.Definitions) },
		enabled: func(cfg config.Config) map[string]struct{} { return enabledIDs(cfg.Runtimes.Enabled) },
		show: func(cfg config.Config, id string) (any, bool) {
			v, ok := cfg.Runtimes.Definitions[id]
			return v, ok
		},
	},
	"tools": {
		ids: func(cfg config.Config) []string { return keys(cfg.Tools) },
		// Tools have no enabled: list of their own (ARCHITECTURE.md): a tool is "on" only by
		// being referenced from an enabled linter's tools: field.
		enabled: func(cfg config.Config) map[string]struct{} {
			out := map[string]struct{}{}
			for id := range enabledIDs(cfg.Lint.Enabled) {
				for _, t := range cfg.Lint.Definitions[id].Tools {
					out[t] = struct{}{}
				}
			}
			return out
		},
		show: func(cfg config.Config, id string) (any, bool) {
			v, ok := cfg.Tools[id]
			return v, ok
		},
	},
}

func keys[T any](m map[string]T) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func set(ids []string) map[string]struct{} {
	out := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		out[id] = struct{}{}
	}
	return out
}

// reorderFlags moves every recognized --name / --name=value / --name value token in args to the
// front (preserving their relative order), leaving every other token — positional arguments — at
// the back in their original order. flag.FlagSet.Parse stops at the first non-flag token, so this
// is what lets a flag be written after a positional argument (e.g. `show <id> --output json`).
// wantsValue says, per flag name (without leading dashes), whether it consumes the next token as
// its value when not written as --name=value.
func reorderFlags(args []string, wantsValue map[string]bool) []string {
	var flags, positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if !strings.HasPrefix(a, "-") {
			positional = append(positional, a)
			continue
		}
		flags = append(flags, a)
		name, _, hasValue := strings.Cut(strings.TrimLeft(a, "-"), "=")
		if !hasValue && wantsValue[name] && i+1 < len(args) {
			i++
			flags = append(flags, args[i])
		}
	}
	return append(flags, positional...)
}

// enabledIDs turns a trunk.yaml enabled: list (each entry optionally pinned as `id@version`) into
// a bare-id set, matching how config.Validate reads the same lists.
func enabledIDs(enabled []string) map[string]struct{} {
	out := make(map[string]struct{}, len(enabled))
	for _, e := range enabled {
		id, _, _ := strings.Cut(e, "@")
		out[id] = struct{}{}
	}
	return out
}
