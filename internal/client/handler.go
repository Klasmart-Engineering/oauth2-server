package client

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/KL-Engineering/oauth2-server/internal/account"
	"github.com/KL-Engineering/oauth2-server/internal/crypto"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type Handler struct {
	repo Repository
}

func NewHandler(client *dynamodb.Client) *Handler {
	return &Handler{
		repo: *NewRepository(client),
	}
}

func (h *Handler) SetupRouter(router *httprouter.Router) {
	router.POST("/clients", h.Create())
}

type CreateClientRequest struct {
	Name string `json:"name"`
}

type CreateClientResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Secret string `json:"secret"`
}

func (h *Handler) Create() httprouter.Handle {
	return account.Middleware(func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := r.Context()
		account_id := account.GetAccountIdFromCtx(ctx)

		var req CreateClientRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// TODO will in future need to accept external `android_id` from accounts rather than generating one here
		// but this functionality is not available currently
		android_id, err := uuid.NewRandom()
		if err != nil {
			log.Printf("ERROR: Failed to create android_id")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		secret, err := crypto.GenerateSecret()
		if err != nil {
			log.Printf("ERROR: crypto.GenerateSecret: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		client, err := h.repo.Create(ctx, CreateOptions{
			secret:     secret,
			name:       req.Name,
			android_id: android_id.String(),
			account_id: account_id,
		})
		if err != nil {
			// TODO specific codes in case of bad request
			log.Printf("ERROR: Create Client: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		clientResponse := CreateClientResponse{
			ID:     client.ID,
			Name:   client.Name,
			Secret: secret,
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(clientResponse)
		if err != nil {
			log.Printf("ERROR: JSON Marshal ClientResponse: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("INFO: Created Client(id=%s)", client.ID)
	})
}
