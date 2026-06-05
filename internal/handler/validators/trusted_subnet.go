package validators

import (
	"errors"
	"fmt"
	"net"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
)

// TrustedSubnet is a validator that verifies whether given IP address belongs to the trusted subnet and
// whether HMAC signature of the IP address (if provided) is correct.
type TrustedSubnet struct {
	subnet net.IPNet
	signer hmacsigner.HMACSigner
}

// NewTrustedSubnet creates an instance of [TrustedSubnet] validator.
func NewTrustedSubnet(subnet net.IPNet, signer hmacsigner.HMACSigner) *TrustedSubnet {
	return &TrustedSubnet{
		subnet: subnet,
		signer: signer,
	}
}

// IsXRealIPTrusted determines whether provided IP address belongs to the trusted subnet and whether its
// HMAC signature (if provided) is valid.
//
// It will return `true` without error only in the following cases:
// (1) Trusted subnet is not configured, therefore all IP addresses are considered trusted.
// (2) There's no HMAC secret key configured, and IP address belongs to the trusted subnet.
// (3) HMAC secret key is configured, HMAC signature of the IP address is valid, and IP address belongs
// to the trusted subnet.
//
// In all other cases, [IsXRealIPTrusted] will return `false`, optionally accompanied by an error.
// The error will only be returned if [hmacsigner.HMACSigner] returns an unexpected error; on signature
// mismatch `false` will be returned without error.
func (m *TrustedSubnet) IsXRealIPTrusted(realIP httpheaders.XRealIP) (bool, error) {
	if !m.isTrustedSubnetConfigured() {
		return true, nil
	}

	// reject requests without IP address
	if realIP.Empty() {
		return false, nil
	}

	ok, err := m.verifyXRealIPHash(realIP)
	if err != nil {
		return false, err
	}

	// reject requests with invalid hash
	if !ok {
		return false, nil
	}

	// reject requests with IP not in trusted subnet
	if !m.subnet.Contains(realIP.IP) {
		return false, nil
	}

	return true, nil
}

// IsTrustedSubnetConfigured returns `true` if validator was configured with non-empty trusted subnet.
func (m *TrustedSubnet) isTrustedSubnetConfigured() bool {
	return len(m.subnet.IP) > 0 && len(m.subnet.Mask) > 0
}

// verifyXRealIPHash performs verification of HMAC signature of the IP address.
// If HMAC secret key is not configured, it will return `true` and no error; otherwise, it will return
// `true` only if HMAC signature is valid.
// On signature mismatch, `false` without error will be returned.
// An error will be returned only when [hmacsigner.HMACSinger] returns an error (not on signature mismatch).
func (m *TrustedSubnet) verifyXRealIPHash(realIP httpheaders.XRealIP) (bool, error) {
	if !m.signer.HasKey() {
		return true, nil // ignore hash if no secret key is configured; assume IP is valid
	}

	err := m.signer.Verify(realIP.IP.To16(), realIP.Hash) // ensure we verify longest possible IP bytes; the sender must sign the same length
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, hmacsigner.ErrSignatureMismatch):
		return false, nil
	default:
		return false, fmt.Errorf("cannot verify X-Real-IP hash: %w", err)
	}
}
