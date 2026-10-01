package cli

import (
	"fmt"
	"io"
	"maps"
	"os"
	"slices"

	"github.com/xunleii/rtunk/internal/cli/render"
	"github.com/xunleii/rtunk/pkg/cache/download"
	"github.com/xunleii/rtunk/pkg/run/engine"
	"github.com/xunleii/rtunk/pkg/trunk/config"
)

// toolboxCategory maps the CLI's `runtime|tools` argument to the cache's category name.
func toolboxCategory(c string) string {
	if c == "runtime" {
		return "runtimes"
	}
	return c
}

// downloadCmd is `rtunk download [{runtime,tools,lint} <id>[@version]...]` (also `rtunk toolbox
// download`): the targeted items (`@version` parsed by cutVersion below), or with no argument
// every enabled tool and runtime, ahead of check/fmt/run downloading them lazily.
type downloadCmd struct {
	Category string   `arg:"" optional:"" enum:"runtime,tools,lint," default:"" help:"Resource category (runtime, tools or lint); omit to download everything enabled."`
	IDs      []string `arg:"" optional:"" name:"id" help:"Resource id(s), optionally @version."`
}

// Run resolves the trunk.yaml in effect and installs the requested refs (everything enabled when
// none), drawing the same live install view check and fmt show on a terminal -- plain
// "installing <item>" lines otherwise -- then a one-line summary.
func (c *downloadCmd) Run(cli *CLI, stdout io.Writer, stderr Stderr) error {
	// Explicit ids resolve against the full catalog: a linter just enabled may match no file yet,
	// and the enabled+used trim would hide it. A bare download keeps the trim (everything in use).
	cfg, err := resolveConfig(cli.Config, cli.CacheDir, c.Category != "")
	if err != nil {
		return err
	}
	refs, err := c.refs(cfg)
	if err != nil {
		return err
	}
	repoRoot, err := logsRepoRoot(cli)
	if err != nil {
		return err
	}

	planned := 0
	var progress render.Renderer = nopRenderer{}
	if f, ok := stderr.(*os.File); ok && render.LiveEnabled(isTerminal(stderr), false, os.Getenv("TERM")) {
		utf8 := render.IsUTF8Locale(os.Getenv("LC_ALL"), os.Getenv("LC_CTYPE"), os.Getenv("LANG"))
		progress = render.NewLive(progress, render.LiveOptions{
			Out: stderr, Size: func() (int, int) { return render.TermSize(f) }, ASCII: !utf8,
		})
	} else {
		progress = installLines{stderr}
	}
	failed, err := engine.Install(cfg, cli.CacheDir, repoRoot, refs, func(ev engine.Event) {
		if ev.Phase == engine.InstallPlanned {
			planned += ev.Total
		}
		progress.Event(ev)
	})
	_ = progress.Close(render.Summary{})
	if err != nil {
		return err
	}

	items := slices.Sorted(maps.Keys(failed))
	for _, item := range items {
		_, _ = fmt.Fprintf(stderr, "✖ %s: %v\n", item, failed[item])
	}
	switch {
	case planned == 0:
		_, _ = fmt.Fprintln(stdout, "Everything is already downloaded.")
	case len(failed) == 0:
		_, _ = fmt.Fprintf(stdout, "Downloaded %d item(s).\n", planned)
	default:
		return fmt.Errorf("%d of %d download(s) failed", len(failed), planned)
	}
	return nil
}

// refs is what to install: the given ids of c.Category (a linter expanding to its tools), or with
// no category every enabled tool and runtime.
func (c *downloadCmd) refs(cfg config.Config) ([]download.Ref, error) {
	var refs []download.Ref
	if c.Category == "" {
		for _, id := range slices.Sorted(maps.Keys(cfg.Tools)) {
			refs = append(refs, download.Ref{Category: "tools", ID: id})
		}
		for _, id := range slices.Sorted(maps.Keys(cfg.Runtimes.Definitions)) {
			refs = append(refs, download.Ref{Category: "runtimes", ID: id})
		}
		return refs, nil
	}
	if len(c.IDs) == 0 {
		return nil, fmt.Errorf("download %s: at least one id is required", c.Category)
	}
	for _, raw := range c.IDs {
		id, version, _ := cutVersion(raw)
		var known bool
		switch c.Category {
		case "lint":
			var linter config.Linter
			if linter, known = cfg.Lint.Definitions[id]; known {
				for _, tool := range linter.Tools {
					refs = append(refs, download.Ref{Category: "tools", ID: tool})
				}
			}
		case "tools":
			_, known = cfg.Tools[id]
		default:
			_, known = cfg.Runtimes.Definitions[id]
		}
		if !known {
			return nil, fmt.Errorf("download %s: unknown id %q", c.Category, id)
		}
		if c.Category != "lint" {
			refs = append(refs, download.Ref{Category: toolboxCategory(c.Category), ID: id, Version: version})
		}
	}
	return refs, nil
}

// nopRenderer is the report-less inner renderer under the live install view: download has no
// findings to report, only progress.
type nopRenderer struct{}

func (nopRenderer) Event(engine.Event)         {}
func (nopRenderer) Close(render.Summary) error { return nil }

// installLines is the non-terminal progress: one line per install started.
type installLines struct{ w io.Writer }

func (l installLines) Event(ev engine.Event) {
	if ev.Phase == engine.InstallStart {
		_, _ = fmt.Fprintf(l.w, "installing %s\n", ev.Item)
	}
}
func (installLines) Close(render.Summary) error { return nil }

// cutVersion splits an `id[@version]` CLI argument, mirroring how trunk.yaml's own enabled:
// entries are parsed (pkg/trunk/config's enabledIDs/checkEnabled).
func cutVersion(s string) (id, version string, pinned bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == '@' {
			return s[:i], s[i+1:], true
		}
	}
	return s, "", false
}
