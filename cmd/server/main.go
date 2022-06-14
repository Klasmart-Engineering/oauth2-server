package main

import (
	"log"
	"net/http"

	"github.com/KL-Engineering/oauth2-server/internal/client"
	"github.com/KL-Engineering/oauth2-server/internal/crypto"
	"github.com/KL-Engineering/oauth2-server/internal/monitoring"
	"github.com/KL-Engineering/oauth2-server/internal/oauth2"
	"github.com/KL-Engineering/oauth2-server/internal/storage"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
)

func NewServer(d *dynamodb.Client) *http.Server {
	router := httprouter.New()

	router.GET("/health", monitoring.HealthHandler)

	router.POST("/oauth2/token", oauth2.TokenHandler)

	router.GET("/.well-known/jwks.json", crypto.JWKS())

	clientHandler := client.NewHandler(d)
	clientHandler.SetupRouter(router)

	return &http.Server{
		Addr:    "localhost:8080",
		Handler: router,
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		// Only necessary for local development
		log.Print("INFO: Did not load .env file")
	}

	dynamodbClient, err := storage.NewDynamoDBClient()
	if err != nil {
		log.Fatalf("Unable to load DynamoDB: %v", err)
	}

	s := NewServer(dynamodbClient)

	log.Println("Listening for requests at http://localhost:8080")
	log.Fatal(s.ListenAndServe())
}
