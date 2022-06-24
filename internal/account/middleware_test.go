package account

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KL-Engineering/oauth2-server/internal/errorsx"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareWithHeader(t *testing.T) {
	a := assert.New(t)

	accountID := uuid.NewString()
	w := httptest.NewRecorder()

	router := httprouter.New()
	router.GET("/", Middleware(func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		id := GetAccountIdFromCtx(r.Context())
		a.Equal(accountID, id)

		_, _ = w.Write([]byte("OK"))
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Add(IDHeader, accountID)

	router.ServeHTTP(w, r)

	res := w.Result()

	a.Equal(http.StatusOK, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	a.NoError(err)
	a.Equal([]byte("OK"), body)
}

func TestMiddlewareWithoutHeader(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()

	router := httprouter.New()
	router.GET("/", Middleware(func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, _ = w.Write([]byte("OK"))
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	router.ServeHTTP(w, r)

	res := w.Result()

	a.Equal(http.StatusBadRequest, res.StatusCode)

	var response errorsx.Errors
	a.NoError(json.NewDecoder(res.Body).Decode(&response))

	a.Equal(
		errorsx.Errors(
			errorsx.Errors{
				Errors: []errorsx.Error{
					{
						Category: "INVALID_REQUEST",
						Code:     "REQUIRED_HEADER",
						Message:  "Header 'X-Account-ID' is required.",
					},
				},
			},
		),
		response,
	)
}
