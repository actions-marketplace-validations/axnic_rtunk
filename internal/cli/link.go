package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/xunleii/rtunk/pkg/cache/download"
	"github.com/xunleii/rtunk/pkg/run/runlog"
	"github.com/xunleii/rtunk/pkg/trunk/config"
)

// linkCmd is `rtunk toolbox link`: (re)builds .rtunk/{logs,tools/<tool-id>,plugins/<source-id>}
// as symlinks into rtunk's cache, plus a .rtunk/.gitignore covering them -- the same fixed local
// path real trunk's own .trunk/{tools,logs,plugins/...} gives shells, editors and other tooling.
// Unlike trunk, rtunk's tools/plugins cache is shared across every repo on the machine, not
// duplicated per repo (ROADMAP.md v0.11 "one cache root, one on-disk layout"), so
// .rtunk/tools/<id> links straight to that shared cache's shim for the version this repo
// actually resolves -- not a per-repo copy -- and .rtunk/plugins/* does the same for checkouts.
type linkCmd struct{}

func (c *linkCmd) Run(cli *CLI, stdout io.Writer) error {
	repoRoot, err := logsRepoRoot(cli)
	if err != nil {
		return err
	}
	// all=false: the enabled+used subset, the same trim `rtunk check`/`rtunk toolbox where` act
	// on -- .rtunk/tools should offer the tools this repo actually resolves, not the full
	// plugin-source catalog (most of which isn't even downloaded).
	cfg, err := resolveConfig(cli.Config, cli.CacheDir, false)
	if err != nil {
		return err
	}

	rtunkDir := filepath.Join(repoRoot, ".rtunk")
	if err := os.MkdirAll(rtunkDir, 0o755); err != nil {
		return err
	}

	logsDir, err := runlog.Dir(cli.CacheDir, repoRoot)
	if err != nil {
		return err
	}
	if err := relink(filepath.Join(rtunkDir, "logs"), logsDir); err != nil {
		return err
	}

	if err := linkTools(rtunkDir, cli.CacheDir, cfg); err != nil {
		return err
	}
	if err := linkPluginSources(rtunkDir, cli.CacheDir, cfg); err != nil {
		return err
	}
	if err := writeLinkGitignore(rtunkDir); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "linked %s\n", rtunkDir)
	return nil
}

// linkTools symlinks .rtunk/tools/<id> directly to each resolved tool's shim executable (e.g.
// .rtunk/tools/shellcheck), at the version this repo actually resolves -- not the flat
// shims/tools cache tree, which buries every tool under <id>/<version>/<platform>/.
func linkTools(rtunkDir, cacheDir string, cfg config.Config) error {
	toolsDir := filepath.Join(rtunkDir, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		return err
	}
	root, err := download.Root(cacheDir)
	if err != nil {
		return err
	}
	for id := range cfg.Tools {
		version, err := resolvedVersionFor(cfg, "tools", id)
		if err != nil {
			return err
		}
		target := download.ShimPath(root, "tools", id, version, id)
		if err := relink(filepath.Join(toolsDir, id), target); err != nil {
			return err
		}
	}
	return nil
}

// linkPluginSources symlinks .rtunk/plugins/<source-id> to each configured git plugin source's
// checkout. A `local:` source is already a real directory in the repo -- nothing to link.
func linkPluginSources(rtunkDir, cacheDir string, cfg config.Config) error {
	pluginsDir := filepath.Join(rtunkDir, "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		return err
	}
	pluginsRoot, err := config.PluginsCacheRoot(cacheDir)
	if err != nil {
		return err
	}
	for id, src := range cfg.Plugins.Sources {
		if src.Local != "" {
			continue
		}
		if err := relink(filepath.Join(pluginsDir, id), config.CheckoutDir(pluginsRoot, src)); err != nil {
			return err
		}
	}
	return nil
}

// relink replaces path with a fresh symlink to target. target need not exist yet -- e.g. no run
// has happened, or nothing has been downloaded -- the symlink is created dangling and resolves
// once something populates it, same as trunk's own tools/logs/plugins symlinks.
func relink(path, target string) error {
	if existing, err := os.Readlink(path); err == nil && existing == target {
		return nil
	}
	_ = os.Remove(path)
	return os.Symlink(target, path)
}

// writeLinkGitignore ensures .rtunk/.gitignore exists, ignoring everything link creates -- only
// rtunk.yaml (and this .gitignore itself) belong in version control.
func writeLinkGitignore(rtunkDir string) error {
	path := filepath.Join(rtunkDir, ".gitignore")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	//nolint:gosec // .rtunk/.gitignore is repo-tracked, readable like every other tracked path
	return os.WriteFile(path, []byte("logs\ntools\nplugins\n"), 0o644)
}
