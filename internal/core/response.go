package core

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/KL-Engineering/oauth2-server/internal/errorsx"
)

func JSONResponse(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("ERROR: JSON Marshal: %v", err)
		InternalErrorResponse(w)
	}
}

func BadRequestResponse(w http.ResponseWriter, err errorsx.Error) {
	w.WriteHeader(http.StatusBadRequest)
	JSONResponse(w, errorsx.Errors{
		Errors: []errorsx.Error{
			// NB: Will probably need to extend this for multiple params, but for now we only need to
			// support a single bad param because of our request bodies
			err,
		},
	})
}

func NotFoundResponse(w http.ResponseWriter, resource string) {
	w.WriteHeader(http.StatusNotFound)
	JSONResponse(w, errorsx.Errors{
		Errors: []errorsx.Error{errorsx.NotFoundError(resource)},
	})
}

func MethodNotAllowedResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	JSONResponse(w, errorsx.Errors{
		Errors: []errorsx.Error{
			errorsx.MethodNotAllowedError(),
		},
	})
}

func InternalErrorResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	// NB: Here we duplicate `JSONResponse` to avoid an infinite loop
	// incase we break `errorsx.InternalError()`
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(errorsx.Errors{
		Errors: []errorsx.Error{errorsx.InternalError()},
	}); err != nil {
		log.Printf("ERROR: JSON Marshal: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
