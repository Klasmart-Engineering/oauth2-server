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

	account_id := uuid.New().String()

	dynamoClient := utils.Must(storage.NewDynamoDBClient())

	router := httprouter.New()
	h := NewHandler(dynamoClient)
	h.SetupRouter(router)

	r := httptest.NewRequest(http.MethodGet, "/clients", nil)
	r.Header.Add("X-Account-Id", account_id)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	res := w.Result()

	var response ListResponse
	err := json.NewDecoder(res.Body).Decode(&response)
	a.Nil(err)

	a.Equal(ListResponse{
		Records: []Client{},
	}, response)
	a.Equal(http.StatusOK, res.StatusCode)
}

func TestList(t *testing.T) {
	a := assert.New(t)

	account_id := uuid.New().String()

	dynamoClient := utils.Must(storage.NewDynamoDBClient())

	repo := NewRepository(dynamoClient)

	client1, err := repo.Create(context.Background(), CreateOptions{
		secret:     "pa$$word",
		name:       "Test1",
		android_id: uuid.NewString(),
		account_id: account_id,
	})
	a.Nil(err)
	client2, err := repo.Create(context.Background(), CreateOptions{
		secret:     "pa$$word",
		name:       "Test2",
		android_id: uuid.NewString(),
		account_id: account_id,
	})
	a.Nil(err)

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
	r.Header.Add("X-Account-Id", account_id)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	res := w.Result()

	expected, err := json.Marshal(ListResponse{
		Records: clients,
	})
	a.Nil(err)

	actual, err := io.ReadAll(res.Body)
	a.Nil(err)

	a.JSONEq(string(expected), string(actual))
	a.Equal(http.StatusOK, res.StatusCode)
}

func TestCreateNoBody(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest("POST", "/clients/", nil)
	w := httptest.NewRecorder()

	(&Handler{
		repo: *NewRepository(utils.Must(storage.NewDynamoDBClient())),
	}).Create()(w, r, nil)

	res := w.Result()

	a.Equal(res.StatusCode, http.StatusBadRequest)
}

func TestCreateValid(t *testing.T) {
	a := assert.New(t)

	body := &CreateClientRequest{Name: "Test client"}
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(body)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest("POST", "/clients/", buf)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// TODO abstract
	account_id := uuid.New().String()
	r.Header.Add("X-Account-Id", account_id)

	dynamoClient := utils.Must(storage.NewDynamoDBClient())

	(&Handler{
		repo: *NewRepository(dynamoClient),
	}).Create()(w, r, nil)

	res := w.Result()

	var response CreateClientResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	a.Nil(err)

	a.True(utils.IsUUID(response.ID))
	a.Equal(response.Name, body.Name)
	a.Len(response.Secret, 40)

	a.Equal(res.StatusCode, http.StatusCreated)

	output, err := dynamoClient.GetItem(context.Background(), &dynamodb.GetItemInput{Key: map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("Account#%s", account_id)},
		"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("Client#%s", response.ID)},
	},
		TableName: aws.String(tableName),
	})
	a.Nil(err)

	var client Client
	a.Nil(attributevalue.UnmarshalMap(output.Item, &client))

	a.Equal(client.Account_ID, account_id)
	a.True(utils.IsUUID(client.Android_ID))
	a.Equal(client.ID, response.ID)
	a.Equal(client.Name, response.Name)
	a.Equal(client.Secret_Prefix, response.Secret[:secretPrefixLength])

	password_match := utils.Must(argon2id.ComparePasswordAndHash(response.Secret, client.Secret_Hash))
	a.True(password_match)
}

func TestGetNotFound(t *testing.T) {
	a := assert.New(t)

	// TODO abstract
	account_id := uuid.New().String()

	dynamoClient := utils.Must(storage.NewDynamoDBClient())
	repo := NewRepository(dynamoClient)

	id := "non-existent"

	r := httptest.NewRequest("GET", fmt.Sprintf("/clients/%s", id), nil)
	r.Header.Add("X-Account-Id", account_id)
	w := httptest.NewRecorder()

	(&Handler{
		repo: *repo,
	}).Get()(w, r, []httprouter.Param{{Key: "id", Value: id}})

	res := w.Result()

	a.Equal(http.StatusNotFound, res.StatusCode)
}

func TestGetValid(t *testing.T) {
	a := assert.New(t)

	// TODO abstract
	account_id := uuid.New().String()

	dynamoClient := utils.Must(storage.NewDynamoDBClient())
	repo := NewRepository(dynamoClient)

	client := utils.Must(repo.Create(context.Background(), CreateOptions{secret: "pa$$word", name: "Test", android_id: uuid.NewString(), account_id: account_id}))

	r := httptest.NewRequest("GET", fmt.Sprintf("/clients/%s", client.ID), nil)
	r.Header.Add("X-Account-Id", account_id)
	w := httptest.NewRecorder()

	(&Handler{
		repo: *repo,
	}).Get()(w, r, []httprouter.Param{{Key: "id", Value: client.ID}})

	res := w.Result()

	var response Client
	err := json.NewDecoder(res.Body).Decode(&response)
	a.Nil(err)

	a.Equal(response.ID, client.ID)
	a.Equal(response.Name, client.Name)
	a.Equal(response.Secret_Prefix, client.Secret_Prefix)
	a.Equal(response.Secret_Hash, "")

	a.Equal(res.StatusCode, http.StatusOK)
}

