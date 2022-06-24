package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KL-Engineering/oauth2-server/internal/errorsx"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestNotFound(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/foo/bar", nil)

	router := httprouter.New()
	NewHandler().SetupRouter(router)
	router.ServeHTTP(w, r)
	res := w.Result()

	a.Equal(http.StatusNotFound, res.StatusCode)

	var response errorsx.Errors
	a.NoError(json.NewDecoder(res.Body).Decode(&response))

	a.Equal(
		errorsx.Errors(
			errorsx.Errors{
				Errors: []errorsx.Error{
					{
						Category: "NOT_FOUND",
						Code:     "NOT_FOUND",
						Message:  "Resource '/foo/bar' not found.",
					},
				},
			},
		),
		response,
	)
}

func TestMethodNotAllowed(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/foo/bar", nil)

	router := httprouter.New()
	NewHandler().SetupRouter(router)
	router.DELETE("/foo/bar", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.Fail("Should not be called")
	})
	router.ServeHTTP(w, r)
	res := w.Result()

	a.Equal(http.StatusMethodNotAllowed, res.StatusCode)

	var response errorsx.Errors
	a.NoError(json.NewDecoder(res.Body).Decode(&response))

	a.Equal(
		errorsx.Errors(
			errorsx.Errors{
				Errors: []errorsx.Error{
					{
						Category: "INVALID_REQUEST",
						Code:     "INVALID_METHOD",
						Message:  "Method not allowed.",
					},
				},
			},
		),
		response,
	)
}
