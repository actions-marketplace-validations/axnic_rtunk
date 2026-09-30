package cli

// lintersEnableCmd is `rtunk linters enable <id>[@version]...`. With no id, it opens an
// interactive checklist over the full catalog (same items as `linters list --all`) instead.
type lintersEnableCmd struct {
	ID []string `arg:"" optional:"" help:"Linter id(s) to enable, optionally @version. Omit for an interactive picker."`
}

func (c *lintersEnableCmd) Run(cli *CLI) error {
	if len(c.ID) > 0 {
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
