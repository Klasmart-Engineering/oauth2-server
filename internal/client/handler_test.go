package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KL-Engineering/oauth2-server/internal/storage"
	"github.com/KL-Engineering/oauth2-server/internal/utils"
	"github.com/alexedwards/argon2id"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

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
