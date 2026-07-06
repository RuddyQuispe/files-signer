// Package keystore loads the signing material (certificate + private key) from
// a source. Today it reads PEM files; new sources (PFX/P12, OS keychain) can be
// added by implementing Loader without touching the signing domain.
package keystore

import (
	"crypto"
	"crypto/x509"
)

// Material is everything needed to produce a signature: the signer certificate,
// its private key, and any intermediate certificates that complete the chain.
type Material struct {
	Certificate *x509.Certificate
	PrivateKey  crypto.PrivateKey
	Chain       []*x509.Certificate
}

// Loader loads signing material from some source. PEM is the only implementation
// in v1; this interface is the extension point for future formats.
type Loader interface {
	Load() (*Material, error)
}
