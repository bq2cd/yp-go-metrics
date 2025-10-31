package servertest

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
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

type ListenAddress struct {
	Host string
	Port uint32
}

func (la ListenAddress) String() string {
	return fmt.Sprintf("%s:%d", la.Host, la.Port)
}

// NewListenAddress parses a listen address in the form of "host:port"
// and returns a [ListenAddress] struct.
func NewListenAddress(t *testing.T, addr string) ListenAddress {
	parts := strings.SplitN(addr, ":", 2)
	port, err := strconv.ParseUint(parts[1], 10, 32)
	require.NoError(t, err)
	return ListenAddress{Host: parts[0], Port: uint32(port)}
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

// NewTempFileFactory tracks created temporary files to facilitate their easy removal with a single call to [RemoveAll],
// typically using `defer` statement.
func NewTempFileFactory(t *testing.T) *tempFileFactory {
	return &tempFileFactory{
		t:       t,
		created: make([]string, 0),
	}
}

type tempFileFactory struct {
	t       *testing.T
	created []string
}

func (ff *tempFileFactory) create(dir string, pattern string) string {
	f, err := os.CreateTemp(dir, pattern)
	require.NoError(ff.t, err)
	require.NoError(ff.t, f.Close())
	return f.Name()
}

// Create creates a temporary file in default directory for temporary files,
// closes the file, records its location internally and
// returns path to the file.
func (ff *tempFileFactory) Create(pattern string) string {
	path := ff.create(os.TempDir(), pattern)
	ff.created = append(ff.created, path)
	return path
}

// RemoveAll attempts to remove all temporary files that were created
// with [Create] method.
func (ff *tempFileFactory) RemoveAll() {
	for _, path := range ff.created {
		_ = os.Remove(path)
	}
}

// NewListenAddressFactory tracks generated random addresses suitable for listening on (e.g. by a HTTP server).
// It uses [GetRandomListenAddress] function for generation.
func NewListenAddressFactory(t *testing.T) *listenAddressFactory {
	return &listenAddressFactory{
		t:         t,
		allocated: make([]string, 0),
	}
}

type listenAddressFactory struct {
	t         *testing.T
	allocated []string
}

// New will generate a new random address with [GetRandomListenAddress] function and store it for future references.
func (f *listenAddressFactory) New() string {
	addr := GetRandomListenAddress(f.t)
	f.allocated = append(f.allocated, addr)
	return addr
}

// Get will attempt to return already generated address by its index,
// but will resort to generating a new one if such address does not yet exist.
// Under the hood, it will grow a slice with stored addresses to accommodate requested indexes, producing a "sparse" slice if
// indexes are not requested in proper order.
func (f *listenAddressFactory) Get(idx int) string {
	n := idx + 1
	if cap(f.allocated) < n {
		f.allocated = slices.Grow(f.allocated, n-cap(f.allocated))
	}
	f.allocated = f.allocated[:max(len(f.allocated), n)]
	if f.allocated[idx] == "" {
		f.allocated[idx] = GetRandomListenAddress(f.t)
	}
	return f.allocated[idx]
}

// GetCwd returns current working directory for the process.
// It uses [os.Getwd] under the hood, but will fail on an error.
func GetCwd(t *testing.T) string {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	return cwd
}
