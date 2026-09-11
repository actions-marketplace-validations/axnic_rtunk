// Package cli implements rtunk's command-line interface on top of
// github.com/alecthomas/kong. See ROADMAP.md for the staged command set; this currently
// implements v0.1 ("read and query an existing trunk configuration") only.
package cli

import (
	"io"

	"github.com/alecthomas/kong"
)

// CLI is Kong's grammar root: the two flags every subcommand needs to locate and resolve a
// trunk.yaml, plus the config subcommand tree.
type CLI struct {
	Config   string `help:"Path to trunk.yaml (default: nearest .trunk/trunk.yaml)."`
	CacheDir string `help:"Plugin cache directory (default: OS cache dir)." env:"RTUNK_CACHE_DIR"`

	ConfigCmd   configCmd   `cmd:"" name:"config" help:"Query the resolved trunk configuration."`
	DownloadCmd downloadCmd `cmd:"" name:"download" help:"Download enabled tools/runtimes into the local cache."`
}

// Run parses args against CLI's grammar and executes the selected command's Run(), writing to
// stdout/stderr. It does not call os.Exit itself -- cmd/rtunk/main.go owns the process exit code
// -- except that Kong's own --help handling exits the process directly (kong.Exit's default).
func Run(args []string, stdout, stderr io.Writer) error {
	var cli CLI
	parser, err := kong.New(&cli,
		kong.Name("rtunk"),
		kong.Description("Query and act on trunk-style repo configuration."),
		kong.Writers(stdout, stderr),
		kong.BindFor[io.Writer](stdout),
	)
	if err != nil {
		return err
	}

	kctx, err := parser.Parse(args)
	if err != nil {
		return err
	}
	return kctx.Run()
}
