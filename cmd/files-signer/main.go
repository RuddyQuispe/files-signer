package main

import (
	"os"

	"files-signer/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
