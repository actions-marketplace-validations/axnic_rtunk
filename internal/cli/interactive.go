package cli

import (
	"fmt"
	"maps"
	"os"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/term"
)

// checklistState is the pure, testable state behind interactiveChecklist: which items are
// checked, the type-to-filter query, where the cursor sits among the visible (filtered) items,
// and which slice of rows the viewport shows. Kept separate from the terminal (checklistModel)
// so it can be unit tested without a real TTY.
type checklistState struct {
	title   string
	items   []listItem
	checked map[string]bool
	filter  string
	visible []int // indexes into items matching filter, in items' order
	cursor  int   // index into visible
	offset  int   // first row (see rows) drawn in the viewport
	height  int   // viewport rows of the last frame
}

func newChecklistState(items []listItem, initiallyChecked map[string]bool) *checklistState {
	checked := make(map[string]bool, len(initiallyChecked))
	maps.Copy(checked, initiallyChecked)
	s := &checklistState{items: items, checked: checked}
	s.refilter()
	return s
}

func (s *checklistState) refilter() {
	s.visible = s.visible[:0]
	for i, it := range s.items {
		if strings.Contains(it.ID, s.filter) {
			s.visible = append(s.visible, i)
		}
	}
	s.cursor, s.offset = 0, 0
}

func (s *checklistState) up() {
	if len(s.visible) > 0 {
		s.cursor = (s.cursor - 1 + len(s.visible)) % len(s.visible)
	}
}

func (s *checklistState) down() {
	if len(s.visible) > 0 {
		s.cursor = (s.cursor + 1) % len(s.visible)
	}
}

func (s *checklistState) toggle() {
	if len(s.visible) > 0 {
		id := s.items[s.visible[s.cursor]].ID
		s.checked[id] = !s.checked[id]
	}
}

// selected returns the ids left checked, in items' own (caller-sorted) order, filter ignored.
func (s *checklistState) selected() []string {
	out := make([]string, 0, len(s.items))
	for _, it := range s.items {
		if s.checked[it.ID] {
			out = append(out, it.ID)
		}
	}
	return out
}

// checklistRow is one viewport line: a visible item (item >= 0), a section header, or the blank
// line separating two sections.
type checklistRow struct {
	header string
	item   int // index into visible; -1 for a header or a blank
}

// rows lays the visible items out under a header each time their section changes, sections
// separated by a blank line like `* list` prints them.
func (s *checklistState) rows() []checklistRow {
	var out []checklistRow
	section := ""
	for vi, i := range s.visible {
		if sec := s.items[i].section; sec != "" && sec != section {
			if section != "" {
				out = append(out, checklistRow{item: -1})
			}
			out = append(out, checklistRow{header: sec, item: -1})
			section = sec
		}
		out = append(out, checklistRow{item: vi})
	}
	return out
}

// scroll moves the viewport the least needed to keep the cursor row (and its section header,
// when it is the section's first item) on screen.
func (s *checklistState) scroll(rows []checklistRow) {
	cur := 0
	for r, row := range rows {
		if row.item == s.cursor {
			cur = r
			break
		}
	}
	top := cur
	if cur > 0 && rows[cur-1].header != "" {
		top = cur - 1
	}
	s.offset = min(s.offset, top, max(len(rows)-s.height, 0))
	if cur >= s.offset+s.height {
		s.offset = cur - s.height + 1
	}
}

