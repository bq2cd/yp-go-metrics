package servertest

import (
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// GetRandomListenAddress returns the first free listen address
// bound to 'localhost' and using TCP protocol.
func GetRandomListenAddress(t *testing.T) string {
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	addr := l.Addr().String()

	err = l.Close()
	require.NoError(t, err)

	return addr
}

// MakeSimpleRequest sends prepared request using provided HTTP client
// and ignores response completely, so it only returns if any network error
// was encountered in the process.
func MakeRequestDiscardResponse(c *http.Client, r *http.Request) error {
	if c == nil {
		c = http.DefaultClient
	}
	resp, err := c.Do(r)
	if err != nil {
		return err
	}
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}
