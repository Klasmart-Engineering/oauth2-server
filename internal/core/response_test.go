package core

import (
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KL-Engineering/oauth2-server/internal/errorsx"
	"github.com/KL-Engineering/oauth2-server/internal/errorsx/category"
	"github.com/KL-Engineering/oauth2-server/internal/errorsx/code"
	"github.com/stretchr/testify/assert"
)

func TestNotFoundResponse(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()

	NotFoundResponse(w, "Foo")

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
						Message:  "Resource 'Foo' not found.",
					},
				},
			},
		),
		response,
	)
}

func TestInternalErrorResponse(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()

	InternalErrorResponse(w)

	res := w.Result()

	a.Equal(http.StatusInternalServerError, res.StatusCode)

	var response errorsx.Errors
	a.NoError(json.NewDecoder(res.Body).Decode(&response))

	a.Equal(
		errorsx.Errors(
			errorsx.Errors{
				Errors: []errorsx.Error{
					{
						Category: category.INTERNAL,
						Code:     code.INTERNAL,
						Message:  "Internal server error.",
					},
				},
			},
		),
		response,
	)
}

// A JSON response that can't be marshalled returns an `InternalErrorResponse`
func TestJSONResponseInvalid(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()

	// An unmarshable value
	JSONResponse(w, math.Inf(1))

	res := w.Result()

	a.Equal(http.StatusInternalServerError, res.StatusCode)

	var response errorsx.Errors
	a.NoError(json.NewDecoder(res.Body).Decode(&response))

	a.Equal(
		errorsx.Errors(
			errorsx.Errors{
				Errors: []errorsx.Error{
					{
						Category: category.INTERNAL,
						Code:     code.INTERNAL,
						Message:  "Internal server error.",
					},
				},
			},
		),
		response,
	)
}

func TestJSONResponseValid(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()

	JSONResponse(w, struct {
		Foo string `json:"foo"`
	}{
		Foo: "bar",
	})

	res := w.Result()

	a.Equal(http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	a.NoError(err)

	a.JSONEq(
		`{"foo": "bar"}`,
		string(body),
	)
}
