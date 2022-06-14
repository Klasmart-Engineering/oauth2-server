package oauth2

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/KL-Engineering/oauth2-server/internal/crypto"
	"github.com/ory/fosite/compose"
	"github.com/pkg/errors"

	// "github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/storage"
	"github.com/ory/fosite/token/jwt"
)

const (
	HMAC_SECRET_PATH = "internal/crypto/hmac_secret"
)

// fosite requires four parameters for the server to get up and running:
// 1. config - for any enforcement you may desire, you can do this using `compose.Config`. You like PKCE, enforce it!
// 2. store - no auth service is generally useful unless it can remember clients and users.
//    fosite is incredibly composable, and the store parameter enables you to build and BYODb (Bring Your Own Database)
// 3. secret - required for code, access and refresh token generation.
// 4. privateKey - required for id/jwt token generation.
var (
	// Check the api documentation of `compose.Config` for further configuration options.
	config = &compose.Config{
		AccessTokenLifespan: time.Minute * 15,
		// ...
	}

	// This is the example storage that contains:
	// * an OAuth2 Client with id "my-client" and secrets "foobar" and "foobaz" capable of all oauth2 and open id connect grant and response types.
	// * a User for the resource owner password credentials grant type with username "peter" and password "secret".
	//
	// You will most likely replace this with your own logic once you set up a real world application.
	store = storage.NewExampleStore()

	// TODO remove hardcode
	secret = Must(func() (interface{}, error) {
		return LoadHMACSecret()
	}).([]byte)

	// privateKey is used to sign JWT tokens. The default strategy uses RS256 (RSA Signature with SHA-256)
	// TODO remove hardcode
	privateKey = Must(func() (interface{}, error) {
		return crypto.LoadPrivateKey()
	}).(*rsa.PrivateKey)
)

func LoadHMACSecret() ([]byte, error) {
	bytes, err := ioutil.ReadFile(HMAC_SECRET_PATH)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to read file at path: %s", HMAC_SECRET_PATH))
	}

	return bytes, nil
}

func Must(fn func() (interface{}, error)) interface{} {
	v, err := fn()
	if err != nil {
		log.Fatalln(err)
	}
	return v
}

// TODO remove global var
var oauth2Provider = compose.Compose(
	config,
	store,
	&compose.CommonStrategy{
		CoreStrategy: compose.NewOAuth2JWTStrategy(
			privateKey,
			compose.NewOAuth2HMACStrategy(config, secret, nil),
		),
		OpenIDConnectTokenStrategy: compose.NewOpenIDConnectStrategy(config, privateKey),
		JWTStrategy: &jwt.RS256JWTStrategy{
			PrivateKey: privateKey,
		},
	},
	nil,

	// OAuth2AuthorizeExplicitFactory,
	// OAuth2AuthorizeImplicitFactory,
	compose.OAuth2ClientCredentialsGrantFactory,
	compose.OAuth2RefreshTokenGrantFactory,
	// OAuth2ResourceOwnerPasswordCredentialsFactory,
	// RFC7523AssertionGrantFactory,

	// OpenIDConnectExplicitFactory,
	// OpenIDConnectImplicitFactory,
	// OpenIDConnectHybridFactory,
	// OpenIDConnectRefreshFactory,

	// OAuth2TokenIntrospectionFactory,
	// OAuth2TokenRevocationFactory,

	// OAuth2PKCEFactory,
)
