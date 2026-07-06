package signing

import (
	"crypto/x509"
	"fmt"

	"github.com/digitorus/pkcs7"
)

// Extract verifies an attached (.p7m) signature and returns the original content
// that is embedded inside it — the whole point of an attached signature: recover
// the intact original after proving it was not modified.
//
// trust behaves like in Verify: nil validates only the signature integrity and
// signer; non-nil also validates the certificate chain against it. Extract fails
// on detached (.p7s) signatures, which carry no content.
func Extract(signature []byte, trust *x509.CertPool) ([]byte, *VerifyResult, error) {
	p7, err := pkcs7.Parse(signature)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing signature: %w", err)
	}
	if len(p7.Content) == 0 {
		return nil, nil, fmt.Errorf("this is a detached signature (.p7s); it does not contain the original file")
	}

	if trust != nil {
		err = p7.VerifyWithChain(trust)
	} else {
		err = p7.Verify()
	}
	if err != nil {
		return nil, nil, fmt.Errorf("signature verification failed: %w", err)
	}

	res := &VerifyResult{
		Signer:       p7.GetOnlySigner(),
		TrustChecked: trust != nil,
	}
	return p7.Content, res, nil
}
