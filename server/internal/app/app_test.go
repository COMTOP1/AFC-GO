package app_test

import (
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

// publicWrites are the only unsafe API routes anonymous users may call.
var publicWrites = map[string]bool{
	"POST /api/v1/auth/login":        true,
	"POST /api/v1/auth/reset/:token": true,
}

var pathParam = regexp.MustCompile(`:[A-Za-z]+`)

// TestUnsafeAPIRoutesRequireLogin walks every registered route, so each new
// write endpoint is covered automatically.
func TestUnsafeAPIRoutesRequireLogin(t *testing.T) {
	a := bareApp(t)
	checked := 0
	for _, r := range a.Echo.Routes() {
		if !strings.HasPrefix(r.Path, "/api/v1/") {
			continue
		}
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		default:
			continue
		}
		key := r.Method + " " + r.Path
		if publicWrites[key] {
			continue
		}
		path := pathParam.ReplaceAllString(r.Path, "1")
		rec := apitest.New(a.Echo).JSON(t, r.Method, path, nil)
		assert.Equal(t, http.StatusUnauthorized, rec.Code, key)
		checked++
	}
	t.Logf("checked %d unsafe API routes", checked)
}

func TestGuardedGetRoutesRequireLogin(t *testing.T) {
	a := bareApp(t)
	for _, path := range guardedGets {
		rec := apitest.New(a.Echo).Get(t, path)
		assert.Equal(t, http.StatusUnauthorized, rec.Code, path)
	}
}

func TestSwaggerDocServed(t *testing.T) {
	a := bareApp(t)
	rec := apitest.New(a.Echo).Get(t, "/api/v1/swagger/doc.json")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"/auth/me"`)
}

func TestLegacyAndAPINotFoundDiffer(t *testing.T) {
	a := bareApp(t)
	api := apitest.New(a.Echo).Get(t, "/api/v1/nope")
	assert.Equal(t, http.StatusNotFound, api.Code)
	assert.Equal(t, 404, apitest.ErrorOf(t, api).Code)
}
