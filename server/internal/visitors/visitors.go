// Package visitors counts unique daily visitors and persists the running
// total in the settings table, shared across instances.
package visitors

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"go.opentelemetry.io/otel"

	"github.com/COMTOP1/AFC-GO/server/internal/setting"
)

var tracer = otel.Tracer("github.com/COMTOP1/AFC-GO/server/internal/visitors")

const (
	settingID  = "visitorCount"
	cookieName = "afc_aldermaston_visited"
)

// SettingStore persists the running total (satisfied by *setting.Store).
type SettingStore interface {
	GetSetting(ctx context.Context, settingID string) (setting.Setting, error)
	IncrementSetting(ctx context.Context, settingID string, delta int) (setting.Setting, error)
}

// Counter records visits locally and flushes them to the database.
type Counter struct {
	store        SettingStore
	seen         *cache.Cache
	interval     time.Duration
	cookieDomain string

	mu      sync.Mutex
	pending int
	total   int

	stop     chan struct{}
	stopOnce sync.Once
}

func New(store SettingStore, interval time.Duration, cookieDomain string) *Counter {
	return &Counter{
		store:        store,
		seen:         cache.New(time.Hour, time.Hour),
		interval:     interval,
		cookieDomain: cookieDomain,
		stop:         make(chan struct{}),
	}
}

// Start seeds the total from the database and flushes every interval until Stop.
func (c *Counter) Start(ctx context.Context) {
	c.Flush(ctx)
	go func() { //nolint:gosec,contextcheck // background flusher outlives any request; context.Background() is intentional
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.Flush(context.Background())
			case <-c.stop:
				return
			}
		}
	}()
}

// Stop ends the flusher. It is safe to call more than once.
func (c *Counter) Stop() {
	c.stopOnce.Do(func() { close(c.stop) })
}

// Record counts visitorID once per hour-long cache window.
func (c *Counter) Record(visitorID string) {
	if _, found := c.seen.Get(visitorID); found {
		return
	}
	c.seen.Set(visitorID, true, cache.DefaultExpiration)
	c.mu.Lock()
	c.pending++
	c.mu.Unlock()
}

// Count is the last known total across all instances.
func (c *Counter) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.total
}

// Flush adds pending visits to the database total, or, with nothing pending,
// re-reads the total so this instance sees other instances' visits.
func (c *Counter) Flush(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "visitors.Flush")
	defer span.End()

	c.mu.Lock()
	delta := c.pending
	c.pending = 0
	c.mu.Unlock()

	var (
		s   setting.Setting
		err error
	)
	if delta == 0 {
		s, err = c.store.GetSetting(ctx, settingID)
		if err != nil {
			return // not created yet; it will be on the first increment
		}
	} else {
		s, err = c.store.IncrementSetting(ctx, settingID, delta)
		if err != nil {
			span.RecordError(err)
			slog.InfoContext(ctx, fmt.Sprintf("failed to increment visitor count: %+v", err))
			c.mu.Lock()
			c.pending += delta // retry on the next flush
			c.mu.Unlock()
			return
		}
	}

	total, err := strconv.Atoi(s.SettingText)
	if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, fmt.Sprintf("failed to parse visitor count: %+v", err))
		return
	}
	c.mu.Lock()
	c.total = total
	c.mu.Unlock()
}

// Middleware counts a visit the first time a browser is seen in 24 hours.
func (c *Counter) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if _, err := ctx.Cookie(cookieName); errors.Is(err, http.ErrNoCookie) {
			ctx.SetCookie(&http.Cookie{
				Name:     cookieName,
				Value:    "visited",
				Expires:  time.Now().Add(24 * time.Hour),
				Domain:   c.cookieDomain,
				Path:     "/",
				SameSite: http.SameSiteStrictMode,
				Secure:   true,
				HttpOnly: true,
			})
			c.Record(ctx.RealIP())
		}
		return next(ctx)
	}
}
