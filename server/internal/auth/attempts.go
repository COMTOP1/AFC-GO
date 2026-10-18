package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Failed sign-ins are counted per email address and per IP address over
// attemptWindow. Once either reaches its limit, sign-in is refused until the
// window ends, whether or not the password is right.
const (
	attemptWindow   = 15 * time.Minute
	maxEmailAttempt = 5
	maxIPAttempt    = 20
)

// ErrTooManyAttempts is returned while sign-in is throttled.
var ErrTooManyAttempts = errors.New("too many sign-in attempts; try again in 15 minutes")

func (t *Tokens) emailAttemptKey(email string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email))))
	return t.prefix + "login-fail:email:" + hex.EncodeToString(sum[:])
}

func (t *Tokens) ipAttemptKey(ip string) string { return t.prefix + "login-fail:ip:" + ip }

func (t *Tokens) failures(ctx context.Context, key string) int {
	if t.redis != nil {
		n, err := t.redis.Get(ctx, key).Int()
		if err != nil && !errors.Is(err, redis.Nil) {
			// Fail open: a Redis outage must not lock everyone out.
			slog.ErrorContext(ctx, "failed to read sign-in attempts", "error", err)
		}
		return n
	}
	v, found := t.cache.Get(key)
	if !found {
		return 0
	}
	n, _ := v.(int)
	return n
}

func (t *Tokens) addFailure(ctx context.Context, key string) {
	if t.redis != nil {
		n, err := t.redis.Incr(ctx, key).Result()
		if err == nil && n == 1 {
			err = t.redis.Expire(ctx, key, attemptWindow).Err()
		}
		if err != nil {
			slog.ErrorContext(ctx, "failed to record sign-in attempt", "error", err)
		}
		return
	}
	if t.cache.Add(key, 1, attemptWindow) != nil {
		_, _ = t.cache.IncrementInt(key, 1) //nolint:errcheck // the key exists; Add just failed because of it
	}
}

func (t *Tokens) clearFailures(ctx context.Context, key string) {
	if t.redis != nil {
		_ = t.redis.Del(ctx, key).Err() //nolint:errcheck // best effort; the count expires anyway
		return
	}
	t.cache.Delete(key)
}
