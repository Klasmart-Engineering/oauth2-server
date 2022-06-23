package crypto

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KL-Engineering/oauth2-server/internal/test"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestJWKS(t *testing.T) {
	a := assert.New(t)

	// Temporarily chdir to project root, otherwise relative PEM filepaths
	// can't be loaded
	defer test.Chdir(t, "../..")()

	jwks, err := JWKS()
	a.NoError(err)

	router := httprouter.New()
	NewHandler(jwks).SetupRouter(router)

	r := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	res := w.Result()

	a.Equal(http.StatusOK, res.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	a.NoError(err)

	a.Len(response["keys"], 1)

	key := response["keys"].([]interface{})[0].(map[string]interface{})
	a.Equal(KID, key["kid"])
	a.Equal("RS256", key["alg"])
}
