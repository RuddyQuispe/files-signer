// Package signing is the domain: it knows how to produce and verify
// PKCS#7/CMS signatures over raw bytes. It does not know where the signing
// material came from, nor how the user invokes it.
package signing

import "fmt"

// OutputMode selects which signature files to produce.
type OutputMode int

const (
	// Both produces the attached and the detached signature.
	Both OutputMode = iota
	// AttachedOnly embeds the content in the signature (larger file, .p7m).
	AttachedOnly
	// DetachedOnly produces only the signature (small file, .p7s).
	DetachedOnly
)

// ParseOutputMode maps a CLI value to an OutputMode. Empty defaults to Both.
func ParseOutputMode(s string) (OutputMode, error) {
	switch s {
	case "", "both":
		return Both, nil
	case "attached":
		return AttachedOnly, nil
	case "detached":
		return DetachedOnly, nil
	default:
		return 0, fmt.Errorf("invalid output mode %q (use: both, attached, detached)", s)
	}
}

func (m OutputMode) wantsAttached() bool { return m == Both || m == AttachedOnly }
func (m OutputMode) wantsDetached() bool { return m == Both || m == DetachedOnly }

// SignatureFilenames returns the output paths for the given input, following the
// standard S/MIME naming: the attached (enveloping) signature uses .p7m and the
// detached signature uses .p7s, both keeping the full original name
// (documento.pdf.p7m / documento.pdf.p7s). A returned path is empty when the mode
// does not produce it. The two extensions differ, so the names never collide.
func SignatureFilenames(input string, mode OutputMode) (attached, detached string) {
	if mode.wantsAttached() {
		attached = input + ".p7m"
	}
	if mode.wantsDetached() {
		detached = input + ".p7s"
	}
	return attached, detached
}
