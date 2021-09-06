package jwt

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
)

type Key struct {
	// id (required) represents the key identifier e.g. the md5 representation of the public key
	id string
	// public (required) rsa key
	public *rsa.PublicKey
	// private (optional) rsa key
	private *rsa.PrivateKey
}

// NewKey return a new Key
func NewKey(id string, public *rsa.PublicKey, private *rsa.PrivateKey) Key {
	return Key{
		id:      id,
		public:  public,
		private: private,
	}
}

// NewKeyFromFilenames returns a new Key from the given file names
func NewKeyFromFilenames(publicKeyPemFilename, privateKeyPemFilename string) (Key, error) {
	var id string
	var public *rsa.PublicKey
	var private *rsa.PrivateKey

	// load private key
	if privateKeyPemFilename != "" {
		if bytes, err := ioutil.ReadFile(privateKeyPemFilename); err != nil {
			return Key{}, errors.Wrap(err, "failed to read private key: "+privateKeyPemFilename)
		} else if key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(strings.ReplaceAll(string(bytes), `\n`, "\n"))); err != nil {
			return Key{}, errors.Wrap(err, "failed to parse private key: "+privateKeyPemFilename)
		} else {
			private = key
		}
	}

	// load public key
	if bytes, err := ioutil.ReadFile(publicKeyPemFilename); err != nil {
		return Key{}, errors.Wrap(err, "failed to read public key: "+publicKeyPemFilename)
	} else if key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(strings.ReplaceAll(string(bytes), `\n`, "\n"))); err != nil {
		return Key{}, errors.Wrap(err, "failed to parse public key: "+publicKeyPemFilename)
	} else {
		hasher := sha256.New()
		hasher.Write(bytes)
		id = hex.EncodeToString(hasher.Sum(nil))
		public = key
	}

	return NewKey(id, public, private), nil
}

// NewDeprecatedKeysFromFilenames returns new Keys from the given file names
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

// NewKeysFromFilenames helper
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
