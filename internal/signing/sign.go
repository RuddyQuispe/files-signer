package signing

import (
	"fmt"

	"files-signer/internal/keystore"

	"github.com/digitorus/pkcs7"
)

// Result holds the produced signatures. A field is nil when the OutputMode did
// not request it.
type Result struct {
	Attached []byte // signature with the content embedded (.p7b)
	Detached []byte // signature only (.p7s)
}

// Signer produces PKCS#7/CMS signatures using the given material.
type Signer struct {
	Material *keystore.Material
}

// Sign signs content and returns the requested signature forms. Signing works on
// raw bytes, so it is independent of the file type (PDF, YAML, JAR, ZIP, ...).
func (s Signer) Sign(content []byte, mode OutputMode) (Result, error) {
	var res Result

	if mode.wantsAttached() {
		attached, err := s.build(content, false)
		if err != nil {
			return Result{}, fmt.Errorf("attached signature: %w", err)
		}
		res.Attached = attached
	}

	if mode.wantsDetached() {
		detached, err := s.build(content, true)
		if err != nil {
			return Result{}, fmt.Errorf("detached signature: %w", err)
		}
		res.Detached = detached
	}

	return res, nil
}

// build creates one CMS structure. When detach is true the content is stripped,
// yielding a small detached signature.
func (s Signer) build(content []byte, detach bool) ([]byte, error) {
	sd, err := pkcs7.NewSignedData(content)
	if err != nil {
		return nil, err
	}
	sd.SetDigestAlgorithm(pkcs7.OIDDigestAlgorithmSHA256)

	m := s.Material
	if len(m.Chain) > 0 {
		err = sd.AddSignerChain(m.Certificate, m.PrivateKey, m.Chain, pkcs7.SignerInfoConfig{})
	} else {
		err = sd.AddSigner(m.Certificate, m.PrivateKey, pkcs7.SignerInfoConfig{})
	}
	if err != nil {
		return nil, fmt.Errorf("adding signer: %w", err)
	}

	if detach {
		sd.Detach()
	}
	return sd.Finish()
}
