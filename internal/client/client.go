package client

type Client struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	SecretPrefix string `json:"secret_prefix" dynamodbav:"secret_prefix"`
	SecretHash   string `json:"-" dynamodbav:"secret"`
	AndroidID    string `json:"-" dynamodbav:"android_id"`
	AccountID    string `json:"-" dynamodbav:"account_id"`
}
