package oauth2

import (
	"github.com/KL-Engineering/oauth2-server/internal/client"
	"github.com/ory/fosite"
)

type FositeClient struct {
	model *client.Client
}

type CustomFositeClient interface {
	fosite.Client
	GetAccountID() string
	GetAndroidID() string
}

var _ fosite.Client = (*FositeClient)(nil)

func NewFositeClient(model *client.Client) *FositeClient {
	return &FositeClient{model: model}
}

func (c *FositeClient) GetID() string {
	return c.model.ID
}

func (c *FositeClient) GetHashedSecret() []byte {
	return []byte(c.model.SecretHash)
}

func (c *FositeClient) GetRedirectURIs() []string {
	// Currently no public client support
	return []string{}
}

func (c *FositeClient) GetGrantTypes() fosite.Arguments {
	// Currently only support client_credentials
	return []string{"client_credentials"}
}

func (c *FositeClient) GetScopes() fosite.Arguments {
	return []string{""}
}

func (c *FositeClient) GetResponseTypes() fosite.Arguments {
	// Currently only support access_tokens
	return []string{"token"}
}

func (c *FositeClient) IsPublic() bool {
	// Currently no public client support
	return false
}

func (c *FositeClient) GetAudience() fosite.Arguments {
	// Currently only support a single (global) platform "audience"
	return []string{}
}

func (c *FositeClient) GetAccountID() string {
	return c.model.AccountID
}

func (c *FositeClient) GetAndroidID() string {
	return c.model.AndroidID
}
