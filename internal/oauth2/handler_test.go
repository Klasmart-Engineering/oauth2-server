package oauth2

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"
	"time"

	"github.com/KL-Engineering/oauth2-server/internal/client"
	"github.com/KL-Engineering/oauth2-server/internal/crypto"
	"github.com/KL-Engineering/oauth2-server/internal/storage"
	"github.com/KL-Engineering/oauth2-server/internal/test"
	"github.com/KL-Engineering/oauth2-server/internal/utils"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"gopkg.in/square/go-jose.v2/jwt"
)

const (
	testSecret = "pa$$word"
)

type Setup struct {
	db *dynamodb.Client
	r  *httprouter.Router
}

func TestROPCInvalid(t *testing.T) {
	a := assert.New(t)
	s := setup(t)

	s.r.GET("/callback", createCallbackHandler(a))

	srv := httptest.NewServer(s.r)
	defer srv.Close()

	client := createClient(a, s.db)

	config := newOAuth2Config(client.ID, srv.URL)

	tokenResponse, err := config.PasswordCredentialsToken(context.Background(), "fake-user", "fake-password")
	a.Nil(tokenResponse)
	a.Error(err)
}

func TestAuthorizationCodeInvalid(t *testing.T) {
	a := assert.New(t)
	s := setup(t)

	s.r.GET("/callback", createCallbackHandler(a))

	srv := httptest.NewServer(s.r)
	defer srv.Close()

	client := createClient(a, s.db)

	config := newOAuth2Config(client.ID, srv.URL)

	url := config.AuthCodeURL("state")
	response, err := http.Get(url)
	a.Nil(err)
	a.Equal(http.StatusNotFound, response.StatusCode)
}

func TestClientCredentialsInvalidCredentials(t *testing.T) {
	a := assert.New(t)
	s := setup(t)

	srv := httptest.NewServer(s.r)
	defer srv.Close()

	client := createClient(a, s.db)

	conf := clientcredentials.Config{
		ClientID:     client.ID,
		ClientSecret: "incorrect-password",
		Scopes:       []string{""},
		TokenURL:     fmt.Sprintf("%s/oauth2/token", srv.URL),
	}

	_, err := conf.Token(context.Background())
	a.Error(err)
}

func TestClientCredentialsNonexistentClient(t *testing.T) {
	s := setup(t)

	srv := httptest.NewServer(s.r)
	defer srv.Close()

	conf := clientcredentials.Config{
		ClientID:     uuid.NewString(),
		ClientSecret: "pa$$word",
		Scopes:       []string{""},
		TokenURL:     fmt.Sprintf("%s/oauth2/token", srv.URL),
	}

	_, err := conf.Token(context.Background())
	assert.Error(t, err)
}

func TestClientCredentialsInvalidOfflineAccess(t *testing.T) {
	a := assert.New(t)
	s := setup(t)

	srv := httptest.NewServer(s.r)
	defer srv.Close()

	client := createClient(a, s.db)

	conf := clientcredentials.Config{
		ClientID:     client.ID,
		ClientSecret: testSecret,
		Scopes:       []string{"offline_access"},
		TokenURL:     fmt.Sprintf("%s/oauth2/token", srv.URL),
	}

	tokenResponse, err := conf.Token(context.Background())
	a.Error(err)
	a.Nil(tokenResponse)
}

func TestClientCredentialsValidAccessToken(t *testing.T) {
	a := assert.New(t)
	s := setup(t)

	srv := httptest.NewServer(s.r)
	defer srv.Close()

	client := createClient(a, s.db)

	conf := clientcredentials.Config{
		ClientID:     client.ID,
		ClientSecret: testSecret,
		Scopes:       []string{""},
		TokenURL:     fmt.Sprintf("%s/oauth2/token", srv.URL),
	}

	tokenResponse, err := conf.Token(context.Background())
	a.NoError(err)

	a.Equal("bearer", tokenResponse.TokenType)

	assertTokenPayloadValid(a, tokenResponse, client)

	assertTokenHeaderValid(a, tokenResponse)

	assertTokenSignatureValid(t, tokenResponse.AccessToken)
}

