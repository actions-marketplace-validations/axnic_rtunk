package render

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/axnic/rtunk/pkg/trunk/output"
)

// githubRenderer writes the run as GitHub Actions workflow commands, one annotation per finding
// (https://docs.github.com/actions/reference/workflow-commands-for-github-actions), and, when
// Options.StepSummary names a file, appends a Markdown job summary to it. Every finding is emitted:
// GitHub itself displays only the first 10 errors and 10 warnings per step (50 annotations per job).
type githubRenderer struct{ base }

// summaryRows caps the findings table of the job summary; the rest is counted, not listed.
const summaryRows = 50

var (
	// Workflow-command data escapes %, CR and LF; properties also escape : and , (their delimiters).
	githubData = strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A")
	githubProp = strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A", ":", "%3A", ",", "%2C")
	// A Markdown table cell is one line and must not open an HTML tag.
	cellText = strings.NewReplacer("|", `\|`, "\r", " ", "\n", " ", "<", "&lt;")
)

// githubLevel maps output.Finding's error/warning/info onto an annotation command: an unknown
// severity is a notice.
func githubLevel(s string) string {
	switch s {
	case "error":
		return "error"
	case "warning":
		return "warning"
	}
	return "notice"
}

// annotation is one workflow command; props are already "key=value" pairs, in order.
func annotation(level string, props []string, message string) string {
	cmd := "::" + level
	if len(props) > 0 {
		cmd += " " + strings.Join(props, ",")
	}
	return cmd + "::" + githubData.Replace(message)
}

// prop is a "key=value" property with the value escaped.
func prop(key, value string) string { return key + "=" + githubProp.Replace(value) }

// workspacePath is file as GitHub wants it: relative to the workspace (Options.Workspace, else
// the root the finding paths are relative to), with forward slashes. A path that does not sit
// under that base is left as reported.
func (g *githubRenderer) workspacePath(file string) string {
	base := g.opts.Workspace
	if base == "" {
		base = g.opts.Root
	}
	if base != "" {
		abs := file
		if !filepath.IsAbs(abs) && g.opts.Root != "" {
			abs = filepath.Join(g.opts.Root, file)
		}
		if filepath.IsAbs(abs) {
			if rel, err := filepath.Rel(base, abs); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				file = rel
			}
		}
	}
	return filepath.ToSlash(file)
}

func (g *githubRenderer) Close(s Summary) error {
	sorted := g.sortedFindings()
	failures := g.sortedFailures(s.Failures...)

	for _, f := range sorted {
		var props []string
		if f.File != "" { // line and column mean nothing without a file
			props = append(props, prop("file", g.workspacePath(f.File)))
			if f.Line > 0 {
				props = append(props, prop("line", fmt.Sprint(f.Line)))
				if f.Column > 0 {
					props = append(props, prop("col", fmt.Sprint(f.Column)))
				}
			}
		}
		if t := linterRule(f); t != "" {
			props = append(props, prop("title", t))
		}
		_, _ = fmt.Fprintln(g.stdout, annotation(githubLevel(f.Severity), props, f.Message))
	}
	for _, f := range failures {
		_, _ = fmt.Fprintln(g.stdout, annotation("error", []string{prop("title", f.Linter)}, "linter failed to run: "+f.Err))
	}
	for _, sk := range s.Skipped { // "linter [note]"
		linter, note, _ := strings.Cut(sk, " ")
		_, _ = fmt.Fprintln(g.stdout, annotation("warning", []string{prop("title", linter)}, "linter skipped: "+strings.Trim(note, "[]")))
	}

	if g.opts.StepSummary == "" {
		return nil
	}
	f, err := os.OpenFile(g.opts.StepSummary, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	_, werr := f.WriteString(g.markdown(s, sorted, failures))
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	return werr
}

// markdown is the job summary: totals, counts per severity and per linter, a findings table capped
// at summaryRows, then the failed and skipped linters.
func (g *githubRenderer) markdown(s Summary, sorted []output.Finding, failures []Failure) string {
	var b strings.Builder
	b.WriteString("### rtunk check\n\n")
	fmt.Fprintf(&b, "%s checked by %s in %s: %s, %s, %d suppressed.\n\n",
		plural(len(g.files), "file"), plural(len(g.linters), "linter"), s.Elapsed.Round(time.Millisecond),
		plural(len(sorted), "finding"), plural(len(failures), "failure"), g.suppressed)

	if len(sorted) == 0 {
		b.WriteString("No findings.\n")
	} else {
		bySeverity := map[string]int{}
		byLinter := map[string]int{}
		for _, f := range sorted {
			bySeverity[githubLevel(f.Severity)]++
			byLinter[f.Linter]++
		}
		b.WriteString("| Severity | Count |\n| --- | --- |\n")
		for _, lvl := range []string{"error", "warning", "notice"} {
			if n := bySeverity[lvl]; n > 0 {
				fmt.Fprintf(&b, "| %s | %d |\n", lvl, n)
			}
		}
		linters := make([]string, 0, len(byLinter))
		for l := range byLinter {
			linters = append(linters, l)
		}
		sort.Strings(linters)
		b.WriteString("\n| Linter | Findings |\n| --- | --- |\n")
		for _, l := range linters {
			fmt.Fprintf(&b, "| %s | %d |\n", cellText.Replace(l), byLinter[l])
		}

		b.WriteString("\n| Severity | Location | Rule | Message |\n| --- | --- | --- | --- |\n")
		for i, f := range sorted {
			if i == summaryRows {
				fmt.Fprintf(&b, "\n%d more not shown.\n", len(sorted)-summaryRows)
				break
			}
			loc := g.workspacePath(f.File)
			if f.Line > 0 {
				loc += fmt.Sprintf(":%d", f.Line)
			}
			fmt.Fprintf(&b, "| %s | `%s` | %s | %s |\n", githubLevel(f.Severity), loc, cellText.Replace(linterRule(f)), cellText.Replace(f.Message))
		}
	}

	if len(failures) > 0 {
		b.WriteString("\n**Failed linters**\n\n")
		for _, f := range failures {
			fmt.Fprintf(&b, "- %s: %s\n", f.Linter, cellText.Replace(f.Err))
		}
	}
	if len(s.Skipped) > 0 {
		b.WriteString("\n**Skipped linters**\n\n")
		for _, sk := range s.Skipped {
			fmt.Fprintf(&b, "- %s\n", cellText.Replace(sk))
		}
	}
	return b.String()
}
