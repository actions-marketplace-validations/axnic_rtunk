package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/axnic/rtunk/pkg/git"
)

// initScaffold is the exact content `rtunk init` writes to a fresh .rtunk/rtunk.yaml -- v1.11.0 is
// a real, currently-working trunk-io/plugins tag, the same one this repo's own real
// .trunk/trunk.yaml pins (not an arbitrary placeholder). Enabled lists are intentionally omitted
// rather than written empty (`enabled: []`) -- trunkFile's own struct fields already default to
// nil when absent, so there is no functional difference, and a shorter scaffold reads better as a
// starting point built on via the already-existing `rtunk linters enable`/`rtunk actions enable`/
// `rtunk git-hooks sync`.
const initScaffold = `version: "0.1"
plugins:
  sources:
    - id: trunk
      uri: https://github.com/trunk-io/plugins
      ref: v1.11.0
`

// initCmd is `rtunk init`: ROADMAP.md v0.7, scaffolding .rtunk/rtunk.yaml -- or migrating an
// existing .trunk/ setup into .rtunk/ -- then, on a terminal, picking linters and actions and
// downloading them before linking .rtunk/{logs,tools,plugins}.
type initCmd struct {
	Force bool `help:"Overwrite an existing .rtunk/rtunk.yaml."`
}

func (c *initCmd) Run(cli *CLI, stdout io.Writer, stderr Stderr) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	repoRoot, err := git.RepoRoot(cwd)
	if err != nil {
		return err
	}

	rtunkDir := filepath.Join(repoRoot, ".rtunk")
	configPath := filepath.Join(rtunkDir, "rtunk.yaml")

	if !c.Force {
		if _, statErr := os.Stat(configPath); statErr == nil {
			return fmt.Errorf("rtunk: %s already exists; use --force to overwrite", configPath)
		}
	}

	//nolint:gosec // .rtunk/ and its rtunk.yaml are repo-tracked, readable like every other tracked path
	if err := os.MkdirAll(rtunkDir, 0o755); err != nil {
		return err
	}
	if _, statErr := os.Stat(filepath.Join(repoRoot, ".trunk", "trunk.yaml")); statErr == nil {
		if err := migrateTrunk(repoRoot, stdout); err != nil {
			return err
		}
	} else {
		//nolint:gosec // see above
		if err := os.WriteFile(configPath, []byte(initScaffold), 0o644); err != nil {
			return err
		}
	}
	_, _ = fmt.Fprintf(stdout, "initialized rtunk at %s\n", configPath)

	sub := &CLI{Config: configPath, CacheDir: cli.CacheDir}
	if f, ok := stdout.(*os.File); ok && isTerminal(f) && stdinIsTerminal() {
		if _, err := interactiveLintersEnable(sub, stdout, stderr); err != nil {
			return err
		}
		if err := interactiveActionsEnable(sub); err != nil {
			return err
		}
		if err := (&downloadCmd{}).Run(sub, stdout, stderr); err != nil {
			_, _ = fmt.Fprintf(stderr, "warning: %v (check and fmt retry on their first run)\n", err)
		}
	} else {
		_, _ = fmt.Fprintln(stdout, "next: rtunk linters enable, rtunk actions enable, rtunk download")
	}

	if err := (&linkCmd{}).Run(sub, stdout); err != nil {
		_, _ = fmt.Fprintf(stderr, "warning: could not link .rtunk/{logs,tools,plugins}: %v\n", err)
	}
	_, _ = fmt.Fprintln(stdout, "then: rtunk git-hooks sync to install the git hooks of the enabled actions")
	return nil
}

// trunkMigrations are what migrateTrunk carries from .trunk/ to .rtunk/: the config (renamed),
// the linters' config files and the local, untracked user overrides.
var trunkMigrations = [][2]string{
	{"trunk.yaml", "rtunk.yaml"},
	{"configs", "configs"},
	{"user_trunk.yaml", "user_trunk.yaml"},
	{"user.yaml", "user.yaml"},
}

// migrateTrunk moves repoRoot's .trunk/ setup into .rtunk/ (content unchanged: rtunk reads
// trunk.yaml's format as is), replacing what --force lets init overwrite, then removes the rest
// of .trunk/ -- trunk's own links into ~/.cache/trunk, its plugin checkouts and its .gitignore,
// all regenerable. trunk.yaml is version-controlled, so git still has the original.
func migrateTrunk(repoRoot string, w io.Writer) error {
	trunkDir, rtunkDir := filepath.Join(repoRoot, ".trunk"), filepath.Join(repoRoot, ".rtunk")
	for _, m := range trunkMigrations {
		from, to := filepath.Join(trunkDir, m[0]), filepath.Join(rtunkDir, m[1])
		if _, err := os.Lstat(from); err != nil {
			continue
		}
		if err := os.RemoveAll(to); err != nil {
			return err
		}
		if err := os.Rename(from, to); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(w, "migrated .trunk/%s -> .rtunk/%s\n", m[0], m[1])
	}
	if err := os.RemoveAll(trunkDir); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(w, "removed .trunk/")
	return nil
}
