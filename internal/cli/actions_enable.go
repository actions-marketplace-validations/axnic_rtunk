package cli

// actionsEnableCmd is `rtunk actions enable <id>...`. With no id, it opens an interactive
// checklist over the full catalog (same items as `actions list --all`) instead.
type actionsEnableCmd struct {
	ID []string `arg:"" optional:"" help:"Action id(s) to enable. Omit for an interactive picker."`
}

func (c *actionsEnableCmd) Run(cli *CLI) error {
	if len(c.ID) > 0 {
		if err := rejectUnknownIDs(cli, "action", c.ID); err != nil {
			return err
		}
		return editActionsEnabled(cli, c.ID, true)
	}
	return interactiveActionsEnable(cli)
}

func interactiveActionsEnable(cli *CLI) error {
	cfg, err := resolveConfig(cli.Config, cli.CacheDir, true)
	if err != nil {
		return err
	}

	items, checked := flattenListing(buildActionsList(cfg), "action")
	selected, ok, err := interactiveChecklist("Select actions to enable:", items, checked)
	if err != nil || !ok {
		return err
	}

	toAdd, toRemove := diffSelection(checked, selected)
	return editActionsInteractive(cli, toAdd, toRemove)
}
