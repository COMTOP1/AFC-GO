package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenKeyNeverHoldsTheToken(t *testing.T) {
	tokens := &Tokens{prefix: "afc:"}
	token := "0b6f3c1e-8f0a-4c7d-9a43-5e2b1d7c9f10" //nolint:gosec // a made-up test token
	key := tokens.key(token)
	assert.NotContains(t, key, token)
	assert.True(t, strings.HasPrefix(key, "afc:reset-token:"))
	assert.Len(t, strings.TrimPrefix(key, "afc:reset-token:"), 64, "a hex SHA-256")
	assert.Equal(t, key, tokens.key(token), "the same token always maps to the same key")
	assert.NotEqual(t, key, tokens.key(token+"x"))
}
