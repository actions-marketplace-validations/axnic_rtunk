package render

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axnic/rtunk/pkg/run/engine"
	"github.com/axnic/rtunk/pkg/trunk/output"
)

func githubRun(t *testing.T, opts Options, findings []output.Finding, s Summary) string {
	t.Helper()
	opts.Format, opts.Command, opts.NoProgress = GitHub, Check, true
	events := []engine.Event{{Linter: "lint", Phase: engine.Done, Files: []string{"a.go"}, Findings: findings}}
	stdout, _ := run(t, opts, events, s)
	return stdout
}

func TestGitHub_AnnotationPerFinding(t *testing.T) {
	tests := []struct {
		name string
		f    output.Finding
		want string
	}{
		{"error with full location", output.Finding{File: "a.go", Line: 5, Column: 3, Severity: "error", RuleID: "r1", Message: "m"},
			"::error file=a.go,line=5,col=3,title=lint/r1::m"},
		{"warning without rule", output.Finding{File: "a.go", Line: 5, Severity: "warning", Message: "m"},
			"::warning file=a.go,line=5,title=lint::m"},
		{"info is a notice", output.Finding{File: "a.go", Line: 1, Severity: "info", Message: "m"},
			"::notice file=a.go,line=1,title=lint::m"},
		{"unknown severity is a notice", output.Finding{File: "a.go", Line: 1, Severity: "weird", Message: "m"},
			"::notice file=a.go,line=1,title=lint::m"},
		{"column without line is dropped", output.Finding{File: "a.go", Column: 4, Severity: "error", Message: "m"},
			"::error file=a.go,title=lint::m"},
		{"no line", output.Finding{File: "a.go", Severity: "error", Message: "m"},
			"::error file=a.go,title=lint::m"},
		{"no file is a file-less annotation", output.Finding{Line: 3, Column: 2, Severity: "error", RuleID: "r", Message: "m"},
			"::error title=lint/r::m"},
		{"message escapes", output.Finding{File: "a.go", Line: 1, Severity: "error", Message: "100%\r\nnext: a,b"},
			"::error file=a.go,line=1,title=lint::100%25%0D%0Anext: a,b"},
		{"property escapes", output.Finding{File: "dir/a:b,c%.go", Line: 1, Severity: "error", RuleID: "x:y,z", Message: "m"},
			"::error file=dir/a%3Ab%2Cc%25.go,line=1,title=lint/x%3Ay%2Cz::m"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want+"\n", githubRun(t, Options{}, []output.Finding{tc.f}, Summary{}))
		})
	}
}

func TestGitHub_PathsAreWorkspaceRelativeWithForwardSlashes(t *testing.T) {
	ws := filepath.Join(string(filepath.Separator), "ws")
	tests := []struct {
		name string
		opts Options
		file string
		want string
	}{
		{"relative to the root, root is the workspace", Options{Root: ws}, filepath.Join("pkg", "a.go"), "pkg/a.go"},
		{"root below the workspace", Options{Root: filepath.Join(ws, "sub"), Workspace: ws}, "a.go", "sub/a.go"},
		{"absolute path under the workspace", Options{Workspace: ws}, filepath.Join(ws, "pkg", "a.go"), "pkg/a.go"},
		{"outside the workspace stays as reported", Options{Root: filepath.Join(ws, "sub"), Workspace: filepath.Join(ws, "other")}, "a.go", "a.go"},
		{"no root, no workspace", Options{}, "pkg/a.go", "pkg/a.go"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := githubRun(t, tc.opts, []output.Finding{{File: tc.file, Line: 1, Severity: "error", Message: "m"}}, Summary{})
			assert.Equal(t, "::error file="+tc.want+",line=1,title=lint::m\n", out)
		})
	}
}

