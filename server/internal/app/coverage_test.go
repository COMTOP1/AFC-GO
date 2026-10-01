package app_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEveryLegacyActionHasAnAPIRoute is the spec's cleanup checklist: every
// action the classic template site offered has an /api/v1 equivalent (the two
// form-redirect helpers, programmeselect and whatsonselect, are replaced by
// query parameters and intentionally have none).
func TestEveryLegacyActionHasAnAPIRoute(t *testing.T) {
	want := []string{
		"GET /api/v1/site", "GET /api/v1/home", "GET /api/v1/contact",
		"POST /api/v1/auth/login", "POST /api/v1/auth/logout", "GET /api/v1/auth/me",
		"POST /api/v1/auth/password", "GET /api/v1/auth/reset/:token", "POST /api/v1/auth/reset/:token",
		"GET /api/v1/account", "PUT /api/v1/account/image", "DELETE /api/v1/account/image",
		"GET /api/v1/news", "POST /api/v1/news", "GET /api/v1/news/:id", "PATCH /api/v1/news/:id", "DELETE /api/v1/news/:id",
		"GET /api/v1/whatson", "POST /api/v1/whatson", "GET /api/v1/whatson/:id", "PATCH /api/v1/whatson/:id", "DELETE /api/v1/whatson/:id",
		"GET /api/v1/teams", "POST /api/v1/teams", "GET /api/v1/teams/:id", "PATCH /api/v1/teams/:id", "DELETE /api/v1/teams/:id",
		"GET /api/v1/players", "POST /api/v1/players", "PATCH /api/v1/players/:id", "DELETE /api/v1/players/:id",
		"GET /api/v1/programmes", "POST /api/v1/programmes", "DELETE /api/v1/programmes/:id",
		"GET /api/v1/seasons", "POST /api/v1/seasons", "PATCH /api/v1/seasons/:id", "DELETE /api/v1/seasons/:id",
		"GET /api/v1/sponsors", "POST /api/v1/sponsors", "DELETE /api/v1/sponsors/:id",
		"GET /api/v1/affiliations", "POST /api/v1/affiliations", "DELETE /api/v1/affiliations/:id",
		"GET /api/v1/documents", "POST /api/v1/documents", "DELETE /api/v1/documents/:id",
		"GET /api/v1/gallery", "POST /api/v1/gallery", "DELETE /api/v1/gallery/:id",
		"GET /api/v1/info", "PUT /api/v1/info",
		"GET /api/v1/users", "POST /api/v1/users", "GET /api/v1/users/:id", "PATCH /api/v1/users/:id",
		"DELETE /api/v1/users/:id", "POST /api/v1/users/:id/reset",
		"PUT /api/v1/settings/display-email",
		"GET /api/v1/files/:kind/:id",
		"GET /api/v1/health", "GET /api/health",
	}
	have := map[string]bool{}
	for _, r := range bareApp(t).Echo.Routes() {
		have[r.Method+" "+r.Path] = true
	}
	for _, route := range want {
		assert.True(t, have[route], "missing %s", route)
	}
}
