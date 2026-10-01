package cli

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/xunleii/rtunk/pkg/cache/download"
)

type cacheCleanCmd struct{}

// cacheSubtrees is the complete set of subtrees any rtunk code ever creates under the shared
// cache root (see download.Root, download.registryDir, and runlog's logsRoot): downloads,
// plugins, logs, registry. cache clean must remove only these, not the whole --cache-dir, since
// a user may point --cache-dir at a directory shared with other tools.
var cacheSubtrees = []string{"downloads", "plugins", "logs", "registry"}

// Run removes every existing subtree: on a terminal one line each, a spinner turning into a green
// check with the space freed; plain "removed" lines otherwise.
func (c *cacheCleanCmd) Run(cli *CLI, stdout io.Writer) error {
	downloadsRoot, err := download.Root(cli.CacheDir)
	if err != nil {
		return err
	}
	sharedRoot := filepath.Dir(downloadsRoot)
	var rows []*cleanRow
	for _, subtree := range cacheSubtrees {
		if _, err := os.Stat(filepath.Join(sharedRoot, subtree)); err == nil {
			rows = append(rows, &cleanRow{name: subtree, path: filepath.Join(sharedRoot, subtree)})
		}
	}
	if len(rows) == 0 {
		_, _ = fmt.Fprintln(stdout, "The cache is already empty.")
		return nil
	}

	if f, ok := stdout.(*os.File); ok && isTerminal(f) && stdinIsTerminal() {
		m := newCleanModel(rows)
		if _, err := tea.NewProgram(m).Run(); err != nil {
			return err
		}
	} else {
		for _, r := range rows {
			r.size, r.err = removeMeasured(r.path)
			r.done = true
			if r.err == nil {
				_, _ = fmt.Fprintf(stdout, "removed %s (%s)\n", r.name, humanSize(r.size))
			}
		}
	}
	for _, r := range rows {
		if r.err != nil {
			return r.err
		}
	}
	return nil
}

// cleanRow is one subtree being removed; size and err are set once it is done.
type cleanRow struct {
	name, path string
	size       int64
	done       bool
	err        error
}

// removeMeasured deletes path, returning the bytes its files took.
func removeMeasured(path string) (size int64, err error) {
	_ = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			if info, err := d.Info(); err == nil {
				size += info.Size()
			}
		}
		return nil
	})
	return size, os.RemoveAll(path)
}

// humanSize is n bytes in the largest 1024-based unit below it, one decimal past bytes.
func humanSize(n int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	v, u := float64(n), 0
	for v >= 1024 && u < len(units)-1 {
		v /= 1024
		u++
	}
	if u == 0 {
		return fmt.Sprintf("%d B", n)
	}
	return fmt.Sprintf("%.1f %s", v, units[u])
}

// cleanModel removes every row in parallel, drawing them inline; the final frame stays on screen.
type cleanModel struct {
	rows    []*cleanRow
	spin    spinner.Model
	left    int
	ok, bad lipgloss.Style
}

// cleanDoneMsg carries a removal's result back to Update, the only place rows change.
type cleanDoneMsg struct {
	row  *cleanRow
	size int64
	err  error
}

func newCleanModel(rows []*cleanRow) *cleanModel {
	return &cleanModel{
		rows: rows, left: len(rows),
		spin: spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Cyan))),
		ok:   lipgloss.NewStyle().Foreground(lipgloss.Green),
		bad:  lipgloss.NewStyle().Foreground(lipgloss.Red),
	}
}

func (m *cleanModel) Init() tea.Cmd {
	cmds := []tea.Cmd{m.spin.Tick}
	for _, r := range m.rows {
		cmds = append(cmds, func() tea.Msg {
			size, err := removeMeasured(r.path)
			return cleanDoneMsg{row: r, size: size, err: err}
		})
	}
	return tea.Batch(cmds...)
}

func (m *cleanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case cleanDoneMsg:
		msg.row.size, msg.row.err, msg.row.done = msg.size, msg.err, true
		if m.left--; m.left == 0 {
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *cleanModel) View() tea.View {
	width := 0
	for _, r := range m.rows {
		width = max(width, len(r.name))
	}
	var b strings.Builder
	for _, r := range m.rows {
		switch {
		case !r.done:
			fmt.Fprintf(&b, "%s %s\n", m.spin.View(), r.name)
		case r.err != nil:
			fmt.Fprintf(&b, "%s %-*s  %v\n", m.bad.Render("✖"), width, r.name, r.err)
		default:
			fmt.Fprintf(&b, "%s %-*s  %s freed\n", m.ok.Render("✔"), width, r.name, humanSize(r.size))
		}
	}
	return tea.NewView(b.String())
}
