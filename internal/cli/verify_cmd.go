package cli

import (
	"crypto/x509"
	"flag"
	"fmt"
	"os"

	"files-sign/internal/signing"
)

func verifyCmd(args []string) int {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	sigFile := fs.String("sig", "", "detached signature file (.p7s); required to verify a detached signature")
	trust := fs.Bool("trust", false, "also validate the certificate chain against --ca")
	caFile := fs.String("ca", "", "PEM file with trusted CA certificate(s); used with --trust")
	if err := fs.Parse(reorder(fs, args)); err != nil {
		return 2
	}

	rest := fs.Args()
	if len(rest) != 1 {
		fmt.Fprintln(os.Stderr, "error: expected exactly one file to verify")
		return 2
	}
	target := rest[0]

	signature, content, err := loadSignatureAndContent(target, *sigFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
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

	res, err := signing.Verify(signature, content, pool)
	if err != nil {
		fmt.Fprintf(os.Stderr, "INVALID: %v\n", err)
		return 1
	}

	fmt.Println("VALID: signature is intact")
	if res.Signer != nil {
		fmt.Printf("  signer: %s\n", res.Signer.Subject)
	}
	if res.TrustChecked {
		fmt.Println("  trust:  certificate chain validated")
	} else {
		fmt.Println("  trust:  not checked (use --trust --ca ca.pem)")
	}
	return 0
}

// loadSignatureAndContent resolves the two verification inputs. When --sig is
// given, target is the original file and sigFile the detached signature. When
// --sig is empty, target is an attached signature that carries its own content.
func loadSignatureAndContent(target, sigFile string) (signature, content []byte, err error) {
	if sigFile != "" {
		content, err = os.ReadFile(target)
		if err != nil {
			return nil, nil, fmt.Errorf("reading original file: %w", err)
		}
		signature, err = os.ReadFile(sigFile)
		if err != nil {
			return nil, nil, fmt.Errorf("reading signature file: %w", err)
		}
		return signature, content, nil
	}

	signature, err = os.ReadFile(target)
	if err != nil {
		return nil, nil, fmt.Errorf("reading signature file: %w", err)
	}
	return signature, nil, nil
}

func loadTrustPool(caFile string) (*x509.CertPool, error) {
	if caFile == "" {
		return nil, fmt.Errorf("--trust requires --ca with trusted certificate(s)")
	}
	pem, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("reading CA file: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("no valid certificates found in %q", caFile)
	}
	return pool, nil
}
