package signing

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"files-sign/internal/keystore"
)

// newTestMaterial builds an in-memory self-signed certificate + key so tests
// need no external fixtures.
func newTestMaterial(t *testing.T) (*keystore.Material, *x509.Certificate) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "files-sign test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("creating certificate: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parsing certificate: %v", err)
	}
	return &keystore.Material{Certificate: cert, PrivateKey: key}, cert
}

func TestAttachedRoundTrip(t *testing.T) {
	material, _ := newTestMaterial(t)
	signer := Signer{Material: material}
	content := []byte("any bytes: pretend this is a Dockerfile\n")

	res, err := signer.Sign(content, AttachedOnly)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if res.Attached == nil || res.Detached != nil {
		t.Fatalf("AttachedOnly should produce only the attached signature")
	}
	// Attached carries its own content: verify with nil content.
	if _, err := Verify(res.Attached, nil, nil); err != nil {
		t.Fatalf("verify attached: %v", err)
	}
}

func TestDetachedRoundTrip(t *testing.T) {
	material, _ := newTestMaterial(t)
	signer := Signer{Material: material}
	content := []byte("payload of a zip file")

	res, err := signer.Sign(content, DetachedOnly)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if res.Detached == nil || res.Attached != nil {
		t.Fatalf("DetachedOnly should produce only the detached signature")
	}
	if _, err := Verify(res.Detached, content, nil); err != nil {
		t.Fatalf("verify detached: %v", err)
	}
}

func TestDetachedSmallerThanAttached(t *testing.T) {
	material, _ := newTestMaterial(t)
	signer := Signer{Material: material}
	content := make([]byte, 50_000) // large payload

	res, err := signer.Sign(content, Both)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if len(res.Detached) >= len(res.Attached) {
		t.Fatalf("detached (%d) should be much smaller than attached (%d)",
			len(res.Detached), len(res.Attached))
	}
}

func TestTamperedContentFails(t *testing.T) {
	material, _ := newTestMaterial(t)
	signer := Signer{Material: material}

	res, err := signer.Sign([]byte("original"), DetachedOnly)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if _, err := Verify(res.Detached, []byte("tampered"), nil); err == nil {
		t.Fatalf("verification should fail for tampered content")
	}
}

func TestExtractRecoversOriginal(t *testing.T) {
	material, _ := newTestMaterial(t)
	signer := Signer{Material: material}
	original := []byte("recuperame intacto: bytes de cualquier archivo")

	res, err := signer.Sign(original, AttachedOnly)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	got, vr, err := Extract(res.Attached, nil)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if !bytes.Equal(got, original) {
		t.Fatalf("extracted content differs from original")
	}
	if vr.Signer == nil {
		t.Fatalf("expected signer info")
	}
}

func TestExtractDetachedFails(t *testing.T) {
	material, _ := newTestMaterial(t)
	signer := Signer{Material: material}

	res, err := signer.Sign([]byte("payload"), DetachedOnly)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if _, _, err := Extract(res.Detached, nil); err == nil {
		t.Fatalf("Extract should fail on a detached signature")
	}
}

func TestSignatureFilenames(t *testing.T) {
	cases := []struct {
		name             string
		input            string
		mode             OutputMode
		wantAtt, wantDet string
	}{
		{"both keeps full name", "documento.pdf", Both, "documento.pdf.p7m", "documento.pdf.p7s"},
		{"attached only", "app.yaml", AttachedOnly, "app.yaml.p7m", ""},
		{"detached only", "app.yaml", DetachedOnly, "", "app.yaml.p7s"},
		{"extension-less input", "Dockerfile", Both, "Dockerfile.p7m", "Dockerfile.p7s"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			att, det := SignatureFilenames(c.input, c.mode)
			if att != c.wantAtt || det != c.wantDet {
				t.Fatalf("SignatureFilenames(%q,%v) = (%q,%q), want (%q,%q)",
					c.input, c.mode, att, det, c.wantAtt, c.wantDet)
			}
		})
	}
}

func TestTrustValidation(t *testing.T) {
	material, cert := newTestMaterial(t)
	signer := Signer{Material: material}

	res, err := signer.Sign([]byte("trust me"), DetachedOnly)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	// Trusted: the signer certificate is in the pool.
	trusted := x509.NewCertPool()
	trusted.AddCert(cert)
	got, err := Verify(res.Detached, []byte("trust me"), trusted)
	if err != nil {
		t.Fatalf("verify with trust: %v", err)
	}
	if !got.TrustChecked {
		t.Fatalf("TrustChecked should be true")
	}

	// Untrusted: empty pool must fail the chain check.
	empty := x509.NewCertPool()
	if _, err := Verify(res.Detached, []byte("trust me"), empty); err == nil {
		t.Fatalf("verification should fail against an empty trust pool")
	}
}
