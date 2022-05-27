package keys

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"gopkg.in/square/go-jose.v2"
)

const (
	PRIVATE_KEY_PATH = "internal/keys/private.pem"
	PUBLIC_KEY_PATH  = "internal/keys/public.pem"
	KID              = "2c7ef7a0-913f-458d-8c84-be44b3091cb3"
)

func LoadPrivateKey() (*rsa.PrivateKey, error) {
	bytes, err := loadRSAKeyFile(PRIVATE_KEY_PATH)
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

func JWKS() httprouter.Handle {
	// TODO: remove panics, replace with AWS KMS
	bytes, err := loadRSAKeyFile(PUBLIC_KEY_PATH)
	if err != nil {
		panic(err)
	}

	key, err := parseRSAPublicKey(bytes)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// TODO: consider caching/saving to disk
		set := jose.JSONWebKeySet{
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
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(set)
	}
}
