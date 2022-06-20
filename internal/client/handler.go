package client

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/KL-Engineering/oauth2-server/internal/account"
	"github.com/KL-Engineering/oauth2-server/internal/core"
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
	router.GET("/clients", h.List())
	router.POST("/clients", h.Create())
	router.GET("/clients/:id", h.Get())
	router.DELETE("/clients/:id", h.Delete())
	router.PATCH("/clients/:id", h.Update())
	router.PATCH("/clients/:id/secret", h.RegenerateSecret())
}

type ListResponse struct {
	Records []Client `json:"records"`
}

func (h *Handler) List() httprouter.Handle {
	return account.Middleware(func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// NB: This endpoint could easily support limit pagination
		// However, we would need to implement cursor based pagination to have "offset" functionality
		ctx := r.Context()
		account_id := account.GetAccountIdFromCtx(ctx)

		clients, err := h.repo.List(ctx, ListOptions{account_id: account_id})
		if err != nil {
			log.Printf("ERROR: List Client: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		core.JSONResponse(w, ListResponse{
			Records: clients,
		})
	})
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

		log.Printf("INFO: Created Client(id=%s)", client.ID)

		clientResponse := CreateClientResponse{
			ID:     client.ID,
			Name:   client.Name,
			Secret: secret,
		}

		w.WriteHeader(http.StatusCreated)
		core.JSONResponse(w, clientResponse)
	})
}

func (h *Handler) Get() httprouter.Handle {
	return account.Middleware(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()
		account_id := account.GetAccountIdFromCtx(ctx)
		id := ps.ByName("id")

		client, err := h.repo.Get(ctx, GetOptions{account_id: account_id, id: id})
		if err != nil {
			if err == core.ErrNotFound {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				log.Printf("ERROR: Get Client: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		core.JSONResponse(w, client)
	})
}

func (h *Handler) Delete() httprouter.Handle {
	return account.Middleware(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()
		account_id := account.GetAccountIdFromCtx(ctx)
		id := ps.ByName("id")

		err := h.repo.Delete(ctx, DeleteOptions{account_id: account_id, id: id})
		if err != nil {
			if err == core.ErrNotFound {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				log.Printf("ERROR: Delete Client: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusNoContent)
		w.Header().Set("Content-Type", "application/json")
	})
}

type UpdateClientRequest struct {
	Name string `json:"name"`
}

func (h *Handler) Update() httprouter.Handle {
	return account.Middleware(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()
		account_id := account.GetAccountIdFromCtx(ctx)
		id := ps.ByName("id")

		var req UpdateClientRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		client, err := h.repo.Update(ctx, UpdateOptions{account_id: account_id, id: id, name: req.Name})

		if err != nil {
			if err == core.ErrNotFound {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				log.Printf("ERROR: Update Client: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		core.JSONResponse(w, client)
	})
}

type RegenerateSecretResponse struct {
	Secret string `json:"secret"`
}

func (h *Handler) RegenerateSecret() httprouter.Handle {
	return account.Middleware(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()
		account_id := account.GetAccountIdFromCtx(ctx)
		id := ps.ByName("id")

		secret, err := crypto.GenerateSecret()
		if err != nil {
			log.Printf("ERROR: crypto.GenerateSecret: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = h.repo.Update(ctx, UpdateOptions{account_id: account_id, id: id, secret: secret})

		if err != nil {
			if err == core.ErrNotFound {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				log.Printf("ERROR: Update Client: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		core.JSONResponse(w, RegenerateSecretResponse{
			Secret: secret,
		})
	})
}
