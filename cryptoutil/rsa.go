package cryptoutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"errors"
)

// ParseASN1RSAPublicKey parses a DER encoded RSA public key
func ParseASN1RSAPublicKey(data []byte) (*rsa.PublicKey, error) {
	key, err := x509.ParsePKIXPublicKey(data)

	if err != nil {
		return nil, err
	}

	pubKey, ok := key.(*rsa.PublicKey)

	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return pubKey, nil
}

// MustParseASN1RSAPublicKey is like ParseASN1RSAPublicKey but panics instead of returning an error.
func MustParseASN1RSAPublicKey(data []byte) *rsa.PublicKey {
	key, err := ParseASN1RSAPublicKey(data)

	if err != nil {
		panic(err)
	}

	return key
}

// Encrypts a message with the given public key using RSA-OAEP and the SHA1 hash function.
func RSAEncrypt(pubkey *rsa.PublicKey, msg []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha1.New(), rand.Reader, pubkey, msg, nil)
}
