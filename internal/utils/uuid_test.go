package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUUID(t *testing.T) {
	assert.True(t, IsUUID("bec025ea-4dc0-4239-b2ac-39bb8450003c"))
}

func TestIsNotUUID(t *testing.T) {
	assert.False(t, IsUUID("abcde"))
}
