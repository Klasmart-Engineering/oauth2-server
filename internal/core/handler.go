package core

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) SetupRouter(router *httprouter.Router) {
	router.NotFound = h.NotFound()
	router.MethodNotAllowed = h.MethodNotAllowed()
}

func (h *Handler) NotFound() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NotFoundResponse(w, r.URL.Path)
	})
}

func (h *Handler) MethodNotAllowed() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MethodNotAllowedResponse(w)
	})
}
