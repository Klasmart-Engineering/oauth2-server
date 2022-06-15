package client

type Client struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Secret_Prefix string `json:"secret_prefix"`
	Secret_Hash   string `json:"-" dynamodbav:"secret"`
	Android_ID    string `json:"-"`
	Account_ID    string `json:"-"`
}
