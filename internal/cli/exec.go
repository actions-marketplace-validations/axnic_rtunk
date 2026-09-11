package cli

import (
	"io"
	"os"
	"os/exec"

	"github.com/xunleii/rtunk/pkg/trunk/download"
)

// execCmd is `rtunk exec`/`rtunk x`: downloads the shim first if it isn't already cached (via the
// same Download engine as `rtunk download`), then runs it with Args, inheriting stdio.
type execCmd struct {
	Category string   `arg:"" enum:"tools,runtimes,actions" help:"Resource category."`
	ID       string   `arg:"" help:"Resource id, optionally @version."`
	Args     []string `arg:"" optional:"" passthrough:"" help:"Arguments passed to the tool."`
}

// Run's stderr param is typed Stderr (not io.Writer) so Kong's by-type DI binds it to the
// separate Stderr binding registered in cli.go, rather than collapsing onto the stdout binding
// both params would otherwise share.
func (c *execCmd) Run(cli *CLI, stdout io.Writer, stderr Stderr) error {
	cfg, err := resolveConfig(cli.Config, cli.CacheDir, false)
	if err != nil {
		return err
	}
	id, version, pinned := cutVersion(c.ID)
	if !pinned {
		version, err = resolvedVersionFor(cfg, c.Category, id)
		if err != nil {
			return err
		}
	}

	root, err := download.Root(cli.CacheDir)
	if err != nil {
		return err
	}
	shimPath := download.ShimPath(root, c.Category, id, version, id)
	if _, statErr := os.Stat(shimPath); statErr != nil {
		events, err := download.Download(cfg, cli.CacheDir, download.Ref{Category: c.Category, ID: id, Version: version})
		if err != nil {
			return err
		}
		for ev := range events {
			if ev.Phase == download.Failed {
				return ev.Err
			}
		}
	}

	// Kong's passthrough positional keeps a literal "--" separator as Args[0] when the caller wrote
	// one (it only pops "--" for a non-passthrough next positional) -- strip it so the shim sees
	// the caller's arguments, not the separator that introduced them.
	args := c.Args
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}

	cmd := exec.Command(shimPath, args...)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = stdout, stderr, os.Stdin
	return cmd.Run()
}
