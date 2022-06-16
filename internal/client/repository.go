package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/KL-Engineering/oauth2-server/internal/core"
	"github.com/alexedwards/argon2id"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
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

type ListOptions struct {
	account_id string
}

func (repo *Repository) List(ctx context.Context, opts ListOptions) ([]Client, error) {
	key := expression.Key("pk").Equal(expression.Value(fmt.Sprintf("Account#%s", opts.account_id)))
	expr, err := expression.NewBuilder().WithKeyCondition(key).Build()
	if err != nil {
		return nil, fmt.Errorf("expression.NewBuilder: %w", err)
	}

	output, err := repo.dynamodb.Query(
		ctx,
		&dynamodb.QueryInput{
			TableName:                 aws.String(tableName),
			KeyConditionExpression:    expr.KeyCondition(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("dynamodb.Query Client: %w", err)
	}

	var clients []Client
	err = attributevalue.UnmarshalListOfMaps(output.Items, &clients)
	if err != nil {
		return nil, fmt.Errorf("dynamodb.UnmarshalMap Client: %w", err)
	}

	return clients, nil
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

type GetOptions struct {
	account_id string
	id         string
}

func (repo *Repository) Get(ctx context.Context, opts GetOptions) (*Client, error) {
	output, err := repo.dynamodb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("Account#%s", opts.account_id)},
			"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("Client#%s", opts.id)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dynamodb.GetItem Client: %w", err)
	}

	if output.Item == nil {
		return nil, core.ErrNotFound
	}

	var client Client
	err = attributevalue.UnmarshalMap(output.Item, &client)
	if err != nil {
		return nil, fmt.Errorf("dynamodb.UnmarshalMap Client: %w", err)
	}

	return &client, nil
}

type DeleteOptions struct {
	account_id string
	id         string
}

func (repo *Repository) Delete(ctx context.Context, opts DeleteOptions) error {
	_, err := repo.dynamodb.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("Account#%s", opts.account_id)},
			"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("Client#%s", opts.id)},
		},
		ConditionExpression: aws.String("attribute_exists(pk)"),
	})
	if err != nil {
		if apiErr := new(types.ConditionalCheckFailedException); errors.As(err, &apiErr) {
			return core.ErrNotFound
		}
		return fmt.Errorf("dynamodb.DeleteItem Client: %w", err)
	}

	return nil
}

type UpdateOptions struct {
	account_id string
	id         string
	name       string
	secret     string
}

func (repo *Repository) Update(ctx context.Context, opts UpdateOptions) (*Client, error) {
	update := expression.UpdateBuilder{}
	if opts.name != "" {
		update = update.Set(expression.Name("name"), expression.Value(opts.name))
	}

	if opts.secret != "" {
		hash, err := argon2id.CreateHash(opts.secret, argon2id.DefaultParams)
		if err != nil {
			return nil, fmt.Errorf("argon2id.CreateHash: %w", err)
		}
		update = update.Set(expression.Name("secret"), expression.Value(hash)).Set(
			expression.Name("secret_prefix"), expression.Value(opts.secret[:secretPrefixLength]),
		)
	}

	expr, err := expression.NewBuilder().WithCondition(
		expression.AttributeExists(expression.Name("pk")),
	).WithUpdate(
		update,
	).Build()

	if err != nil {
		return nil, fmt.Errorf("expression.NewBuilder: %w", err)
	}

	output, err := repo.dynamodb.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("Account#%s", opts.account_id)},
			"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("Client#%s", opts.id)},
		},
		ConditionExpression:       expr.Condition(),
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ReturnValues:              "ALL_NEW",
	})
	if err != nil {
		if apiErr := new(types.ConditionalCheckFailedException); errors.As(err, &apiErr) {
			return nil, core.ErrNotFound
		}
		return nil, fmt.Errorf("dynamodb.UpdateItem Client: %w", err)
	}

	if output.Attributes == nil {
		return nil, core.ErrNotFound
	}

	var client Client
	err = attributevalue.UnmarshalMap(output.Attributes, &client)
	if err != nil {
		return nil, fmt.Errorf("dynamodb.UnmarshalMap Client: %w", err)
	}

	return &client, nil
}
