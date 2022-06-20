package monitoring

import (
	"net/http"

	"github.com/KL-Engineering/oauth2-server/internal/core"
	"github.com/julienschmidt/httprouter"
)

type HealthResponse struct {
	Status string `json:"status"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	core.JSONResponse(w, &HealthResponse{
		Status: "OK",
	})
}
