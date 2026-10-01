package cli

import (
	"fmt"
	"strings"
)

// lintersEnableCmd is `rtunk linters enable <id>[@version]...`. With no id, it opens an
// interactive checklist over the full catalog (same items as `linters list --all`) instead.
type lintersEnableCmd struct {
	ID []string `arg:"" optional:"" help:"Linter id(s) to enable, optionally @version. Omit for an interactive picker."`
}

func (c *lintersEnableCmd) Run(cli *CLI) error {
	if len(c.ID) > 0 {
		if err := rejectUnknownIDs(cli, "linter", c.ID); err != nil {
			return err
		}
		return editEnabled(cli, "lint", func(existing []string) []string {
			return addEnabled(existing, c.ID)
		})
	}
	return interactiveLintersEnable(cli)
}

func interactiveLintersEnable(cli *CLI) error {
	cfg, err := resolveConfig(cli.Config, cli.CacheDir, true)
	if err != nil {
		return err
	}
	repoRoot, err := logsRepoRoot(cli)
	if err != nil {
		return err
	}
	files, err := repoFiles(repoRoot)
	if err != nil {
		return err
	}

	items, checked := flattenListing(buildLintersList(cfg, repoRoot, files))
	selected, ok, err := interactiveChecklist("Select linters to enable:", items, checked)
	if err != nil || !ok {
		return err
	}

	toAdd, toRemove := diffSelection(checked, selected)
	return editEnabled(cli, "lint", func(existing []string) []string {
		return addEnabled(removeEnabled(existing, toRemove), toAdd)
	})
}

// rejectUnknownIDs errors, naming every id (version suffix ignored) no plugin defines, when
// enabling kind ("linter" or "action") by explicit id -- `linters list`/`actions list` read the
// same resolved catalog (config.ResolveAll), so the check can't drift from what they show.
func rejectUnknownIDs(cli *CLI, kind string, ids []string) error {
	cfg, err := resolveConfig(cli.Config, cli.CacheDir, true)
	if err != nil {
		return err
	}
	var unknown []string
	for _, id := range ids {
		name, _, _ := cutVersion(id)
		var ok bool
		if kind == "linter" {
			_, ok = cfg.Lint.Definitions[name]
		} else {
			_, ok = cfg.Actions.Definitions[name]
		}
		if !ok {
			unknown = append(unknown, name)
		}
	}
	if len(unknown) > 0 {
		return fmt.Errorf("unknown %s id(s): %s (run `rtunk %ss list --all` to see available ids)", kind, strings.Join(unknown, ", "), kind)
	}
	return nil
}
