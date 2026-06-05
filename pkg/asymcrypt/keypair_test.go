package asymcrypt_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/pkg/asymcrypt"
)

func TestX25519KeyPair(t *testing.T) {
	kp, err := asymcrypt.NewX25519KeyPair()
	require.NoErrorf(t, err, "keypair generation error")

	private, err := asymcrypt.ParsePrivateKey(kp.Private)
	require.NoErrorf(t, err, "cannot parse generated private key")
	require.NotNil(t, private)

	public, err := asymcrypt.ParsePublicKey(kp.Public)
	require.NoErrorf(t, err, "cannot parse generated public key")
	require.NotNil(t, public)
}
