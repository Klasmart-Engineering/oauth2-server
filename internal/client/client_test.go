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
		ID:           uuid.NewString(),
		Name:         "Test",
		SecretPrefix: "abc",
		SecretHash:   "abcdef",
		AndroidID:    uuid.NewString(),
		AccountID:    uuid.NewString(),
	}

	bytes, err := json.Marshal(client)
	a.NoError(err)

	expected := utils.Must(json.Marshal(map[string]interface{}{
		"id":            client.ID,
		"name":          client.Name,
		"secret_prefix": client.SecretPrefix,
	}))
	a.JSONEq(string(bytes), string(expected), "Does not include 'secret'")
}
