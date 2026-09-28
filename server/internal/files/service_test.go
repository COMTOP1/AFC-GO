package files_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/files"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

var ctx = context.Background()

func newService() *files.Service {
	objects := uploadtest.New()
	objects.Objects["document/rules.pdf"] = "PDF"
	objects.Objects["player/adult.png"] = "IMG"
	return files.NewService(upload.New(objects), map[string]files.Resolver{
		"document": func(_ context.Context, id int) (string, error) {
			switch id {
			case 1:
				return "document/rules.pdf", nil
			case 2:
				return "document/missing.pdf", nil // row exists, object doesn't
			case 3:
				return "", nil // row exists with no file
			}
			return "", fmt.Errorf("failed to get document: %w", sql.ErrNoRows)
		},
		// Stands in for player.Service.PhotoKey, which returns NotFound for
		// youth-team and under-18 players (tested in the player package).
		"player": func(_ context.Context, id int) (string, error) {
			if id == 1 {
				return "player/adult.png", nil
			}
			return "", svcerr.NotFound("player photo not found", nil)
		},
	})
}

func kindOf(t *testing.T, err error) svcerr.Kind {
	t.Helper()
	se, ok := svcerr.As(err)
	require.True(t, ok, "want svcerr, got %v", err)
	return se.Kind
}

func TestURL(t *testing.T) {
	svc := newService()
	u, err := svc.URL(ctx, "document", 1)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.test/document/rules.pdf", u)

	for name, tc := range map[string]struct {
		kind string
		id   int
	}{
		"unknown kind":     {"passwords", 1},
		"unknown row":      {"document", 99},
		"object missing":   {"document", 2},
		"row without file": {"document", 3},
	} {
		_, err = svc.URL(ctx, tc.kind, tc.id)
		assert.Equal(t, svcerr.KindNotFound, kindOf(t, err), name)
	}
}

// TestPlayerFileHiddenForMinor pins Review Focus #1 for /files/player.
func TestPlayerFileHiddenForMinor(t *testing.T) {
	e := apitest.NewEcho()
	files.NewHandlers(newService()).Register(web.NewAPI(e, false), authtest.Everyone().Guards())
	c := apitest.New(e)

	rec := c.Get(t, "/api/v1/files/player/1")
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "https://cdn.test/player/adult.png", rec.Header().Get("Location"))
	assert.Equal(t, "public, max-age=31536000, immutable", rec.Header().Get("Cache-Control"))

	rec = c.Get(t, "/api/v1/files/player/2")
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Empty(t, rec.Header().Get("Location"))
}

func TestLegacyKind(t *testing.T) {
	for code, kind := range map[string]string{
		"a": "affiliation", "d": "document", "g": "gallery", "l": "player", "n": "news",
		"p": "programme", "s": "sponsor", "t": "team", "u": "user", "w": "whatson",
	} {
		got, ok := files.LegacyKind(code)
		assert.True(t, ok, code)
		assert.Equal(t, kind, got, code)
	}
	_, ok := files.LegacyKind("x")
	assert.False(t, ok)
}
