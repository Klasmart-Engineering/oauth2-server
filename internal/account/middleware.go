package account

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Add X-Account-Id header to request context
// (set by microgateway from parsed `android_id` claim in JWT)
func Middleware(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		account_id := r.Header.Get("X-Account-Id")
		if account_id == "" {
			http.Error(w, "Missing X-Account-Id header", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), accountCtxKey, account_id)
		next(w, r.WithContext(ctx), ps)
	}
}
