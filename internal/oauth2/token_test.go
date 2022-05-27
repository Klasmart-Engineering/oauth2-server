package oauth2

import "testing"

func TestROPCInvalid(t *testing.T) {
	// returns 400 on `grant_type=password`
}

func TestAuthorizationCodeInvalid(t *testing.T) {
	// returns 400 on `code=`
}

func TestClientCredentialsInvalidCredentials(t *testing.T) {
	// returns 400 if credentials are invalid
}

func TestClientCredentialsNonexistentClient(t *testing.T) {
	// returns 400 if credentials are invalid
}

func TestClientCredentialsValid(t *testing.T) {
	// can retrieve a valid access_token

	// access_token payload contains:
	// * aud
	// * exp
	// * iat
	// * iss
	// * jti
	// * scp
	// * sub

	// access_token header contains:
	// * alg
	// * kid
	// * typ

	// access_token can be validated using the JWKS endpoint
}
