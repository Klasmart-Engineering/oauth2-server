package monitoring

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func HealthHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(struct {
		Status string `json:"status"`
	}{
		Status: "OK",
	},
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
