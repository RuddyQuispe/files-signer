// Package cli is the terminal interface. It parses commands and delegates all
// real work to the signing domain, so the domain stays independent of the UI.
package cli

import (
	"fmt"
	"os"
)

const usage = `files-sign — sign, verify and extract files with PKCS#7/CMS

Usage:
  files-sign sign    [flags] <file...>
  files-sign verify  [flags] <file>
  files-sign extract [flags] <file.p7m>

Run "files-sign sign -h", "files-sign verify -h" or "files-sign extract -h" for command flags.`

// Run dispatches a subcommand and returns the process exit code.
func Run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}

	switch args[0] {
	case "sign":
		return signCmd(args[1:])
	case "verify":
		return verifyCmd(args[1:])
	case "extract":
		return extractCmd(args[1:])
	case "-h", "--help", "help":
		fmt.Println(usage)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s\n", args[0], usage)
		return 2
	}
}
