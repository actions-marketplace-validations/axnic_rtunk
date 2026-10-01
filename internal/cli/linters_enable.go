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

func (c *lintersEnableCmd) Run(cli *CLI, stdout io.Writer, stderr Stderr) error {
	added := c.ID
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
		added = warnOverrides(cli, stderr, c.ID, nil)
	} else {
		var err error
		if added, err = interactiveLintersEnable(cli, stdout, stderr); err != nil {
			return err
		}
	}
	if len(added) > 0 {
		_, _ = fmt.Fprintf(stdout, "\nDownload them now with:\n  rtunk download lint %s\n", strings.Join(bareIDs(added), " "))
	}
	return nil
}

// reportEnabled prints what enabling changed.
func reportEnabled(w io.Writer, added, removed []string) {
	if len(added) > 0 {
		_, _ = fmt.Fprintf(w, "Enabled: %s\n", strings.Join(bareIDs(added), ", "))
	}
	if len(removed) > 0 {
		_, _ = fmt.Fprintf(w, "Disabled: %s\n", strings.Join(removed, ", "))
	}
}

// warnOverrides says, on w, what an edit of the shared config file could not change because a
// local override file decides otherwise: a linter lint.disabled keeps off after being enabled,
// and one an override enables that stays on after being disabled. It returns added without the
// linters that stay off. Best effort: a config that no longer resolves prints nothing and
// returns added whole.
func warnOverrides(cli *CLI, w io.Writer, added, removed []string) (enabled []string) {
	cfg, err := resolveConfig(cli.Config, cli.CacheDir, true)
	if err != nil {
		return added
	}
	for _, e := range added {
		id, _, _ := cutVersion(e)
		if file, off := cfg.Lint.DisabledFrom[id]; off {
			_, _ = fmt.Fprintf(w, "warning: %s stays disabled: %s lists it under lint.disabled\n", id, file)
			continue
		}
		enabled = append(enabled, e)
	}
	on := enabledVersions(cfg.Lint.Enabled)
	for _, id := range bareIDs(removed) {
		_, still := on[id]
		if file, ok := cfg.Lint.EnabledFrom[id]; ok && still {
			_, _ = fmt.Fprintf(w, "warning: %s stays enabled: %s enables it (remove it there, or list it under lint.disabled)\n", id, file)
		}
	}
	return enabled
}

// bareIDs is ids without their @version suffix.
func bareIDs(ids []string) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i], _, _ = cutVersion(id)
	}
	return out
}

// interactiveLintersEnable runs the picker, writes the selection and reports it, returning the
// newly enabled ids that really are (see warnOverrides).
func interactiveLintersEnable(cli *CLI, stdout io.Writer, stderr Stderr) ([]string, error) {
	cfg, err := resolveConfig(cli.Config, cli.CacheDir, true)
	if err != nil {
		return nil, err
	}
	repoRoot, err := logsRepoRoot(cli)
	if err != nil {
		return nil, err
	}
	files, err := repoFiles(repoRoot)
	if err != nil {
		return nil, err
	}

	items, checked := flattenListing(buildLintersList(cfg, repoRoot, files), "linter")
	selected, ok, err := interactiveChecklist("Select linters to enable:", items, checked)
	if err != nil || !ok {
		return nil, err
	}

	toAdd, toRemove := diffSelection(checked, selected)
	if err := editEnabled(cli, "lint", func(existing []string) []string {
		return addEnabled(removeEnabled(existing, toRemove), toAdd)
	}); err != nil {
		return nil, err
	}
	reportEnabled(stdout, toAdd, toRemove)
	return warnOverrides(cli, stderr, toAdd, toRemove), nil
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
