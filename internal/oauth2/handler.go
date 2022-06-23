package oauth2

import (
	"log"
	"net/http"

	"github.com/KL-Engineering/oauth2-server/internal/crypto"
	"github.com/julienschmidt/httprouter"
	"github.com/ory/fosite"
)

type Handler struct {
	provider fosite.OAuth2Provider
}

func NewHandler(provider fosite.OAuth2Provider) *Handler {
	return &Handler{provider: provider}
}

func (h *Handler) SetupRouter(router *httprouter.Router) {
	router.POST("/oauth2/token", h.Token)
}

func (h *Handler) Token(rw http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	ctx := req.Context()

	session := NewSession("")

	// This will create an access request object and iterate through the registered TokenEndpointHandlers to validate the request.
	accessRequest, err := h.provider.NewAccessRequest(ctx, req, session)

	// Catch any errors, e.g.:
	// * unknown client
	// * invalid redirect
	// * ...
	if err != nil {
		log.Printf("Error occurred in NewAccessRequest: %+v", err)
		h.provider.WriteAccessError(rw, accessRequest, err)
		return
	}

	// If this is a client_credentials grant, grant all requested scopes
	// NewAccessRequest validated that all requested scopes the client is allowed to perform
	// based on configured scope matching strategy.
	if accessRequest.GetGrantTypes().ExactOne("client_credentials") {
		for _, scope := range accessRequest.GetRequestedScopes() {
			accessRequest.GrantScope(scope)
		}

		session.Subject = accessRequest.GetClient().GetID()
		// TODO remove hardcode
		session.KID = crypto.KID
	}

	// Next we create a response for the access request. Again, we iterate through the TokenEndpointHandlers
	// and aggregate the result in response.
	response, err := h.provider.NewAccessResponse(ctx, accessRequest)
	if err != nil {
		log.Printf("Error occurred in NewAccessResponse: %+v", err)
		h.provider.WriteAccessError(rw, accessRequest, err)
		return
	}

	// All done, send the response.
	h.provider.WriteAccessResponse(rw, accessRequest, response)
}
