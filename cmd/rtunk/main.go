// Command rtunk is the CLI entry point. See ROADMAP.md for the staged command set; this file
// wires the process's argv/stdio/exit code to internal/cli, which implements the grammar itself.
package main

import (
	"fmt"
	"os"

	"github.com/xunleii/rtunk/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "rtunk:", err)
		os.Exit(1)
	}
}
