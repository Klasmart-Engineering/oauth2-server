package crypto

import (
	"net/http"

	"github.com/KL-Engineering/oauth2-server/internal/core"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/square/go-jose.v2"
)

type Handler struct {
	jwks *jose.JSONWebKeySet
}

func NewHandler(jwks *jose.JSONWebKeySet) *Handler {
	return &Handler{jwks}
}

func (h *Handler) SetupRouter(router *httprouter.Router) {
	router.GET("/.well-known/jwks.json", h.WellKnown())
}

func (h *Handler) WellKnown() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		core.JSONResponse(w, h.jwks)
	}
}
