package crypto

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func testChdir(t *testing.T, dir string) func() {
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("err: %v", err)
	}

	return func() {
		if err := os.Chdir(old); err != nil {
			t.Fatalf("err: %v", err)
		}
	}
}

func TestJWKS(t *testing.T) {
	a := assert.New(t)

	// Temporarily chdir to project root, otherwise relative PEM filepaths
	// can't be loaded
	defer testChdir(t, "../..")()

	router := httprouter.New()
	router.GET("/", JWKS())

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	res := w.Result()

	a.Equal(http.StatusOK, res.StatusCode)

	var response map[string]interface{}
	err := json.NewDecoder(res.Body).Decode(&response)
	a.NoError(err)

	a.Len(response["keys"], 1)

	key := response["keys"].([]interface{})[0].(map[string]interface{})
	a.Equal(KID, key["kid"])
	a.Equal("RS256", key["alg"])
}
