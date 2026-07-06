package keystore

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"go.step.sm/crypto/pemutil"
)

// PEMLoader loads signing material from PEM data. The certificate and the
// private key may live in the same file or in separate ones. The private key
// may be password-protected (both legacy and PKCS#8 encryption are supported).
type PEMLoader struct {
	// CertFile holds the certificate (and, if KeyFile is empty, the key too).
	CertFile string
	// KeyFile holds the private key. Defaults to CertFile when empty.
	KeyFile string
	// Password decrypts the private key. Empty means the key is unencrypted.
	Password string
}

// Load implements Loader.
func (l PEMLoader) Load() (*Material, error) {
	certs, err := certificatesFromFile(l.CertFile)
	if err != nil {
		return nil, err
	}
	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificate found in %q", l.CertFile)
	}

	keyFile := l.KeyFile
	if keyFile == "" {
		keyFile = l.CertFile
	}
	key, err := privateKeyFromFile(keyFile, l.Password)
	if err != nil {
		return nil, err
	}

	return &Material{
		Certificate: certs[0],
		PrivateKey:  key,
		Chain:       certs[1:],
	}, nil
}

// certificatesFromFile decodes every CERTIFICATE block in a PEM file, in order.
func certificatesFromFile(path string) ([]*x509.Certificate, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading certificate file: %w", err)
	}

	var certs []*x509.Certificate
	rest := raw
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parsing certificate: %w", err)
		}
		certs = append(certs, cert)
	}
	return certs, nil
}

// privateKeyFromFile extracts the PRIVATE KEY block from a PEM file and decrypts
// it with the password if needed. pemutil handles both legacy and PKCS#8
// encryption, which the standard library does not.
func privateKeyFromFile(path, password string) (crypto.PrivateKey, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading key file: %w", err)
	}

	rest := raw
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if !strings.Contains(block.Type, "PRIVATE KEY") {
			continue
		}

		var opts []pemutil.Options
		if password != "" {
			opts = append(opts, pemutil.WithPassword([]byte(password)))
		}
		key, err := pemutil.Parse(pem.EncodeToMemory(block), opts...)
		if err != nil {
			return nil, fmt.Errorf("parsing private key (wrong password?): %w", err)
		}
		return key, nil
	}
	return nil, fmt.Errorf("no private key found in %q", path)
}
