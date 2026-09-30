package cli

import (
	"bytes"
	"os"

	"gopkg.in/yaml.v3"
)

// editActionsEnabled keeps actions.enabled/actions.disabled in sync in one file write: enabling
// adds to enabled and removes from disabled (and vice versa for disabling), mirroring what the
// user's own real trunk CLI does to this repo's .trunk/trunk.yaml.
func editActionsEnabled(cli *CLI, ids []string, enable bool) error {
	return editActionsLists(cli, func(enabledList, disabledList []string) ([]string, []string) {
		if enable {
			return addEnabled(enabledList, ids), removeEnabled(disabledList, ids)
		}
		return removeEnabled(enabledList, ids), addEnabled(disabledList, ids)
	})
}

// editActionsInteractive applies an interactiveChecklist result: toEnable/toDisable are the ids
// whose checked state changed (see diffSelection) -- everything else in actions.enabled/disabled
// is left untouched.
func editActionsInteractive(cli *CLI, toEnable, toDisable []string) error {
	return editActionsLists(cli, func(enabledList, disabledList []string) ([]string, []string) {
		enabledList = addEnabled(removeEnabled(enabledList, toDisable), toEnable)
		disabledList = addEnabled(removeEnabled(disabledList, toEnable), toDisable)
		return enabledList, disabledList
	})
}

// editActionsLists loads the trunk.yaml in effect, hands actions.enabled/disabled (as plain
// string slices) to mutate, and writes its result back to the same file. Reuses linters_edit.go's
// own findOrCreateMapKey/detectIndentWidth/addEnabled/removeEnabled (same package, unexported)
// rather than editEnabled itself, which only ever touches one list.
func editActionsLists(cli *CLI, mutate func(enabledList, disabledList []string) (newEnabled, newDisabled []string)) error {
	configPath := cli.Config
	if configPath == "" {
		found, err := findConfig()
		if err != nil {
			return err
		}
		configPath = found
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return err
	}
	if len(doc.Content) == 0 {
		doc.Kind = yaml.DocumentNode
		doc.Content = []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}}
	}
	root := doc.Content[0]
	catNode := findOrCreateMapKey(root, "actions")

	enabledNode := findOrCreateSeqKey(catNode, "enabled")
	disabledNode := findOrCreateSeqKey(catNode, "disabled")

	newEnabled, newDisabled := mutate(nodeStrings(enabledNode), nodeStrings(disabledNode))

	setNodeStrings(enabledNode, newEnabled)
	setNodeStrings(disabledNode, newDisabled)

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(detectIndentWidth(data))
	if err := enc.Encode(&doc); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}
	//nolint:gosec // trunk.yaml is a repo-tracked config file, readable like every other tracked file
	return os.WriteFile(configPath, buf.Bytes(), 0o644)
}

func findOrCreateSeqKey(mapping *yaml.Node, key string) *yaml.Node {
	n := findOrCreateMapKey(mapping, key)
	if n.Kind != yaml.SequenceNode {
		n.Kind = yaml.SequenceNode
		n.Tag = "!!seq"
		n.Content = nil
	}
	// Force block style even when the source used flow style (e.g. "enabled: []") -- an existing
	// flow-style sequence node otherwise keeps rendering inline (e.g. "enabled: [commitlint]")
	// since only Content is rebuilt above, not Style.
	n.Style = 0
	return n
}

func nodeStrings(n *yaml.Node) []string {
	out := make([]string, len(n.Content))
	for i, c := range n.Content {
		out[i] = c.Value
	}
	return out
}

func setNodeStrings(n *yaml.Node, values []string) {
	n.Content = make([]*yaml.Node, len(values))
	for i, v := range values {
		n.Content[i] = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v}
	}
}
