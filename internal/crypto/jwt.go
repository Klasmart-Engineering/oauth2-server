package crypto

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/square/go-jose.v2/jwt"
)

type Headers = map[string]interface{}
type Payload = map[string]interface{}

func DecodeJWTPayload(token string) (Payload, error) {
	t, err := jwt.ParseSigned(token)
	if err != nil {
		return nil, fmt.Errorf("token parsing failed: %w", err)
	}

	claims := make(Payload)
	if err := t.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return nil, fmt.Errorf("token decode failed: %w", err)
	}

	return claims, nil
}

func DecodeJWTHeader(token string) (Headers, error) {
	headersB64 := strings.Split(token, ".")[0]
	var headers Headers
	if err := json.NewDecoder(base64.NewDecoder(base64.RawURLEncoding, strings.NewReader(headersB64))).Decode(&headers); err != nil {
		return nil, err
	}

	return headers, nil
}

// Helper to decode and prettify the contents of a JWT
func PrintJWT(token string) (string, error) {
	payload, err := DecodeJWTPayload(token)
	if err != nil {
		return "", err
	}

	payloadJSON, err := json.MarshalIndent(payload, "", "    ")
	if err != nil {
		return "", err
	}

	headers, err := DecodeJWTHeader(token)
	if err != nil {
		return "", err
	}

	headersJSON, err := json.MarshalIndent(headers, "", "    ")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s", string(headersJSON), string(payloadJSON)), nil
}
