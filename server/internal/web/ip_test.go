package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

func TestIPExtractorTrustsOnlyTheProxies(t *testing.T) {
	trusted, err := web.ParseCIDRs(" 10.20.0.0/16 ")
	require.NoError(t, err)
	extract := web.IPExtractor(trusted)

	req := func(remote, xff string) *http.Request {
		r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		r.RemoteAddr = remote
		if xff != "" {
			r.Header.Set("X-Forwarded-For", xff)
		}
		return r
	}
	assert.Equal(t, "203.0.113.7", extract(req("10.20.3.4:5000", "203.0.113.7")), "nginx's header is believed")
	assert.Equal(t, "203.0.113.7", extract(req("10.20.3.4:5000", "198.51.100.1, 203.0.113.7")),
		"a client can't put a fake address before the real one")
	assert.Equal(t, "198.51.100.9", extract(req("198.51.100.9:5000", "203.0.113.7")), "a direct client can't choose its IP")
	assert.Equal(t, "10.99.0.1", extract(req("10.99.0.1:5000", "203.0.113.7")), "other private ranges aren't trusted")

	direct := web.IPExtractor(nil)
	assert.Equal(t, "198.51.100.9", direct(req("198.51.100.9:5000", "203.0.113.7")), "no proxies: header ignored")
}

func TestParseCIDRs(t *testing.T) {
	nets, err := web.ParseCIDRs("10.20.0.0/16, 192.0.2.0/24,")
	require.NoError(t, err)
	assert.Len(t, nets, 2)
	empty, err := web.ParseCIDRs("")
	require.NoError(t, err)
	assert.Empty(t, empty)
	_, err = web.ParseCIDRs("10.20.0.0")
	assert.Error(t, err)
}