func TestGitHub_FailuresAndSkippedSurface(t *testing.T) {
	events := []engine.Event{{Linter: "gitleaks", Phase: engine.Failed, Err: errors.New("timed out: 30s\nsecond")}}
	stdout, _ := run(t, Options{Format: GitHub, Command: Check, NoProgress: true}, events,
		Summary{Skipped: []string{"shfmt [no runtime]"}, Failures: []Failure{{Linter: "prettier", Err: "crashed"}}})
	assert.Equal(t, "::error title=gitleaks::linter failed to run: timed out: 30s\n"+
		"::error title=prettier::linter failed to run: crashed\n"+
		"::warning title=shfmt::linter skipped: no runtime\n", stdout)
}

func TestGitHub_EmptyResultWritesNothingToStdout(t *testing.T) {
	assert.Empty(t, githubRun(t, Options{}, nil, Summary{}))
}

func TestGitHub_FindingsAreSortedAndNeverTruncated(t *testing.T) {
	var fs []output.Finding
	for i := 100; i > 0; i-- {
		fs = append(fs, output.Finding{File: "a.go", Line: i, Severity: "error", Message: "m"})
	}
	lines := strings.Split(strings.TrimSpace(githubRun(t, Options{}, fs, Summary{})), "\n")
	require.Len(t, lines, 100)
	assert.Equal(t, "::error file=a.go,line=1,title=lint::m", lines[0])
}

func TestGitHub_StepSummary(t *testing.T) {
	findings := []output.Finding{
		{File: "b.md", Line: 2, Severity: "warning", RuleID: "MD001", Message: "a | b\nsecond <tag>"},
		{File: "a.md", Line: 1, Severity: "error", RuleID: "MD002", Message: "m"},
		{Severity: "info", Message: "no location"},
	}

	t.Run("appended to the file when the path is set", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "summary.md")
		require.NoError(t, os.WriteFile(path, []byte("previous\n"), 0o644))
		githubRun(t, Options{StepSummary: path}, findings, Summary{Skipped: []string{"shfmt [x]"}, Failures: []Failure{{Linter: "gitleaks", Err: "boom"}}})
		got, err := os.ReadFile(path)
		require.NoError(t, err)
		md := string(got)
		assert.True(t, strings.HasPrefix(md, "previous\n### rtunk check\n"))
		assert.Contains(t, md, "1 file checked by 1 linter in 0s: 3 findings, 1 failure, 0 suppressed.")
		assert.Contains(t, md, "| error | 1 |\n| warning | 1 |\n| notice | 1 |\n")
		assert.Contains(t, md, "| lint | 3 |\n")
		assert.Contains(t, md, "| error | `a.md:1` | lint/MD002 | m |\n")
		assert.Contains(t, md, "| warning | `b.md:2` | lint/MD001 | a \\| b second &lt;tag> |")
		assert.Contains(t, md, "**Failed linters**\n\n- gitleaks: boom\n")
		assert.Contains(t, md, "**Skipped linters**\n\n- shfmt [x]\n")
	})

	t.Run("nothing written when the path is empty", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)
		githubRun(t, Options{}, findings, Summary{})
		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		assert.Empty(t, entries)
	})

	t.Run("empty result", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "summary.md")
		githubRun(t, Options{StepSummary: path}, nil, Summary{})
		got, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Contains(t, string(got), "No findings.\n")
		assert.NotContains(t, string(got), "| Severity")
	})

	t.Run("table is capped with an N more line", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "summary.md")
		var many []output.Finding
		for i := 1; i <= summaryRows+7; i++ {
			many = append(many, output.Finding{File: "a.go", Line: i, Severity: "error", Message: "m"})
		}
		githubRun(t, Options{StepSummary: path}, many, Summary{})
		got, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, summaryRows, strings.Count(string(got), "| error | `a.go:"))
		assert.Contains(t, string(got), "7 more not shown.")
	})

	t.Run("an unwritable path is reported by Close", func(t *testing.T) {
		r := New(&strings.Builder{}, &strings.Builder{}, Options{Format: GitHub, StepSummary: filepath.Join(t.TempDir(), "missing", "s.md")})
		assert.Error(t, r.Close(Summary{}))
	})
}
