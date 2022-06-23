package oauth2

import (
	"context"
	"time"

	"github.com/KL-Engineering/oauth2-server/internal/client"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
)

// The required sub-interfaces for `client_credentials` grant with JWTs
type FositeStore interface {
	fosite.ClientManager
	oauth2.AccessTokenStorage
}

type Store struct {
	repo *client.Repository
}

var _ FositeStore = (*Store)(nil)

func NewStore(db *dynamodb.Client) *Store {
	// TODO pass the repository directly (or some kind of "Registry" object which includes the repository)
	return &Store{repo: client.NewRepository(db)}
}

func (s *Store) GetClient(ctx context.Context, id string) (fosite.Client, error) {
	client, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fosite.ErrNotFound
	}
	return NewFositeClient(client), nil
}

// NB: No-op as we haven't implemented JTI based blacklists
func (s *Store) ClientAssertionJWTValid(_ context.Context, jti string) error {
	return nil
}

// NB: No-op as we haven't implemented JTI based blacklists
func (s *Store) SetClientAssertionJWT(_ context.Context, jti string, exp time.Time) error {
	return nil
}

// NB: No-op as we are using stateless JWTs
func (s *Store) CreateAccessTokenSession(ctx context.Context, signature string, request fosite.Requester) (err error) {
	return nil
}

// NB: No-op as we are using stateless JWTs
func (s *Store) DeleteAccessTokenSession(ctx context.Context, signature string) error {
	return nil
}

func (s *Store) GetAccessTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	return &fosite.Request{Session: session}, nil
}