func TestDeleteNotFound(t *testing.T) {
	a := assert.New(t)

	account_id := uuid.New().String()

	dynamoClient := utils.Must(storage.NewDynamoDBClient())

	router := httprouter.New()
	h := NewHandler(dynamoClient)
	h.SetupRouter(router)

	r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/clients/%s", uuid.New()), nil)
	r.Header.Add("X-Account-Id", account_id)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	res := w.Result()

	a.Equal(http.StatusNotFound, res.StatusCode)
}

func TestDelete(t *testing.T) {
	a := assert.New(t)

	account_id := uuid.New().String()

	dynamoClient := utils.Must(storage.NewDynamoDBClient())
	repo := NewRepository(dynamoClient)
	router := httprouter.New()
	h := &Handler{
		repo: *repo,
	}
	h.SetupRouter(router)

	client := utils.Must(
		repo.Create(
			context.Background(), CreateOptions{
				secret:     "pa$$word",
				name:       "Test",
				android_id: uuid.NewString(),
				account_id: account_id,
			},
		),
	)

	r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/clients/%s", client.ID), nil)
	r.Header.Add("X-Account-Id", account_id)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	res := w.Result()

	a.Equal(http.StatusNoContent, res.StatusCode, "First DELETE returns NoContent")

	empty_client, err := repo.Get(context.Background(), GetOptions{account_id: client.Account_ID, id: client.ID})
	a.Equal(err, core.ErrNotFound, "Client is deleted")
	a.Nil(empty_client)

	w = httptest.NewRecorder()

	router.ServeHTTP(w, r)

	res = w.Result()

	a.Equal(http.StatusNotFound, res.StatusCode, "Second DELETE returns NotFound")
}

func TestUpdateNotFound(t *testing.T) {
	a := assert.New(t)

	account_id := uuid.New().String()

	dynamoClient := utils.Must(storage.NewDynamoDBClient())

	router := httprouter.New()
	h := NewHandler(dynamoClient)
	h.SetupRouter(router)

	body := &UpdateClientRequest{Name: "Test2"}
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(body)
	a.NoError(err)

	r := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/clients/%s", uuid.New()), buf)
	r.Header.Add("X-Account-Id", account_id)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	res := w.Result()

	a.Equal(http.StatusNotFound, res.StatusCode)
}

func TestUpdate(t *testing.T) {
	a := assert.New(t)

	account_id := uuid.New().String()

	dynamoClient := utils.Must(storage.NewDynamoDBClient())
	repo := NewRepository(dynamoClient)
	router := httprouter.New()
	h := &Handler{
		repo: *repo,
	}
	h.SetupRouter(router)

	client := utils.Must(
		repo.Create(
			context.Background(), CreateOptions{
				secret:     "pa$$word",
				name:       "Test1",
				android_id: uuid.NewString(),
				account_id: account_id,
			},
		),
	)

	body := &UpdateClientRequest{Name: "Test2"}
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(body)
	a.NoError(err)

	r := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/clients/%s", client.ID), buf)
	r.Header.Add("X-Account-Id", account_id)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	res := w.Result()

	var response Client
	err = json.NewDecoder(res.Body).Decode(&response)
	a.NoError(err)

	a.Equal(client.ID, response.ID)
	a.Equal(body.Name, response.Name, "Name has been updated")
	a.Equal(client.Secret_Prefix, response.Secret_Prefix)

	a.Equal(http.StatusOK, res.StatusCode)
}

func TestRegenerateSecret(t *testing.T) {
	a := assert.New(t)

	account_id := uuid.New().String()

	dynamoClient := utils.Must(storage.NewDynamoDBClient())
	repo := NewRepository(dynamoClient)
	router := httprouter.New()
	h := &Handler{
		repo: *repo,
	}
	h.SetupRouter(router)

	client := utils.Must(
		repo.Create(
			context.Background(), CreateOptions{
				secret:     "pa$$word",
				name:       "Test1",
				android_id: uuid.NewString(),
				account_id: account_id,
			},
		),
	)

	r := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/clients/%s/secret", client.ID), nil)
	r.Header.Add("X-Account-Id", account_id)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	res := w.Result()

	var response RegenerateSecretResponse
	err := json.NewDecoder(res.Body).Decode(&response)
	a.NoError(err)

	updated_client, err := repo.Get(context.Background(), GetOptions{id: client.ID, account_id: account_id})
	a.NoError(err)

	a.Equal(client.ID, updated_client.ID, "Client.ID is unchanged")
	a.Equal(client.Name, updated_client.Name, "Client.Name is unchanged")

	a.Equal(updated_client.Secret_Prefix, response.Secret[:secretPrefixLength])
	a.NotEqual(client.Secret_Prefix, updated_client.Secret_Prefix)

	a.True(utils.Must(argon2id.ComparePasswordAndHash(response.Secret, updated_client.Secret_Hash)))
	a.NotEqual(updated_client.Secret_Hash, client.Secret_Hash)
}
