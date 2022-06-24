package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/KL-Engineering/oauth2-server/internal/account"
	"github.com/KL-Engineering/oauth2-server/internal/core"
	"github.com/KL-Engineering/oauth2-server/internal/storage"
	"github.com/KL-Engineering/oauth2-server/internal/utils"
	"github.com/alexedwards/argon2id"
	"github.com/julienschmidt/httprouter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestListEmpty(t *testing.T) {
	a := assert.New(t)

	accountID := uuid.New().String()

	dynamoClient := utils.Must(storage.NewDynamoDBClient())

	router := httprouter.New()
	h := NewHandler(dynamoClient)
	h.SetupRouter(router)

	r := httptest.NewRequest(http.MethodGet, "/clients", nil)
	r.Header.Add(account.IDHeader, accountID)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	res := w.Result()

	var response ListResponse
	err := json.NewDecoder(res.Body).Decode(&response)
	a.NoError(err)

	a.Equal(ListResponse{
		Records: []Client{},
	}, response)
	a.Equal(http.StatusOK, res.StatusCode)
}

func TestList(t *testing.T) {
	a := assert.New(t)

	accountID := uuid.New().String()

	dynamoClient := utils.Must(storage.NewDynamoDBClient())

	repo := NewRepository(dynamoClient)

	client1, err := repo.Create(context.Background(), CreateOptions{
		Secret:    "pa$$word",
		Name:      "Test1",
		AndroidID: uuid.NewString(),
		AccountID: accountID,
	})
	a.NoError(err)
	client2, err := repo.Create(context.Background(), CreateOptions{
		Secret:    "pa$$word",
		Name:      "Test2",
		AndroidID: uuid.NewString(),
		AccountID: accountID,
	})
	a.NoError(err)

	clients := []Client{
		*client1,
		*client2,
	}
	sort.Slice(clients, func(i, j int) bool {
		return clients[i].ID < clients[j].ID
	})

	router := httprouter.New()
	h := NewHandler(dynamoClient)
	h.SetupRouter(router)

	r := httptest.NewRequest(http.MethodGet, "/clients", nil)
	r.Header.Add(account.IDHeader, accountID)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	res := w.Result()

	expected, err := json.Marshal(ListResponse{
		Records: clients,
	})
	a.NoError(err)

	actual, err := io.ReadAll(res.Body)
	a.NoError(err)

	a.JSONEq(string(expected), string(actual))
	a.Equal(http.StatusOK, res.StatusCode)
}

func TestCreateNoBody(t *testing.T) {
	a := assert.New(t)

	db := utils.Must(storage.NewDynamoDBClient())
	h := NewHandler(db)
	router := httprouter.New()
	h.SetupRouter(router)

	r := httptest.NewRequest(http.MethodPost, "/clients", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	res := w.Result()

	a.Equal(res.StatusCode, http.StatusBadRequest)
}

func TestCreateValid(t *testing.T) {
	a := assert.New(t)

	db := utils.Must(storage.NewDynamoDBClient())
	h := NewHandler(db)
	router := httprouter.New()
	h.SetupRouter(router)

	body := &CreateClientRequest{Name: "Test client"}
	buf := new(bytes.Buffer)
	a.NoError(json.NewEncoder(buf).Encode(body))

	r := httptest.NewRequest(http.MethodPost, "/clients", buf)
	w := httptest.NewRecorder()

	accountID := uuid.New().String()
	r.Header.Add(account.IDHeader, accountID)

	router.ServeHTTP(w, r)
	res := w.Result()

	var response CreateClientResponse
	a.NoError(json.NewDecoder(res.Body).Decode(&response))

	a.True(utils.IsUUID(response.ID))
	a.Equal(response.Name, body.Name)
	a.Len(response.Secret, 40)

	a.Equal(res.StatusCode, http.StatusCreated)

	output, err := db.GetItem(context.Background(), &dynamodb.GetItemInput{Key: map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("Account#%s", accountID)},
		"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("Client#%s", response.ID)},
	},
		TableName: aws.String(tableName),
	})
	a.NoError(err)

	var client Client
	a.Nil(attributevalue.UnmarshalMap(output.Item, &client))

	a.Equal(client.AccountID, accountID)
	a.True(utils.IsUUID(client.AndroidID))
	a.Equal(client.ID, response.ID)
	a.Equal(client.Name, response.Name)
	a.Equal(client.SecretPrefix, response.Secret[:secretPrefixLength])

	passwordMatch := utils.Must(argon2id.ComparePasswordAndHash(response.Secret, client.SecretHash))
	a.True(passwordMatch)
}

func TestGetNotFound(t *testing.T) {
	a := assert.New(t)

	db := utils.Must(storage.NewDynamoDBClient())
	h := NewHandler(db)
	router := httprouter.New()
	h.SetupRouter(router)

	r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/clients/%s", uuid.NewString()), nil)
	r.Header.Add(account.IDHeader, uuid.NewString())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	res := w.Result()

	a.Equal(http.StatusNotFound, res.StatusCode)
}

