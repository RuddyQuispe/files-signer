package cli

import (
	"crypto/x509"
	"flag"
	"fmt"
	"os"
	"strings"

	"files-sign/internal/signing"
)

func extractCmd(args []string) int {
	fs := flag.NewFlagSet("extract", flag.ContinueOnError)
	out := fs.String("o", "", "output file (default: the signature name without .p7m)")
	force := fs.Bool("f", false, "overwrite the output file if it already exists")
	trust := fs.Bool("trust", false, "validate the certificate chain against --ca")
	caFile := fs.String("ca", "", "PEM file with trusted CA certificate(s); used with --trust")
	if err := fs.Parse(reorder(fs, args)); err != nil {
		return 2
	}

	rest := fs.Args()
	if len(rest) != 1 {
		fmt.Fprintln(os.Stderr, "error: expected exactly one attached signature (.p7m) file")
		return 2
	}
	sigPath := rest[0]

	outPath := *out
	if outPath == "" {
		if !strings.HasSuffix(sigPath, ".p7m") {
			fmt.Fprintln(os.Stderr, "error: cannot infer output name; pass -o <file>")
			return 2
		}
		outPath = strings.TrimSuffix(sigPath, ".p7m")
	}

	if !*force {
		if _, err := os.Stat(outPath); err == nil {
			fmt.Fprintf(os.Stderr, "error: %q already exists (use -f to overwrite)\n", outPath)
			return 1
		}
	}

	signature, err := os.ReadFile(sigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading signature: %v\n", err)
		return 1
	}

	var pool *x509.CertPool
	if *trust {
		pool, err = loadTrustPool(*caFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
	}

	content, res, err := signing.Extract(signature, pool)
	if err != nil {
		fmt.Fprintf(os.Stderr, "INVALID: %v\n", err)
		return 1
	}
	if err := os.WriteFile(outPath, content, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
		return 1
	}

	fmt.Printf("extracted original → %s\n", outPath)
	if res.Signer != nil {
		fmt.Printf("  signer: %s\n", res.Signer.Subject)
	}
	return 0
}
