package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

// findTrunkYAML walks up from the working directory looking for .trunk/trunk.yaml, the same way
// git locates .git -- the nearest match wins. rtunk's own .rtunk/rtunk.yaml is not read yet (see
// pkg/trunk/config's package doc).
func findTrunkYAML() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	start := dir
	for {
		candidate := filepath.Join(dir, ".trunk", "trunk.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .trunk/trunk.yaml found (searched from %s upward); use --config to specify one", start)
		}
		dir = parent
	}
}
