package monitoring

import (
	"net/http"

	"github.com/KL-Engineering/oauth2-server/internal/core"
	"github.com/julienschmidt/httprouter"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) SetupRouter(router *httprouter.Router) {
	router.GET("/health", h.HealthHandler())
}

type HealthResponse struct {
	Status string `json:"status"`
}

func (h *Handler) HealthHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		core.JSONResponse(w, &HealthResponse{
			Status: "OK",
		})
	}
}
