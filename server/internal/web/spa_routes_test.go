package web_test

import (
	"os"
	"regexp"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// TestSPARoutesMatchTheClient keeps the server's route list (used to send a
// real 404 for unknown pages) in step with the React router.
func TestSPARoutesMatchTheClient(t *testing.T) {
	src, err := os.ReadFile("../../../client/App.tsx")
	require.NoError(t, err)

	client := []string{}
	for _, m := range regexp.MustCompile(`<Route\s+index\b`).FindAllString(string(src), -1) {
		_ = m
		client = append(client, "/")
	}
	for _, m := range regexp.MustCompile(`path="([^"]+)"`).FindAllStringSubmatch(string(src), -1) {
		if m[1] == "*" {
			continue
		}
		client = append(client, "/"+m[1])
	}
	require.NotEmpty(t, client)

	server := append([]string(nil), web.SPARoutes...)
	sort.Strings(client)
	sort.Strings(server)
	assert.Equal(t, client, server, "web.SPARoutes must list exactly the routes in client/App.tsx")
}
