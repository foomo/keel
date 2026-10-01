package jwt

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

// Key is an RSA key pair identified by ID.
type Key struct {
	// ID (required) is the key identifier used as the token's "kid" header,
	// e.g. a hash of the public key.
	ID string
	// Public (required) is the RSA public key used for verification.
	Public *rsa.PublicKey
	// Private (optional) is the RSA private key used for signing.
	Private *rsa.PrivateKey
}

// NewKey returns a [Key] with the given ID and keys.
func NewKey(id string, public *rsa.PublicKey, private *rsa.PrivateKey) Key {
	return Key{
		ID:      id,
		Public:  public,
		Private: private,
	}
}

// NewKeyFromFilenames loads a [Key] from PEM files. The private key is
// optional and skipped if privateKeyPemFilename is empty. Literal \n escape
// sequences in the file contents are converted to newlines before parsing.
// The key ID is the hex-encoded SHA-256 of the trimmed public key file
// contents. It returns an error if a file cannot be read or parsed.
func NewKeyFromFilenames(publicKeyPemFilename, privateKeyPemFilename string) (Key, error) {
	var (
		id      string
		public  *rsa.PublicKey
		private *rsa.PrivateKey
	)

	// load private key

	if privateKeyPemFilename != "" {
		if value, err := os.ReadFile(privateKeyPemFilename); err != nil {
			return Key{}, errors.Wrap(err, "failed to read private key: "+privateKeyPemFilename)
		} else if key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(strings.ReplaceAll(string(value), `\n`, "\n"))); err != nil {
			return Key{}, errors.Wrap(err, "failed to parse private key: "+privateKeyPemFilename)
		} else {
			private = key
		}
	}

	// load public key
	if v, err := os.ReadFile(publicKeyPemFilename); err != nil {
		return Key{}, errors.Wrap(err, "failed to read public key: "+publicKeyPemFilename)
	} else if key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(strings.ReplaceAll(string(v), `\n`, "\n"))); err != nil {
		return Key{}, errors.Wrap(err, "failed to parse public key: "+publicKeyPemFilename)
	} else {
		hasher := sha256.New()
		_, _ = hasher.Write(bytes.TrimSpace(v))
		id = hex.EncodeToString(hasher.Sum(nil))
		public = key
	}

	return NewKey(id, public, private), nil
}

// NewDeprecatedKeysFromFilenames loads public-only keys from the given PEM
// files, for use with [WithDeprecatedKeys]. It returns the first load error.
func NewDeprecatedKeysFromFilenames(publicKeyPemFilenames []string) ([]Key, error) {
	deprecatedKeys := make([]Key, 0, len(publicKeyPemFilenames))
	for _, publicKeyPemFilename := range publicKeyPemFilenames {
		if value, err := NewKeyFromFilenames(publicKeyPemFilename, ""); err != nil {
			return nil, err
		} else {
			deprecatedKeys = append(deprecatedKeys, value)
		}
	}

	return deprecatedKeys, nil
}

// NewKeysFromFilenames loads the current key via [NewKeyFromFilenames] and
// the deprecated keys via [NewDeprecatedKeysFromFilenames].
func NewKeysFromFilenames(publicKeyPemFilename, privateKeyPemFilename string, deprecatedPublicKeyPemFilenames []string) (Key, []Key, error) {
	key, err := NewKeyFromFilenames(publicKeyPemFilename, privateKeyPemFilename)
	if err != nil {
		return Key{}, nil, err
	}

	deprecatedKeys, err := NewDeprecatedKeysFromFilenames(deprecatedPublicKeyPemFilenames)
	if err != nil {
		return Key{}, nil, err
	}

	return key, deprecatedKeys, nil
}
