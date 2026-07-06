package main

import (
	"os"

	"files-sign/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
