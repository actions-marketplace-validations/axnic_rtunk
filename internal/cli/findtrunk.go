package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// findTrunkYAML walks up from the working directory looking for .trunk/trunk.yaml, the same way
// git locates .git -- the nearest match wins. The walk is bounded by the git repository root (if
// any): trunk.yaml belongs to the repo you're in, so a miss must not fall through to an unrelated
// .trunk/trunk.yaml sitting further up the filesystem, e.g. in a parent repo or the home dir.
// Outside a git repo, it falls back to walking to the filesystem root. rtunk's own
// .rtunk/rtunk.yaml is not read yet (see pkg/trunk/config's package doc).
func findTrunkYAML() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	start := dir

	var gitRoot string
	if out, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output(); err == nil {
		gitRoot = strings.TrimSpace(string(out))
	}

	for {
		candidate := filepath.Join(dir, ".trunk", "trunk.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		if dir == gitRoot {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("no .trunk/trunk.yaml found (searched from %s upward); use --config to specify one", start)
}
