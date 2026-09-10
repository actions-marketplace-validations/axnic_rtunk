package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFilterEnabled exercises the reference graph filterEnabled walks directly on a hand-built
// Config, independent of any plugin.yaml fixture: an enabled linter pulls in its tool, that tool's
// runtime and download, an enabled action pulls in its own runtime, and that runtime's download —
// while everything unreferenced (an unrelated tool/runtime/download/action/linter) is dropped.
func TestFilterEnabled(t *testing.T) {
	cfg := Config{
		Lint: CategoryConfig[Linter]{
			Enabled: []string{"eslint@8.10.0"}, // pinned: filterEnabled must strip the @version
			Definitions: map[string]Linter{
				"eslint": {Name: "eslint", Tools: []string{"eslint-bin"}},
				"unused": {Name: "unused", Tools: []string{"orphan-tool"}},
			},
		},
		Actions: CategoryConfig[Action]{
			Enabled: []string{"commitlint"},
			Definitions: map[string]Action{
				"commitlint": {ID: "commitlint", Runtime: "python"},
				"unused":     {ID: "unused", Runtime: "go"},
			},
		},
		Runtimes: CategoryConfig[Runtime]{
			// "node" is never in an enabled: list — it's only reachable via eslint-bin's tool
			// definition below, proving the tool -> runtime hop works.
			Definitions: map[string]Runtime{
				"node":   {Type: "node", Download: "node"},
				"python": {Type: "python", Download: "python"},
				"go":     {Type: "go", Download: "go"},
			},
		},
		Tools: map[string]Tool{
			"eslint-bin":  {Name: "eslint-bin", Runtime: "node"},
			"orphan-tool": {Name: "orphan-tool", Download: "orphan-download"},
		},
		Downloads: map[string]Download{
			"node":            {Name: "node"},
			"python":          {Name: "python"},
			"go":              {Name: "go"},
			"orphan-download": {Name: "orphan-download"},
		},
	}

	filterEnabled(&cfg)

	assert.Equal(t, map[string]Linter{"eslint": cfg.Lint.Definitions["eslint"]}, cfg.Lint.Definitions)
	assert.Equal(t, map[string]Action{"commitlint": cfg.Actions.Definitions["commitlint"]}, cfg.Actions.Definitions)
	assert.Equal(t, map[string]Tool{"eslint-bin": cfg.Tools["eslint-bin"]}, cfg.Tools)
	assert.ElementsMatch(t, []string{"node", "python"}, keysOf(cfg.Runtimes.Definitions))
	assert.ElementsMatch(t, []string{"node", "python"}, keysOf(cfg.Downloads))
}

func keysOf[T any](m map[string]T) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
