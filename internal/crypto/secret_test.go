package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSecret(t *testing.T) {
	a := assert.New(t)

	secret, err := GenerateSecret()

	a.Nil(err)

	a.Len(secret, secretLength)
}

func TestGenerateSecretUnique(t *testing.T) {
	a := assert.New(t)

	s1, err := GenerateSecret()
	a.Nil(err)

	s2, err := GenerateSecret()
	a.Nil(err)

	a.NotEqual(s1, s2)
}
