package cli

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

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

var (
	keySpace = tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	keyUp    = tea.KeyPressMsg{Code: tea.KeyUp}
	keyDown  = tea.KeyPressMsg{Code: tea.KeyDown}
	keyEnter = tea.KeyPressMsg{Code: tea.KeyEnter}
	keyBack  = tea.KeyPressMsg{Code: tea.KeyBackspace}
)

func stripAll(lines []string) []string {
	for i, l := range lines {
		lines[i] = ansi.Strip(l)
	}
	return lines
}

func keyText(s string) tea.KeyPressMsg { return tea.KeyPressMsg{Code: []rune(s)[0], Text: s} }

// press feeds msgs to a fresh model over items, returning it and whether the last msg quit.
func press(initial map[string]bool, msgs ...tea.Msg) (*checklistModel, bool) {
	m := &checklistModel{s: newChecklistState(testChecklistItems(), initial)}
	quit := false
	for _, msg := range msgs {
		_, cmd := m.Update(msg)
		quit = cmd != nil
	}
	return m, quit
}

func TestChecklistModel_SpaceTogglesCursorItem(t *testing.T) {
	m, quit := press(nil, keySpace, keyEnter)
	assert.True(t, quit)
	assert.True(t, m.confirmed)
	assert.True(t, m.View().AltScreen, "alt screen: the terminal comes back untouched on exit")
	assert.Equal(t, []string{"alpha"}, m.s.selected())
}

func TestChecklistModel_DownMovesCursorBeforeToggle(t *testing.T) {
	m, _ := press(nil, keyDown, keySpace, keyEnter)
	assert.Equal(t, []string{"beta"}, m.s.selected())
}

func TestChecklistModel_UpWrapsToLastItem(t *testing.T) {
	m, _ := press(nil, keyUp, keySpace, keyEnter)
	assert.Equal(t, []string{"gamma"}, m.s.selected())
}

func TestChecklistModel_ToggleTwiceUnchecks(t *testing.T) {
	m, _ := press(map[string]bool{"alpha": true}, keySpace, keyEnter)
	assert.Empty(t, m.s.selected())
}

func TestChecklistModel_EscAndCtrlCCancel(t *testing.T) {
	for _, k := range []tea.KeyPressMsg{{Code: tea.KeyEscape}, {Code: 'c', Mod: tea.ModCtrl}} {
		m, quit := press(nil, keySpace, k)
		assert.True(t, quit)
		assert.False(t, m.confirmed)
	}
}

func TestChecklistModel_FilterNarrowsThenToggles(t *testing.T) {
	// "gam" leaves only gamma under the cursor; backspacing back to "" keeps its check.
	m, _ := press(nil, keyText("gam"), keySpace, keyBack, keyBack, keyBack, keyEnter)
	assert.Equal(t, []string{"gamma"}, m.s.selected())
	assert.Len(t, m.s.visible, 3)
}

func TestChecklistModel_FilterWithNoMatchToggleIsNoop(t *testing.T) {
	m, _ := press(nil, keyText("zzz"), keySpace, keyDown, keyEnter)
	assert.True(t, m.confirmed)
	assert.Empty(t, m.s.selected())
}

func TestChecklistModel_ResizeShrinksView(t *testing.T) {
	m, _ := press(nil, tea.WindowSizeMsg{Width: 80, Height: 24})
	assert.Equal(t, 5, strings.Count(m.View().Content, "\n")+1, "filter + 3 items + hint")

	_, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 4})
	assert.Equal(t, 4, strings.Count(m.View().Content, "\n")+1, "never taller than the terminal")
}

