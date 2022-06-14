package client

import (
	"encoding/json"
	"testing"

	"github.com/KL-Engineering/oauth2-server/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestJSONMarshal(t *testing.T) {
	a := assert.New(t)

	client := Client{
		ID:            uuid.NewString(),
		Name:          "Test",
		Secret_Prefix: "abc",
		Secret_Hash:   "abcdef",
		Android_ID:    uuid.NewString(),
		Account_ID:    uuid.NewString(),
	}

	bytes, err := json.Marshal(client)
	a.Nil(err)

	expected := utils.Must(json.Marshal(map[string]interface{}{
		"id":            client.ID,
		"name":          client.Name,
		"secret_prefix": client.Secret_Prefix,
		"android_id":    client.Android_ID,
		"account_id":    client.Account_ID,
	}))
	a.JSONEq(string(bytes), string(expected), "Does not include 'secret'")
}
