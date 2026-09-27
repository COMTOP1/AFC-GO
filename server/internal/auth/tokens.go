package auth

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("github.com/COMTOP1/AFC-GO/server/internal/auth")

// RedisConfig configures the optional Redis/Valkey used to share reset
// tokens across instances. With no Addresses an in-process cache is used,
// which only works for a single instance.
type RedisConfig struct {
	Addresses  []string
	MasterName string
	Username   string
	Password   string
	DB         int
	TLS        bool
	// KeyPrefix namespaces keys per environment, e.g. "afc:prod:". Defaults to "afc:".
	KeyPrefix string
}

// Tokens stores one-time password reset tokens.
type Tokens struct {
	redis  redis.UniversalClient
	cache  *cache.Cache
	prefix string
}

func NewTokens(conf RedisConfig) *Tokens {
	t := &Tokens{cache: cache.New(time.Hour, time.Hour), prefix: conf.KeyPrefix}
	if t.prefix == "" {
		t.prefix = "afc:"
	}
	if len(conf.Addresses) > 0 {
		opts := &redis.UniversalOptions{
			Addrs:      conf.Addresses,
			MasterName: conf.MasterName,
			Username:   conf.Username,
			Password:   conf.Password,
			DB:         conf.DB,
		}
		if conf.TLS {
			opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		}
		t.redis = redis.NewUniversalClient(opts)
		slog.Info("using redis/valkey for shared cache")
	} else {
		slog.Info("no redis addresses configured, using in-process cache")
	}
	return t
}

func (t *Tokens) key(token string) string { return t.prefix + "reset-token:" + token }

// Set maps token to userID for ttl.
func (t *Tokens) Set(ctx context.Context, token string, userID int, ttl time.Duration) error {
	ctx, span := tracer.Start(ctx, "auth.Tokens.Set")
	defer span.End()
	if t.redis != nil {
		if err := t.redis.Set(ctx, t.key(token), userID, ttl).Err(); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to store reset token: %w", err)
		}
		return nil
	}
	t.cache.Set(token, userID, ttl)
	return nil
}

// Get returns the user ID a token was issued for.
func (t *Tokens) Get(ctx context.Context, token string) (int, bool) {
	ctx, span := tracer.Start(ctx, "auth.Tokens.Get")
	defer span.End()
	if t.redis != nil {
		id, err := t.redis.Get(ctx, t.key(token)).Int()
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				span.RecordError(err)
			}
			return 0, false
		}
		return id, true
	}
	v, found := t.cache.Get(token)
	if !found {
		return 0, false
	}
	id, ok := v.(int)
	return id, ok
}

// Delete invalidates a token after use.
func (t *Tokens) Delete(ctx context.Context, token string) {
	ctx, span := tracer.Start(ctx, "auth.Tokens.Delete")
	defer span.End()
	if t.redis != nil {
		if err := t.redis.Del(ctx, t.key(token)).Err(); err != nil {
			span.RecordError(err)
			slog.ErrorContext(ctx, fmt.Sprintf("failed to delete reset token from redis: %+v", err))
		}
		return
	}
	t.cache.Delete(token)
}

// Close releases the Redis client, if any.
func (t *Tokens) Close() {
	if t.redis == nil {
		return
	}
	if err := t.redis.Close(); err != nil {
		slog.Error(fmt.Sprintf("failed to close redis client: %+v", err))
	}
}