// frame is the picker's lines for a cols x termRows terminal: title, filter, as many list rows
// as fit, hint. Lines are cut to the width (color sequences kept intact) so none wraps.
func (s *checklistState) frame(cols, termRows int) []string {
	var out []string
	add := func(line string) {
		if cols > 0 {
			line = ansi.Truncate(line, cols-1, "")
		}
		out = append(out, line)
	}

	if s.title != "" {
		add(st.header.Render(s.title))
	}
	add(st.faint.Render("filter: ") + st.filter.Render(s.filter) + st.faint.Render(fmt.Sprintf("_  (%d/%d)", len(s.visible), len(s.items))))

	rows := s.rows()
	s.height = max(termRows-len(out)-1, 1)
	s.scroll(rows)
	width := 0
	for _, it := range s.items {
		width = max(width, len(it.name()))
	}
	end := min(s.offset+s.height, len(rows))
	for _, row := range rows[s.offset:end] {
		if row.item < 0 {
			if row.header != "" {
				add(st.header.Render(row.header))
			} else {
				add("")
			}
			continue
		}
		it := s.items[s.visible[row.item]]
		add(listLine(row.item == s.cursor, s.checked[it.ID], it, width))
	}

	more, below := "", 0
	for _, row := range rows[end:] {
		if row.item >= 0 {
			below++
		}
	}
	if below > 0 {
		more = fmt.Sprintf("↓ %d more · ", below)
	}
	add(st.faint.Render(more + "type to filter · space toggle · enter confirm · esc cancel"))
	return out
}

// checklistModel is the bubbletea program around s: bubbletea owns the raw terminal, key
// parsing and redraws (resizes included); the model only maps keys onto s.
type checklistModel struct {
	s          *checklistState
	cols, rows int
	confirmed  bool
}

func (m *checklistModel) Init() tea.Cmd { return nil }

// Update: enter confirms, esc/ctrl-c cancel, up/down move, space toggles, backspace edits the
// filter and any other typed text is appended to it.
func (m *checklistModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.cols, m.rows = msg.Width, msg.Height
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			m.confirmed = true
			return m, tea.Quit
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "up":
			m.s.up()
		case "down":
			m.s.down()
		case "space":
			m.s.toggle()
		case "backspace":
			if m.s.filter != "" {
				m.s.filter = m.s.filter[:len(m.s.filter)-1]
				m.s.refilter()
			}
		default:
			if msg.Text != "" {
				m.s.filter += msg.Text
				m.s.refilter()
			}
		}
	}
	return m, nil
}

func (m *checklistModel) View() tea.View {
	v := tea.NewView(strings.Join(m.s.frame(m.cols, m.rows), "\n"))
	// Alt screen, like fzf: the terminal comes back untouched on exit, whatever the key.
	v.AltScreen = true
	return v
}

// interactiveChecklist runs the picker and lets the user toggle items checked in
// place, starting from initiallyChecked. It returns the final checked ids and ok=true on enter,
// or ok=false (no error) on cancel. Requires a real terminal on both stdin and stdout -- rtunk's
// non-interactive commands stay fully deterministic; this is the one opt-in exception, and it
// refuses to run at all when it can't be interactive.
func interactiveChecklist(title string, items []listItem, initiallyChecked map[string]bool) (selected []string, ok bool, err error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return nil, false, fmt.Errorf("interactive mode requires a terminal; pass explicit id(s) instead")
	}
	if len(items) == 0 {
		return nil, false, fmt.Errorf("nothing to select")
	}

	s := newChecklistState(items, initiallyChecked)
	s.title = title
	m := &checklistModel{s: s}
	if _, err := tea.NewProgram(m).Run(); err != nil || !m.confirmed {
		return nil, false, err
	}
	return s.selected(), true, nil
}

// diffSelection compares selected against the ids that were initially checked, returning the
// ones newly checked (toAdd) and the ones newly unchecked (toRemove), both sorted for
// deterministic output.
func diffSelection(initiallyChecked map[string]bool, selected []string) (toAdd, toRemove []string) {
	selectedSet := make(map[string]bool, len(selected))
	for _, id := range selected {
		selectedSet[id] = true
		if !initiallyChecked[id] {
			toAdd = append(toAdd, id)
		}
	}
	for id, was := range initiallyChecked {
		if was && !selectedSet[id] {
			toRemove = append(toRemove, id)
		}
	}
	sort.Strings(toAdd)
	sort.Strings(toRemove)
	return toAdd, toRemove
}