func setup(t *testing.T) *Setup {
	db, err := storage.NewDynamoDBClient()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// Temporarily chdir to project root, otherwise relative PEM filepaths
	// can't be loaded
	defer test.Chdir(t, "../..")()
	p, err := NewProvider(db)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	r := httprouter.New()
	NewHandler(p).SetupRouter(r)

	return &Setup{
		db,
		r,
	}
}

func createClient(a *assert.Assertions, db *dynamodb.Client) *client.Client {
	r := client.NewRepository(db)
	client, err := r.Create(context.Background(), client.CreateOptions{
		Secret:    testSecret,
		Name:      "Client",
		AccountID: uuid.NewString(),
		AndroidID: uuid.NewString(),
	})
	a.NoError(err)
	return client
}

func newOAuth2Config(clientID string, baseURL string) oauth2.Config {
	return oauth2.Config{
		ClientID:     clientID,
		ClientSecret: testSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth2/auth", baseURL),
			TokenURL: fmt.Sprintf("%s/oauth2/token", baseURL),
		},
		// In reality this would be a separate server, but for simplicity we'll use the same server
		RedirectURL: fmt.Sprintf("%s/callback", baseURL),
		Scopes:      []string{},
	}
}

func createCallbackHandler(a *assert.Assertions) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		reqDump, err := httputil.DumpRequestOut(r, true)
		a.NoError(err)

		fmt.Printf("REQUEST:\n%s", string(reqDump))
		a.FailNow("oauth2 callback should not be called")
	}
}

func assertTokenPayloadValid(a *assert.Assertions, tokenResponse *oauth2.Token, client *client.Client) {
	token, err := jwt.ParseSigned(tokenResponse.AccessToken)
	a.NoError(err)

	claims := make(map[string]interface{})
	err = token.UnsafeClaimsWithoutVerification(&claims)
	a.NoError(err)

	a.Equal([]interface{}{}, claims["aud"], "audience is empty")

	now := time.Now()
	exp := test.ParseUnix(claims["exp"].(float64))
	a.WithinDuration(exp, now.Add(time.Minute*15), time.Second)
	a.WithinDuration(tokenResponse.Expiry, exp, time.Second)

	a.WithinDuration(test.ParseUnix(claims["iat"].(float64)), now, time.Second)

	a.Equal("https://platform.kidsloop.live", claims["iss"])
	a.True(utils.IsUUID(claims["jti"].(string)))
	a.Equal([]interface{}{}, claims["scp"], "scopes is empty")
	a.Equal(client.ID, claims["sub"])

	a.Equal(client.Account_ID, claims["account_id"])
	a.Equal(client.Android_ID, claims["android_id"])
	// TODO `subscription_id` claim
}

func assertTokenHeaderValid(a *assert.Assertions, tokenResponse *oauth2.Token) {
	headersB64 := strings.Split(tokenResponse.AccessToken, ".")[0]
	var headers map[string]interface{}
	err := json.NewDecoder(
		base64.NewDecoder(base64.RawURLEncoding, strings.NewReader(headersB64)),
	).Decode(&headers)
	a.NoError(err)

	a.Equal("RS256", headers["alg"])
	a.Equal("JWT", headers["typ"])
	a.Equal(crypto.KID, headers["kid"])
}

func assertTokenSignatureValid(t *testing.T, rawToken string) {
	defer test.Chdir(t, "../..")()
	jwks, err := crypto.JWKS()
	assert.NoError(t, err)

	token, err := jwt.ParseSigned(rawToken)
	assert.NoError(t, err)

	claims := jwt.Claims{}
	assert.NoError(t, token.Claims(jwks, &claims))
}
