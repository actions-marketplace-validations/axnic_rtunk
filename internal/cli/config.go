package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"

	"github.com/xunleii/rtunk/pkg/trunk/config"
)

// configCmd is `rtunk config`: currently just print, the fully resolved configuration.
type configCmd struct {
	Print printCmd `cmd:"" help:"Print the fully resolved configuration."`
}

type printCmd struct {
	Output string `help:"Output format." enum:"yaml,json" default:"yaml"`
	All    bool   `help:"Print the full merged plugin catalog instead of only enabled+used."`
}

func (c *printCmd) Run(cli *CLI, stdout io.Writer) error {
	cfg, err := resolveConfig(cli.Config, cli.CacheDir, c.All)
	if err != nil {
		return err
	}
	return printValue(stdout, cfg, c.Output)
}

// resolveConfig finds (unless configPath is already set) and resolves the trunk.yaml in effect.
// all selects ResolveAll (the full merged catalog) over Resolve (enabled+used only) -- only
// `config print --all` sets it.
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

// printValue marshals v as YAML (every definition's field tags, e.g. Linter/Tool/Runtime, are
// already YAML tags matching plugin.yaml's own vocabulary) and, for --output json, round-trips
// that through yaml.Unmarshal into a generic any -- yaml.v3 decodes mappings into
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
