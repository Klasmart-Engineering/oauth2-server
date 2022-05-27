package main

import (
	"log"
	"net/http"

	"github.com/KL-Engineering/oauth2-server/internal/keys"
	"github.com/KL-Engineering/oauth2-server/internal/monitoring"
	"github.com/KL-Engineering/oauth2-server/internal/oauth2"
	"github.com/julienschmidt/httprouter"
)

func NewServer() *http.Server {
	router := httprouter.New()

	router.GET("/health", monitoring.HealthHandler)

	router.POST("/oauth2/token", oauth2.TokenHandler)

	router.GET("/.well-known/jwks.json", keys.JWKS())

	return &http.Server{
		Addr:    "localhost:8080",
		Handler: router,
	}
}

func main() {
	s := NewServer()

	log.Println("Listening for requests at http://localhost:8080")
	log.Fatal(s.ListenAndServe())
}
