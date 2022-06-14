package client

import (
	"context"
	"fmt"

	"github.com/alexedwards/argon2id"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

const (
	tableName          = "authentication"
	secretPrefixLength = 3
)

type Repository struct {
	dynamodb *dynamodb.Client
}

func NewRepository(dynamodbClient *dynamodb.Client) *Repository {
	return &Repository{
		dynamodb: dynamodbClient,
	}
}

type CreateOptions struct {
	secret     string
	name       string
	android_id string
	account_id string
}

func (repo *Repository) Create(ctx context.Context, opts CreateOptions) (*Client, error) {
	hash, err := argon2id.CreateHash(opts.secret, argon2id.DefaultParams)
	if err != nil {
		return nil, fmt.Errorf("argon2id.CreateHash: %w", err)
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("uuid.NewRandom: %w", err)
	}

	client := Client{
		ID:            id.String(),
		Secret_Prefix: opts.secret[:secretPrefixLength],
		Secret_Hash:   hash,
		Name:          opts.name,
		Android_ID:    opts.android_id,
		Account_ID:    opts.account_id,
	}

	input := dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			"pk":            &types.AttributeValueMemberS{Value: fmt.Sprintf("Account#%s", client.Account_ID)},
			"sk":            &types.AttributeValueMemberS{Value: fmt.Sprintf("Client#%s", client.ID)},
			"id":            &types.AttributeValueMemberS{Value: client.ID},
			"secret":        &types.AttributeValueMemberS{Value: client.Secret_Hash},
			"secret_prefix": &types.AttributeValueMemberS{Value: client.Secret_Prefix},
			"name":          &types.AttributeValueMemberS{Value: client.Name},
			"android_id":    &types.AttributeValueMemberS{Value: client.Android_ID},
			"account_id":    &types.AttributeValueMemberS{Value: client.Account_ID},
		},
		ConditionExpression: aws.String("attribute_not_exists(pk)"),
	}

	_, err = repo.dynamodb.PutItem(ctx, &input)
	if err != nil {
		return nil, fmt.Errorf("dynamodb.PutItem Client: %w", err)
	}

	return &client, nil
}
