package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"files-signer/internal/keystore"
	"files-signer/internal/signing"
)

func signCmd(args []string) int {
	fs := flag.NewFlagSet("sign", flag.ContinueOnError)
	pemFile := fs.String("pem", "", "PEM file with the certificate (and key, if --key is not set)")
	keyFile := fs.String("key", "", "PEM file with the private key (default: --pem file)")
	password := fs.String("password", "", "password for the private key")
	passStdin := fs.Bool("password-stdin", false, "read the key password from stdin")
	out := fs.String("out", "both", "which files to produce: both | attached | detached")
	outdir := fs.String("outdir", "", "directory for the output files (default: next to each input)")
	if err := fs.Parse(reorder(fs, args)); err != nil {
		return 2
	}

	if *pemFile == "" {
		fmt.Fprintln(os.Stderr, "error: --pem is required")
		return 2
	}
	files := fs.Args()
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "error: no input files")
		return 2
	}

	mode, err := signing.ParseOutputMode(*out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	pass := *password
	if *passStdin {
		pass = readPasswordStdin()
	}

	material, err := keystore.PEMLoader{CertFile: *pemFile, KeyFile: *keyFile, Password: pass}.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading signing material: %v\n", err)
		return 1
	}
	signer := signing.Signer{Material: material}

	failed := false
	for _, file := range files {
		if err := signOne(signer, file, mode, *outdir); err != nil {
			fmt.Fprintf(os.Stderr, "error signing %q: %v\n", file, err)
			failed = true
			continue
		}
	}
	if failed {
		return 1
	}
	return 0
}

func signOne(signer signing.Signer, file string, mode signing.OutputMode, outdir string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	res, err := signer.Sign(content, mode)
	if err != nil {
		return err
	}

	base := file
	if outdir != "" {
		base = filepath.Join(outdir, filepath.Base(file))
	}

	attachedPath, detachedPath := signing.SignatureFilenames(base, mode)

	if res.Attached != nil {
		if err := writeFile(attachedPath, res.Attached); err != nil {
			return err
		}
		fmt.Printf("signed (attached) → %s\n", attachedPath)
	}
	if res.Detached != nil {
		if err := writeFile(detachedPath, res.Detached); err != nil {
			return err
		}
		fmt.Printf("signed (detached) → %s\n", detachedPath)
	}
	return nil
}

func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}

func readPasswordStdin() string {
	var b strings.Builder
	buf := make([]byte, 512)
	for {
		n, err := os.Stdin.Read(buf)
		b.Write(buf[:n])
		if err != nil {
			break
		}
	}
	return strings.TrimRight(b.String(), "\r\n")
}
