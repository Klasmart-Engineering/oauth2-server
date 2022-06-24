package monitoring

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)

	router := httprouter.New()
	NewHandler().SetupRouter(router)
	router.ServeHTTP(w, r)
	res := w.Result()

	a.Equal(http.StatusOK, res.StatusCode)

	var response HealthResponse
	a.NoError(json.NewDecoder(res.Body).Decode(&response))

	a.Equal(
		HealthResponse{
			Status: "OK",
		},
		response,
	)
}
