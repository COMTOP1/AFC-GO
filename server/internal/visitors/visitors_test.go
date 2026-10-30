package visitors_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/setting"
	"github.com/COMTOP1/AFC-GO/server/internal/visitors"
)

type fakeSettings struct {
	mu    sync.Mutex
	value int
	incs  []int
}

func (f *fakeSettings) GetSetting(_ context.Context, id string) (setting.Setting, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if id != "visitorCount" {
		return setting.Setting{}, errors.New("unexpected id " + id)
	}
	return setting.Setting{ID: id, SettingText: strconv.Itoa(f.value)}, nil
}

func (f *fakeSettings) IncrementSetting(_ context.Context, id string, delta int) (setting.Setting, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.incs = append(f.incs, delta)
	f.value += delta
	return setting.Setting{ID: id, SettingText: strconv.Itoa(f.value)}, nil
}

func TestRecordCountsEachVisitorOnceAndFlushes(t *testing.T) {
	store := &fakeSettings{value: 40}
	c := visitors.New(store, time.Hour, "afcaldermaston.co.uk")
	c.Flush(context.Background()) // seeds from the DB: nothing pending
	assert.Equal(t, 40, c.Count())

	c.Record("1.1.1.1")
	c.Record("1.1.1.1")
	c.Record("2.2.2.2")
	c.Flush(context.Background())

	assert.Equal(t, []int{2}, store.incs)
	assert.Equal(t, 42, c.Count())
}

func TestFlushPicksUpOtherInstances(t *testing.T) {
	store := &fakeSettings{value: 10}
	c := visitors.New(store, time.Hour, "afcaldermaston.co.uk")
	c.Flush(context.Background())
	store.value = 15 // another instance flushed
	c.Flush(context.Background())
	assert.Equal(t, 15, c.Count())
	assert.Empty(t, store.incs)
}

func TestMiddlewareSetsCookieAndRecordsOnce(t *testing.T) {
	store := &fakeSettings{}
	c := visitors.New(store, time.Hour, "afcaldermaston.co.uk")
	e := echo.New()
	e.Use(c.Middleware)
	e.GET("/", func(ctx *echo.Context) error { return ctx.NoContent(http.StatusOK) })

	first := httptest.NewRecorder()
	e.ServeHTTP(first, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil))
	cookies := first.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "afc_aldermaston_visited", cookies[0].Name)

	again := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	again.AddCookie(cookies[0])
	e.ServeHTTP(httptest.NewRecorder(), again)

	c.Flush(context.Background())
	assert.Equal(t, []int{1}, store.incs)
}

func TestMiddlewareSkipsVisitorsWhoOptOut(t *testing.T) {
	store := &fakeSettings{}
	c := visitors.New(store, time.Hour, "afcaldermaston.co.uk")
	e := echo.New()
	e.Use(c.Middleware)
	e.GET("/", func(ctx *echo.Context) error { return ctx.NoContent(http.StatusOK) })

	withCookie := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	withCookie.AddCookie(&http.Cookie{Name: visitors.OptOutCookie, Value: "1", Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	withCookie.RemoteAddr = "192.0.2.1:1234"
	withGPC := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	withGPC.Header.Set("Sec-GPC", "1")
	withGPC.RemoteAddr = "192.0.2.2:1234"

	for _, req := range []*http.Request{withCookie, withGPC} {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Empty(t, rec.Result().Cookies(), "no visit cookie for an opted-out visitor")
	}
	c.Flush(context.Background())
	assert.Empty(t, store.incs)
}
