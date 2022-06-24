package account

import (
	"context"
	"net/http"

	"github.com/KL-Engineering/oauth2-server/internal/core"
	"github.com/KL-Engineering/oauth2-server/internal/errorsx"
	"github.com/julienschmidt/httprouter"
)

const IDHeader = "X-Account-ID"

// Add X-Account-Id header to request context
// (set by microgateway from parsed `android_id` claim in JWT)
func Middleware(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		account_id := r.Header.Get(IDHeader)
		if account_id == "" {
			core.BadRequestResponse(
				w,
				errorsx.RequiredHeaderError(IDHeader),
			)
			return
		}

		ctx := context.WithValue(r.Context(), accountCtxKey, account_id)
		next(w, r.WithContext(ctx), ps)
	}
}
