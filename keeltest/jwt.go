package keeltest

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	keeljwt "github.com/foomo/keel/jwt"
	"github.com/stretchr/testify/require"
)

// NewJWT returns a [keeljwt.JWT] signing with a freshly generated RSA key
// pair. It fails the test on error.
func NewJWT(t *testing.T) *keeljwt.JWT {
	t.Helper()

	publicPem, privatePem := testingx.GenerateRSAKeyPair(t)

	jwtKey, _, err := keeljwt.NewKeysFromFilenames(publicPem, privatePem, nil)
	require.NoError(t, err)

	return keeljwt.New(jwtKey)
}
