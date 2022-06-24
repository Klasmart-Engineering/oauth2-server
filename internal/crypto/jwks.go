package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/square/go-jose.v2"
)

const (
	privateKeyPath = "internal/crypto/private.pem"
	publicKeyPath  = "internal/crypto/public.pem"
	KID              = "2c7ef7a0-913f-458d-8c84-be44b3091cb3"
)

func LoadPrivateKey() (*rsa.PrivateKey, error) {
	bytes, err := loadRSAKeyFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	privateKey, err := parseRSAPrivateKey(bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func loadRSAKeyFile(path string) ([]byte, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to read file at path: %s", path))
	}

	return bytes, nil
}

func parseRSAPublicKey(bytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(bytes)

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key.(*rsa.PublicKey), nil
}

func parseRSAPrivateKey(bytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(bytes)

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func JWKS() (*jose.JSONWebKeySet, error) {
	// TODO: replace with AWS KMS
	bytes, err := loadRSAKeyFile(publicKeyPath)
	if err != nil {
		return &jose.JSONWebKeySet{}, err
	}

	key, err := parseRSAPublicKey(bytes)
	if err != nil {
		return &jose.JSONWebKeySet{}, err
	}

	return &jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{
			{
				Algorithm: "RS256",
				Use:       "sig",
				Key:       key,
				// TODO: remove hardcoded kid - either store kid & some unique identifier for public key
				// or use a one-way deterministic approach to derive kid from key
				KeyID: KID,
			},
		},
	}, nil
}