func TestGetUnauthorized(t *testing.T) {
	a := assert.New(t)

	db := utils.Must(storage.NewDynamoDBClient())

	router := httprouter.New()
	h := NewHandler(db)
	client, err := h.repo.Create(context.Background(), CreateOptions{
		Secret:    "pa$$word",
		Name:      "Test",
		AndroidID: uuid.NewString(),
		AccountID: uuid.NewString(),
	})
	a.NoError(err)

	r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/clients/%s", client.ID), nil)
	r.Header.Add(account.IDHeader, uuid.NewString())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	res := w.Result()

	a.Equal(http.StatusNotFound, res.StatusCode, "Client belongs to another AccountID")
}

func TestGetValid(t *testing.T) {
	a := assert.New(t)

	// TODO abstract
	accountID := uuid.New().String()

	dynamoClient := utils.Must(storage.NewDynamoDBClient())

	router := httprouter.New()
	h := NewHandler(dynamoClient)
	h.SetupRouter(router)

	client, err := h.repo.Create(
		context.Background(),
		CreateOptions{
			Secret:    "pa$$word",
			Name:      "Test",
			AndroidID: uuid.NewString(),
			AccountID: accountID,
		})
	a.NoError(err)

	r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/clients/%s", client.ID), nil)
	r.Header.Add(account.IDHeader, accountID)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	res := w.Result()

	var response Client
	err = json.NewDecoder(res.Body).Decode(&response)
	a.NoError(err)

	a.Equal(response.ID, client.ID)
	a.Equal(response.Name, client.Name)
	a.Equal(response.SecretPrefix, client.SecretPrefix)
	a.Equal(response.SecretHash, "")

	a.Equal(res.StatusCode, http.StatusOK)
}

func TestDeleteNotFound(t *testing.T) {
	a := assert.New(t)

	accountID := uuid.New().String()

	dynamoClient := utils.Must(storage.NewDynamoDBClient())

	router := httprouter.New()
	h := NewHandler(dynamoClient)
	h.SetupRouter(router)

	r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/clients/%s", uuid.New()), nil)
	r.Header.Add(account.IDHeader, accountID)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	res := w.Result()

	a.Equal(http.StatusNotFound, res.StatusCode)
}

func TestDeleteUnauthorized(t *testing.T) {
	a := assert.New(t)

	db := utils.Must(storage.NewDynamoDBClient())

	router := httprouter.New()
	h := NewHandler(db)
	client, err := h.repo.Create(context.Background(), CreateOptions{
		Secret:    "pa$$word",
		Name:      "Test",
		AndroidID: uuid.NewString(),
		AccountID: uuid.NewString(),
	})
	a.NoError(err)

	r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/clients/%s", client.ID), nil)
	r.Header.Add(account.IDHeader, uuid.NewString())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	res := w.Result()

	a.Equal(http.StatusNotFound, res.StatusCode, "Client belongs to another AccountID")
}

func TestDelete(t *testing.T) {
	a := assert.New(t)

	accountID := uuid.New().String()

	dynamoClient := utils.Must(storage.NewDynamoDBClient())
	router := httprouter.New()
	h := NewHandler(dynamoClient)
	h.SetupRouter(router)

	client, err := h.repo.Create(
		context.Background(), CreateOptions{
			Secret:    "pa$$word",
			Name:      "Test",
			AndroidID: uuid.NewString(),
			AccountID: accountID,
		},
	)
	a.NoError(err)

	r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/clients/%s", client.ID), nil)
	r.Header.Add(account.IDHeader, accountID)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	res := w.Result()

	a.Equal(http.StatusNoContent, res.StatusCode, "First DELETE returns NoContent")

	emptyClient, err := h.repo.Get(context.Background(), GetOptions{AccountID: client.AccountID, ID: client.ID})
	a.Equal(err, core.ErrNotFound, "Client is deleted")
	a.Nil(emptyClient)

	w = httptest.NewRecorder()

	router.ServeHTTP(w, r)

	res = w.Result()

	a.Equal(http.StatusNotFound, res.StatusCode, "Second DELETE returns NotFound")
}

func TestUpdateNotFound(t *testing.T) {
	a := assert.New(t)

	accountID := uuid.New().String()

	dynamoClient := utils.Must(storage.NewDynamoDBClient())

	router := httprouter.New()
	h := NewHandler(dynamoClient)
	h.SetupRouter(router)

	body := &UpdateClientRequest{Name: "Test2"}
	buf := new(bytes.Buffer)
	a.NoError(json.NewEncoder(buf).Encode(body))

	r := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/clients/%s", uuid.New()), buf)
	r.Header.Add(account.IDHeader, accountID)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	res := w.Result()

	a.Equal(http.StatusNotFound, res.StatusCode)
}

