package cli

import (
	"bufio"
	"fmt"
	"io"
	"maps"
	"os"
	"sort"

	"golang.org/x/term"
)

// checklistState is the pure, testable state behind interactiveChecklist: which items are
// checked and where the cursor sits. Kept separate from raw-terminal I/O so the key-handling
// logic (runChecklist below) can be unit tested with a plain io.Reader/io.Writer, no real TTY.
type checklistState struct {
	items   []listItem
	checked map[string]bool
	cursor  int
}

func newChecklistState(items []listItem, initiallyChecked map[string]bool) *checklistState {
	checked := make(map[string]bool, len(initiallyChecked))
	maps.Copy(checked, initiallyChecked)
	return &checklistState{items: items, checked: checked}
}

func (s *checklistState) up()   { s.cursor = (s.cursor - 1 + len(s.items)) % len(s.items) }
func (s *checklistState) down() { s.cursor = (s.cursor + 1) % len(s.items) }
func (s *checklistState) toggle() {
	id := s.items[s.cursor].ID
	s.checked[id] = !s.checked[id]
}

// selected returns the ids left checked, in items' own (caller-sorted) order.
func (s *checklistState) selected() []string {
	out := make([]string, 0, len(s.items))
	for _, it := range s.items {
		if s.checked[it.ID] {
			out = append(out, it.ID)
		}
	}
	return out
}

// runChecklist drives s from raw key bytes read off r, redrawing to w after every change, until
// the user confirms (enter, returns true) or cancels (q, esc, ctrl-c, or r hits EOF -- returns
// false). Arrow keys (ESC [ A / ESC [ B) and space (toggle) are the only other recognized input;
// anything else is ignored.
// ponytail: a lone Esc keypress is indistinguishable from the start of an arrow sequence without
// a read timeout, so a bare Esc blocks until the next key -- acceptable for a local, opt-in picker;
// add a timeout-based reader if that stall is ever reported as an actual complaint.
func runChecklist(s *checklistState, r *bufio.Reader, w io.Writer) (bool, error) {
	renderChecklist(w, s, false)
	for {
		b, err := r.ReadByte()
		if err != nil {
			return false, err
		}

		redraw := true
		switch b {
		case '\r', '\n':
			return true, nil
		case 'q', 3: // q, ctrl-c
			return false, nil
		case 27: // esc, or the start of an arrow-key escape sequence
			next, err := r.ReadByte()
			if err != nil || next != '[' {
				return false, nil
			}
			dir, err := r.ReadByte()
			if err != nil {
				return false, nil
			}
			switch dir {
			case 'A':
				s.up()
			case 'B':
				s.down()
			default:
				redraw = false
			}
		case ' ':
			s.toggle()
		default:
			redraw = false
		}

		if redraw {
			renderChecklist(w, s, true)
		}
	}
}

// renderChecklist draws s's full item list plus a hint line. reprint moves the cursor back up
// over the previous draw first, so repeated calls update in place instead of scrolling the
// terminal.
func renderChecklist(w io.Writer, s *checklistState, reprint bool) {
	if reprint {
		fmt.Fprintf(w, "\x1b[%dA", len(s.items)+1)
	}

	width := 0
	for _, it := range s.items {
		width = max(width, len(it.name()))
	}

	for i, it := range s.items {
		mark, pointer := " ", "  "
		if s.checked[it.ID] {
			mark = "x"
		}
		if i == s.cursor {
			pointer = "> "
		}
		fmt.Fprintf(w, "\x1b[2K%s[%s] %-*s  %s\r\n", pointer, mark, width, it.name(), it.label)
	}
	fmt.Fprint(w, "\x1b[2Kspace toggle · enter confirm · q/esc cancel\r\n")
}

// interactiveChecklist puts the terminal in raw mode and lets the user toggle items checked in
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

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = term.Restore(fd, oldState) }()

	fmt.Fprint(os.Stdout, title+"\r\n")
	s := newChecklistState(items, initiallyChecked)
	confirmed, err := runChecklist(s, bufio.NewReader(os.Stdin), os.Stdout)
	fmt.Fprint(os.Stdout, "\r\n")
	if err != nil || !confirmed {
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
