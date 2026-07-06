package signing

import (
	"crypto/x509"
	"fmt"

	"github.com/digitorus/pkcs7"
)

// VerifyResult reports who signed and whether trust was checked.
type VerifyResult struct {
	// Signer is the certificate that produced the signature.
	Signer *x509.Certificate
	// TrustChecked is true when the certificate chain was validated against a
	// trust store (the optional --trust switch).
	TrustChecked bool
}

// Verify checks a signature's integrity and identifies the signer.
//
//   - signature: the .p7b (attached) or .p7s (detached) bytes.
//   - content: the original file bytes; required for detached signatures, pass
//     nil for attached ones (the content travels inside the signature).
//   - trust: when non-nil, the certificate chain is validated against it; when
//     nil, only signature integrity and signer identity are checked.
func Verify(signature, content []byte, trust *x509.CertPool) (*VerifyResult, error) {
	p7, err := pkcs7.Parse(signature)
	if err != nil {
		return nil, fmt.Errorf("parsing signature: %w", err)
	}

	if len(content) > 0 {
		p7.Content = content // reattach the original for detached signatures
	}

	if trust != nil {
		err = p7.VerifyWithChain(trust)
	} else {
		err = p7.Verify()
	}
	if err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	return &VerifyResult{
		Signer:       p7.GetOnlySigner(),
		TrustChecked: trust != nil,
	}, nil
}
