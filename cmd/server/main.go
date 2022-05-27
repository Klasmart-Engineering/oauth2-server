package main

import (
	"log"
	"net/http"

	"github.com/KL-Engineering/oauth2-server/internal/monitoring"
	"github.com/julienschmidt/httprouter"
)

func NewServer() *http.Server {
	router := httprouter.New()

	router.GET("/health", monitoring.HealthHandler)

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
