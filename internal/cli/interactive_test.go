package cli

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testChecklistItems() []listItem {
	return []listItem{
		{ID: "alpha", label: "a"},
		{ID: "beta", label: "b"},
		{ID: "gamma", label: "g"},
	}
}

func TestRunChecklist_SpaceTogglesCursorItem(t *testing.T) {
	s := newChecklistState(testChecklistItems(), nil)
	var out bytes.Buffer

	confirmed, err := runChecklist(s, bufio.NewReader(strings.NewReader(" \r")), &out)
	require.NoError(t, err)
	assert.True(t, confirmed)
	assert.Equal(t, []string{"alpha"}, s.selected())
}

func TestRunChecklist_ArrowDownMovesCursorBeforeToggle(t *testing.T) {
	s := newChecklistState(testChecklistItems(), nil)
	var out bytes.Buffer

	// down arrow (ESC [ B), then space toggles "beta", then enter confirms.
	confirmed, err := runChecklist(s, bufio.NewReader(strings.NewReader("\x1b[B \r")), &out)
	require.NoError(t, err)
	assert.True(t, confirmed)
	assert.Equal(t, []string{"beta"}, s.selected())
}

func TestRunChecklist_ArrowUpWrapsToLastItem(t *testing.T) {
	s := newChecklistState(testChecklistItems(), nil)
	var out bytes.Buffer

	confirmed, err := runChecklist(s, bufio.NewReader(strings.NewReader("\x1b[A \r")), &out)
	require.NoError(t, err)
	assert.True(t, confirmed)
	assert.Equal(t, []string{"gamma"}, s.selected())
}

func TestRunChecklist_ToggleTwiceUnchecks(t *testing.T) {
	initial := map[string]bool{"alpha": true}
	s := newChecklistState(testChecklistItems(), initial)
	var out bytes.Buffer

	confirmed, err := runChecklist(s, bufio.NewReader(strings.NewReader(" \r")), &out)
	require.NoError(t, err)
	assert.True(t, confirmed)
	assert.Empty(t, s.selected())
}

func TestRunChecklist_QCancelsWithoutApplyingToggles(t *testing.T) {
	s := newChecklistState(testChecklistItems(), nil)
	var out bytes.Buffer

	confirmed, err := runChecklist(s, bufio.NewReader(strings.NewReader(" q")), &out)
	require.NoError(t, err)
	assert.False(t, confirmed)
}

func TestRunChecklist_BareEscCancels(t *testing.T) {
	s := newChecklistState(testChecklistItems(), nil)
	var out bytes.Buffer

	confirmed, err := runChecklist(s, bufio.NewReader(strings.NewReader("\x1bq")), &out)
	require.NoError(t, err)
	assert.False(t, confirmed)
}

func TestRunChecklist_CtrlCCancels(t *testing.T) {
	s := newChecklistState(testChecklistItems(), nil)
	var out bytes.Buffer

	confirmed, err := runChecklist(s, bufio.NewReader(strings.NewReader("\x03")), &out)
	require.NoError(t, err)
	assert.False(t, confirmed)
}

func TestRunChecklist_UnrecognizedKeyIgnored(t *testing.T) {
	s := newChecklistState(testChecklistItems(), nil)
	var out bytes.Buffer

	confirmed, err := runChecklist(s, bufio.NewReader(strings.NewReader("z \r")), &out)
	require.NoError(t, err)
	assert.True(t, confirmed)
	assert.Equal(t, []string{"alpha"}, s.selected())
}

func TestDiffSelection_AddsAndRemoves(t *testing.T) {
	initial := map[string]bool{"alpha": true, "beta": true}
	toAdd, toRemove := diffSelection(initial, []string{"alpha", "gamma"})
	assert.Equal(t, []string{"gamma"}, toAdd)
	assert.Equal(t, []string{"beta"}, toRemove)
}

func TestDiffSelection_NoChange(t *testing.T) {
	initial := map[string]bool{"alpha": true}
	toAdd, toRemove := diffSelection(initial, []string{"alpha"})
	assert.Empty(t, toAdd)
	assert.Empty(t, toRemove)
}

func TestInteractiveChecklist_NonTerminal_ErrorsClearly(t *testing.T) {
	// In tests (and CI) stdin/stdout aren't a real TTY, so interactiveChecklist must refuse to
	// run rather than hang reading raw key bytes from whatever is actually attached.
	_, ok, err := interactiveChecklist("pick:", testChecklistItems(), nil)
	require.Error(t, err)
	assert.False(t, ok)
	assert.Contains(t, err.Error(), "terminal")
}

func TestFlattenListing_SortsAndMarksEnabled(t *testing.T) {
	l := listing{
		Enabled:   []listItem{{ID: "beta"}},
		Available: []listItem{{ID: "alpha"}},
		Other:     []listItem{{ID: "gamma"}},
	}
	items, checked := flattenListing(l)

	require.Len(t, items, 3)
	assert.Equal(t, []string{"alpha", "beta", "gamma"}, []string{items[0].ID, items[1].ID, items[2].ID})
	assert.Equal(t, map[string]bool{"beta": true}, checked)
}