func TestUpdateUnauthorized(t *testing.T) {
	a := assert.New(t)

	db := utils.Must(storage.NewDynamoDBClient())

	router := httprouter.New()
	h := NewHandler(db)
	client, err := h.repo.Create(context.Background(), CreateOptions{
		Secret:    "pa$$word",
		Name:      "Test",
		AndroidID: uuid.NewString(),
		AccountID: uuid.NewString(),
	})
	a.NoError(err)

	body := &UpdateClientRequest{Name: "Test2"}
	buf := new(bytes.Buffer)
	a.NoError(json.NewEncoder(buf).Encode(body))

	r := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/clients/%s", client.ID), buf)
	r.Header.Add(account.IDHeader, uuid.NewString())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	res := w.Result()

	a.Equal(http.StatusNotFound, res.StatusCode, "Client belongs to another AccountID")
}

func TestUpdate(t *testing.T) {
	a := assert.New(t)

	accountID := uuid.New().String()

	dynamoClient := utils.Must(storage.NewDynamoDBClient())
	router := httprouter.New()
	h := NewHandler(dynamoClient)
	h.SetupRouter(router)

	client, err := h.repo.Create(
		context.Background(), CreateOptions{
			Secret:    "pa$$word",
			Name:      "Test1",
			AndroidID: uuid.NewString(),
			AccountID: accountID,
		},
	)
	a.NoError(err)

	body := &UpdateClientRequest{Name: "Test2"}
	buf := new(bytes.Buffer)
	a.NoError(json.NewEncoder(buf).Encode(body))

	r := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/clients/%s", client.ID), buf)
	r.Header.Add(account.IDHeader, accountID)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	res := w.Result()

	var response Client
	a.NoError(json.NewDecoder(res.Body).Decode(&response))

	a.Equal(client.ID, response.ID)
	a.Equal(body.Name, response.Name, "Name has been updated")
	a.Equal(client.SecretPrefix, response.SecretPrefix)

	a.Equal(http.StatusOK, res.StatusCode)
}

func TestRegenerateSecret(t *testing.T) {
	a := assert.New(t)

	accountID := uuid.New().String()

	dynamoClient := utils.Must(storage.NewDynamoDBClient())
	router := httprouter.New()
	h := NewHandler(dynamoClient)
	h.SetupRouter(router)

	client, err := h.repo.Create(
		context.Background(), CreateOptions{
			Secret:    "pa$$word",
			Name:      "Test1",
			AndroidID: uuid.NewString(),
			AccountID: accountID,
		},
	)
	a.NoError(err)

	r := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/clients/%s/secret", client.ID), nil)
	r.Header.Add(account.IDHeader, accountID)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	res := w.Result()

	var response RegenerateSecretResponse
	a.NoError(json.NewDecoder(res.Body).Decode(&response))

	updatedClient, err := h.repo.Get(
		context.Background(),
		GetOptions{
			ID:        client.ID,
			AccountID: accountID,
		},
	)
	a.NoError(err)

	a.Equal(client.ID, updatedClient.ID, "Client.ID is unchanged")
	a.Equal(client.Name, updatedClient.Name, "Client.Name is unchanged")

	a.Equal(updatedClient.SecretPrefix, response.Secret[:secretPrefixLength])
	a.NotEqual(client.SecretPrefix, updatedClient.SecretPrefix)

	a.True(utils.Must(argon2id.ComparePasswordAndHash(response.Secret, updatedClient.SecretHash)))
	a.NotEqual(updatedClient.SecretHash, client.SecretHash)
}

func TestRegenerateSecretNotFound(t *testing.T) {
	a := assert.New(t)

	dynamoClient := utils.Must(storage.NewDynamoDBClient())

	router := httprouter.New()
	h := NewHandler(dynamoClient)
	h.SetupRouter(router)

	r := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/clients/%s/secret", uuid.NewString()), nil)
	r.Header.Add(account.IDHeader, uuid.NewString())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	res := w.Result()

	a.Equal(http.StatusNotFound, res.StatusCode)
}

func TestRegenerateSecretUnauthorized(t *testing.T) {
	a := assert.New(t)
	dynamoClient := utils.Must(storage.NewDynamoDBClient())

	router := httprouter.New()
	h := NewHandler(dynamoClient)
	h.SetupRouter(router)

	client, err := h.repo.Create(
		context.Background(), CreateOptions{
			Secret:    "pa$$word",
			Name:      "Test1",
			AndroidID: uuid.NewString(),
			AccountID: uuid.NewString(),
		},
	)
	a.NoError(err)

	r := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/clients/%s/secret", client.ID), nil)
	r.Header.Add(account.IDHeader, uuid.NewString())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	res := w.Result()

	a.Equal(http.StatusNotFound, res.StatusCode, "Client belongs to another AccountID")
}
