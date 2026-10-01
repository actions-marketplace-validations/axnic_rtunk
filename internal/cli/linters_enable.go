package cli

import (
	"fmt"
	"io"
	"strings"
)

// lintersEnableCmd is `rtunk linters enable <id>[@version]...`. With no id, it opens an
// interactive checklist over the full catalog (same items as `linters list --all`) instead.
type lintersEnableCmd struct {
	ID []string `arg:"" optional:"" help:"Linter id(s) to enable, optionally @version. Omit for an interactive picker."`
}

func (c *lintersEnableCmd) Run(cli *CLI, stdout io.Writer) error {
	if len(c.ID) > 0 {
		if err := rejectUnknownIDs(cli, "linter", c.ID); err != nil {
			return err
		}
		if err := editEnabled(cli, "lint", func(existing []string) []string {
			return addEnabled(existing, c.ID)
		}); err != nil {
			return err
		}
		reportEnabled(stdout, c.ID, nil)
		return nil
	}
	return interactiveLintersEnable(cli, stdout)
}

// reportEnabled prints what enabling changed and, for newly enabled linters, the command that
// downloads their tools now instead of on the first check.
func reportEnabled(w io.Writer, added, removed []string) {
	bare := make([]string, len(added))
	for i, id := range added {
		bare[i], _, _ = cutVersion(id)
	}
	if len(added) > 0 {
		_, _ = fmt.Fprintf(w, "Enabled: %s\n", strings.Join(bare, ", "))
	}
	if len(removed) > 0 {
		_, _ = fmt.Fprintf(w, "Disabled: %s\n", strings.Join(removed, ", "))
	}
	if len(added) > 0 {
		_, _ = fmt.Fprintf(w, "\nDownload them now with:\n  rtunk download lint %s\n", strings.Join(bare, " "))
	}
}

func interactiveLintersEnable(cli *CLI, stdout io.Writer) error {
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

	items, checked := flattenListing(buildLintersList(cfg, repoRoot, files), "linter")
	selected, ok, err := interactiveChecklist("Select linters to enable:", items, checked)
	if err != nil || !ok {
		return err
	}

	toAdd, toRemove := diffSelection(checked, selected)
	if err := editEnabled(cli, "lint", func(existing []string) []string {
		return addEnabled(removeEnabled(existing, toRemove), toAdd)
	}); err != nil {
		return err
	}
	reportEnabled(stdout, toAdd, toRemove)
	return nil
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