func TestChecklistFrame_ViewportFollowsCursorAndFitsTerminal(t *testing.T) {
	items, checked := flattenListing(listing{
		Enabled:   []listItem{{ID: "a1"}, {ID: "a2"}},
		Available: []listItem{{ID: "b1"}, {ID: "b2"}, {ID: "b3"}},
	}, "linter")
	s := newChecklistState(items, checked)
	for range 4 { // cursor onto b3
		s.down()
	}

	// 6 terminal rows: filter + 4 list rows + hint.
	frame := stripAll(s.frame(80, 6))
	require.Len(t, frame, 6)
	assert.Equal(t, "> ◯ b3  ", frame[4])
	assert.NotContains(t, strings.Join(frame, "\n"), "a1")

	// Taller terminal (a resize): everything fits again, sections separated by a blank line.
	frame = stripAll(s.frame(80, 30))
	assert.Equal(t, []string{
		"filter: _  (5/5)",
		"Enabled",
		"  ✔ a1  ",
		"  ✔ a2  ",
		"",
		"Available for this repo (not enabled)",
		"  ◯ b1  ",
		"  ◯ b2  ",
		"> ◯ b3  ",
		"type to filter · space toggle · enter confirm · esc cancel",
	}, frame)
}

func TestChecklistFrame_CutsLinesToWidthAndDimsUnused(t *testing.T) {
	zero, two := 0, 2
	s := newChecklistState([]listItem{
		{ID: "used", Files: &two, label: "2 files"},
		{ID: "unused", Files: &zero, label: "0 files and a long tail"},
	}, nil)
	frame := s.frame(20, 24)
	assert.Equal(t, "> ◯ used    2 files", ansi.Strip(frame[1]))
	assert.NotContains(t, frame[1], "\x1b[2m2 files", "cursor row label never dimmed")
	assert.Equal(t, "  ◯ unused  0 files", ansi.Strip(frame[2]), "cut to 19 columns")
	assert.Contains(t, frame[2], "\x1b[2m", "unused linter dimmed")
}

func TestChecklistFrame_CheckedUnusedLinterShowsItsMark(t *testing.T) {
	zero := 0
	s := newChecklistState([]listItem{{ID: "a"}, {ID: "unused", Files: &zero}}, map[string]bool{"unused": true})

	frame := stripAll(s.frame(80, 24))
	assert.Equal(t, "  ✔ unused  ", frame[2], "toggling a dimmed linter must show")
}

func TestChecklistFrame_MoreCountsItemsNotRows(t *testing.T) {
	items, checked := flattenListing(listing{
		Enabled:   []listItem{{ID: "a1"}},
		Available: []listItem{{ID: "b1"}, {ID: "b2"}},
	}, "linter")
	s := newChecklistState(items, checked)

	// filter + 2 rows (Enabled, a1) + hint: below are a blank, a header and 2 items.
	frame := s.frame(80, 4)
	assert.Contains(t, frame[len(frame)-1], "↓ 2 more")
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

func TestFlattenListing_GroupsBySectionAndMarksEnabled(t *testing.T) {
	l := listing{
		Enabled:   []listItem{{ID: "beta"}},
		Available: []listItem{{ID: "alpha"}},
		Other:     []listItem{{ID: "gamma"}},
	}
	items, checked := flattenListing(l, "linter")

	require.Len(t, items, 3)
	assert.Equal(t, []string{"beta", "alpha", "gamma"}, []string{items[0].ID, items[1].ID, items[2].ID})
	assert.Equal(t, "Other", items[2].section)
	assert.Equal(t, map[string]bool{"beta": true}, checked)
}

func TestReportEnabled_SummaryAndDownloadHint(t *testing.T) {
	var out strings.Builder
	reportEnabled(&out, []string{"yamlfmt@1.0.0", "actionlint"}, []string{"checkov"})
	assert.Equal(t, "Enabled: yamlfmt, actionlint\n"+
		"Disabled: checkov\n"+
		"\n"+
		"Download them now with:\n"+
		"  rtunk download lint yamlfmt actionlint\n", out.String())

	out.Reset()
	reportEnabled(&out, nil, []string{"checkov"})
	assert.Equal(t, "Disabled: checkov\n", out.String(), "nothing enabled: no download hint")
}
