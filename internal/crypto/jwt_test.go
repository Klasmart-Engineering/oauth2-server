package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const token = "eyJhbGciOiJSUzI1NiIsImtpZCI6IjJjN2VmN2EwLTkxM2YtNDU4ZC04Yzg0LWJlNDRiMzA5MWNiMyIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X2lkIjoiZmU3NTU5YjItY2EzOC00ZDdjLThlYTUtNzZkZGNkMTNiMDQ1IiwiYW5kcm9pZF9pZCI6ImM5OTI3YWQzLThkOTgtNDA1Ni05OThjLTk2NGIxYWNiNDQ5YiIsImF1ZCI6W10sImV4cCI6MTY1NTk4OTY2OSwiaWF0IjoxNjU1OTg4NzY5LCJpc3MiOiJodHRwczovL3BsYXRmb3JtLmtpZHNsb29wLmxpdmUiLCJqdGkiOiJjY2IzY2FiMi1hMjQxLTRiOTYtOGQ2YS1lYjFkNGJlODc3NjQiLCJzY3AiOltdLCJzdWIiOiJiZjBmYmI0Ny01ZjlkLTQ5ZDctYTJiMC1mYzAzMmIwOTEyZjcifQ.k_Z-7vyCd50_bZ4Oozn-XpwuJ1tbOgoWWOTnmaWmFCimKZypv8cYN5SAqzIkCATgOq43eN6nOfoZzCBnPhkaG3YMGpplYc85ZrMezbKQPJqQ_pDCRRS7CcnkyMiOwQiEsM4lnpdzamS_Ef-DKlKCkPDHGsRnXsUTuKYCJlMGr8XW3WkYd-a4KYJUaBYIaY96Gqi10CLPS3wTQIskyvN-PO-w6mINas27NpowOFwXl4GHUKJxycP2Li94pabJbd8BNVZ6LILPXSvOdCoQ8GdOXXKAwwcOq7335EVq7Z9MU5zvLcGzPoSaknrRIr-Gtl3kPlFEBc0WY9Rd0HCOb8jyGw"

func TestPrintJWT(t *testing.T) {
	a := assert.New(t)

	jwt, err := PrintJWT(token)
	a.NoError(err)

	a.Equal(`{
    "alg": "RS256",
    "kid": "2c7ef7a0-913f-458d-8c84-be44b3091cb3",
    "typ": "JWT"
}.{
    "account_id": "fe7559b2-ca38-4d7c-8ea5-76ddcd13b045",
    "android_id": "c9927ad3-8d98-4056-998c-964b1acb449b",
    "aud": [],
    "exp": 1655989669,
    "iat": 1655988769,
    "iss": "https://platform.kidsloop.live",
    "jti": "ccb3cab2-a241-4b96-8d6a-eb1d4be87764",
    "scp": [],
    "sub": "bf0fbb47-5f9d-49d7-a2b0-fc032b0912f7"
}`, jwt)
}

func TestDecodeJWTPayload(t *testing.T) {
	a := assert.New(t)

	payload, err := DecodeJWTPayload(token)
	a.NoError(err)

	a.Equal(map[string]interface{}{
		"account_id": "fe7559b2-ca38-4d7c-8ea5-76ddcd13b045",
		"android_id": "c9927ad3-8d98-4056-998c-964b1acb449b",
		"aud":        []interface{}{},
		"exp":        float64(1655989669),
		"iat":        float64(1655988769),
		"iss":        "https://platform.kidsloop.live",
		"jti":        "ccb3cab2-a241-4b96-8d6a-eb1d4be87764",
		"scp":        []interface{}{},
		"sub":        "bf0fbb47-5f9d-49d7-a2b0-fc032b0912f7",
	}, payload)
}

func TestDecodeJWTHeader(t *testing.T) {
	a := assert.New(t)

	headers, err := DecodeJWTHeader(token)
	a.NoError(err)

	a.Equal(map[string]interface{}{
		"alg": "RS256",
		"kid": "2c7ef7a0-913f-458d-8c84-be44b3091cb3",
		"typ": "JWT",
	}, headers)
}
