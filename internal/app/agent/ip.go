package agent

import (
	"errors"
	"fmt"
	"net"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
)

var (
	// ErrUDPInvalidLocalAddr is returned when local address of newly created UDP "connection" is not [net.UDPAddr].
	ErrUDPInvalidLocalAddr = errors.New("local address of UDP connection is not net.UDPAddr")
)

// getOutgoingIPv4 attempts to determine outgoing IP address of the current machine for a given target,
// using UDP socket creation to force OS to make a routing decision and assign local IP address to the
// potential UDP "connection".
func getOutgoingIPv4(remoteAddr string) (net.IP, error) {
	conn, err := net.Dial("udp4", remoteAddr)
	if err != nil {
		return nil, fmt.Errorf("cannot dial udp4 to %s: %w", remoteAddr, err)
	}
	defer conn.Close()

	addr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok { // unlikely, but better return an error than get nil dereference later
		return nil, ErrUDPInvalidLocalAddr

	}

	return addr.IP, nil
}

func prepareRealIPHeader(remoteAddr string, signer hmacsigner.HMACSigner) (httpheaders.XRealIP, error) {
	var (
		realIP httpheaders.XRealIP
		err    error
	)

	realIP.IP, err = getOutgoingIPv4(remoteAddr)
	if err != nil {
		return realIP, fmt.Errorf("cannot detect outgoing IPv4: %w", err)
	}

	signature, err := signer.Sign(realIP.IP.To16()) // ensure we sign longest possible IP bytes since the server must do the same
	switch {
	case errors.Is(err, hmacsigner.ErrMissingSecretKey):
		// no secret key, proceed without signature
	case err == nil:
		realIP.Hash = signature
	default:
		return realIP, fmt.Errorf("cannot create HMAC signature for IP %v: %w", realIP.IP, err)
	}

	return realIP, nil
}
