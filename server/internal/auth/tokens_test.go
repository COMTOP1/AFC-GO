package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth"
)

func TestTokensInProcess(t *testing.T) {
	ctx := context.Background()
	tokens := auth.NewTokens(auth.RedisConfig{})
	t.Cleanup(tokens.Close)

	require.NoError(t, tokens.Set(ctx, "abc", 7, time.Hour))
	id, ok := tokens.Get(ctx, "abc")
	assert.True(t, ok)
	assert.Equal(t, 7, id)

	tokens.Delete(ctx, "abc")
	_, ok = tokens.Get(ctx, "abc")
	assert.False(t, ok)

	_, ok = tokens.Get(ctx, "never-issued")
	assert.False(t, ok)
}

func TestTokensExpire(t *testing.T) {
	ctx := context.Background()
	tokens := auth.NewTokens(auth.RedisConfig{})
	require.NoError(t, tokens.Set(ctx, "short", 1, 10*time.Millisecond))
	time.Sleep(30 * time.Millisecond)
	_, ok := tokens.Get(ctx, "short")
	assert.False(t, ok)
}
