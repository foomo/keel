package keeltest

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	keeljwt "github.com/foomo/keel/jwt"
	"github.com/stretchr/testify/require"
)

// NewJWT returns a new JWT instance with test keys for testing purposes.
func NewJWT(t *testing.T) *keeljwt.JWT {
	t.Helper()

	publicPem, privatePem := testingx.GenerateRSAKeyPair(t)

	jwtKey, _, err := keeljwt.NewKeysFromFilenames(publicPem, privatePem, nil)
	require.NoError(t, err)

	return keeljwt.New(jwtKey)
}
