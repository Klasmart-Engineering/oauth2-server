package client

import (
	"context"
	"testing"

	"github.com/KL-Engineering/oauth2-server/internal/core"
	"github.com/KL-Engineering/oauth2-server/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetByIdNotFound(t *testing.T) {
	a := assert.New(t)

	db, err := storage.NewDynamoDBClient()
	a.NoError(err)
	repo := NewRepository(db)

	client, err := repo.GetByID(context.Background(), uuid.NewString())

	a.Nil(client)
	a.Equal(err, core.ErrNotFound)
}

func TestGetById(t *testing.T) {
	a := assert.New(t)

	db, err := storage.NewDynamoDBClient()
	a.NoError(err)
	repo := NewRepository(db)

	ctx := context.Background()

	client, err := repo.Create(ctx, CreateOptions{
		Secret:    "pa$$word",
		Name:      "Test",
		AndroidID: uuid.NewString(),
		AccountID: uuid.NewString(),
	})
	a.NoError(err)

	got, err := repo.GetByID(ctx, client.ID)
	a.NoError(err)

	a.Equal(client, got)
}
