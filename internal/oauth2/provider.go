package oauth2

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/KL-Engineering/oauth2-server/internal/crypto"
	"github.com/alexedwards/argon2id"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/pkg/errors"

	"github.com/ory/fosite/token/jwt"
)

const (
	HMAC_SECRET_PATH = "internal/crypto/hmac_secret"
)

var (
	config = &compose.Config{
		AccessTokenLifespan: time.Minute * 15,
		// ...
	}
)

type Hasher struct {
}

func (h *Hasher) Hash(ctx context.Context, data []byte) ([]byte, error) {
	s, err := argon2id.CreateHash(string(data), argon2id.DefaultParams)
	if err != nil {
		return nil, fmt.Errorf("ERROR: fosite.Hasher.Hash: %w", err)
	}
	return []byte(s), nil
}

func (h *Hasher) Compare(ctx context.Context, hash, data []byte) error {
	ok, err := argon2id.ComparePasswordAndHash(string(data), string(hash))
	if err != nil {
		return fmt.Errorf("ERROR: fosite.Hasher.Compare: %w", err)
	}
	if !ok {
		return errors.New("ERROR: hash does not match")
	}
	return nil
}

func LoadHMACSecret() ([]byte, error) {
	bytes, err := ioutil.ReadFile(HMAC_SECRET_PATH)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to read file at path: %s", HMAC_SECRET_PATH))
	}

	return bytes, nil
}

func NewProvider(db *dynamodb.Client) (fosite.OAuth2Provider, error) {
	store := NewStore(db)

	secret, err := LoadHMACSecret()
	if err != nil {
		return nil, fmt.Errorf("NewProvider: %w", err)
	}

	privateKey, err := crypto.LoadPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("NewProvider: %w", err)
	}

	return compose.Compose(
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
		&Hasher{},
		compose.OAuth2ClientCredentialsGrantFactory,
	), nil
}
