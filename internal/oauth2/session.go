package oauth2

import (
	"time"

	"github.com/KL-Engineering/oauth2-server/internal/crypto"
	"github.com/mohae/deepcopy"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/jwt"
)

const (
	ISSUER = "https://platform.kidsloop.live"
)

type Session struct {
	*openid.DefaultSession `json:"idToken"`
	Extra                  map[string]interface{} `json:"extra"`
	KID                    string
	AccountID              string
	AndroidID              string
	// TODO: SubscriptionID
}

func NewSession(subject string) *Session {
	return &Session{
		DefaultSession: &openid.DefaultSession{
			Claims: &jwt.IDTokenClaims{
				Issuer:      ISSUER,
				Subject:     subject,
				Audience:    []string{ISSUER},
				ExpiresAt:   time.Now().Add(time.Hour * 1),
				IssuedAt:    time.Now(),
				RequestedAt: time.Now(),
				AuthTime:    time.Now(),
			},
			Headers: &jwt.Headers{
				Extra: make(map[string]interface{}),
			},
			Subject: subject,
		},
		Extra: map[string]interface{}{},
	}
}

// Populate Session object based on OAuth2 Client
func (s *Session) WithClient(client fosite.Client) {
	s.Subject = client.GetID()
	// TODO remove hardcode
	s.KID = crypto.KID

	s.AccountID = client.(CustomFositeClient).GetAccountID()
	s.AndroidID = client.(CustomFositeClient).GetAndroidID()
}

func (s *Session) GetJWTClaims() jwt.JWTClaimsContainer {
	claims := &jwt.JWTClaims{
		Subject:   s.Subject,
		Audience:  s.DefaultSession.Claims.Audience,
		Issuer:    s.DefaultSession.Claims.Issuer,
		ExpiresAt: s.GetExpiresAt(fosite.AccessToken),
		IssuedAt:  time.Now(),

		// The JTI MUST NOT BE FIXED or refreshing tokens will yield the SAME token
		// JTI:       s.JTI,

		// These are set by the DefaultJWTStrategy
		// Scope:     s.Scope,

		// Setting these here will cause the token to have the same iat/nbf values always
		// IssuedAt:  s.DefaultSession.Claims.IssuedAt,
		// NotBefore: s.DefaultSession.Claims.IssuedAt,
	}

	if claims.Extra == nil {
		claims.Extra = map[string]interface{}{}
	}
	claims.Extra["account_id"] = s.AccountID
	claims.Extra["android_id"] = s.AndroidID
	return claims
}

func (s *Session) GetJWTHeader() *jwt.Headers {
	return &jwt.Headers{
		Extra: map[string]interface{}{"kid": s.KID},
	}
}

func (s *Session) Clone() fosite.Session {
	if s == nil {
		return nil
	}

	return deepcopy.Copy(s).(fosite.Session)
}
