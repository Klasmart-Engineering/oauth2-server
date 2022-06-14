package crypto

import (
	"crypto/rand"
	"math/big"
)

var (
	rander      = rand.Reader
	secretRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_-.~")
)

const secretLength = 40

// Modified version of https://github.com/ory/hydra/blob/79255970787c4793a57fe79d756aa0364b4a9490/x/secret.go#L31
func GenerateSecret() (string, error) {
	l := secretLength
	c := big.NewInt(int64(len(secretRunes)))
	seq := make([]rune, l)

	for i := 0; i < l; i++ {
		r, err := rand.Int(rander, c)
		if err != nil {
			return "", err
		}
		rn := secretRunes[r.Uint64()]
		seq[i] = rn
	}

	return string(seq), nil
}
