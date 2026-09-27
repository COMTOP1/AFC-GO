package whatson_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

func newHarness(t *testing.T) (anon, editor, manager *apitest.Client, store *fakeStore) {
	t.Helper()
	svc, store, _ := newService(upcoming(), finished())
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	whatson.NewHandlers(svc).Register(web.NewAPI(e, false), sessions.Guards())
	c := apitest.New(e)
	return c, c.As(authtest.Cookie(t, sessions, authtest.Webmaster)), c.As(authtest.Cookie(t, sessions, authtest.Manager)), store
}

func TestListPeriodQuery(t *testing.T) {
	anon, _, _, _ := newHarness(t)
	rec := anon.Get(t, "/api/v1/whatson?period=future")
	require.Equal(t, http.StatusOK, rec.Code)
	got := apitest.Decode[[]whatson.Event](t, rec)
	require.Len(t, got, 1)
	assert.Equal(t, 1, got[0].ID)

	assert.Equal(t, http.StatusUnprocessableEntity, anon.Get(t, "/api/v1/whatson?period=soon").Code)
	assert.Equal(t, http.StatusOK, anon.Get(t, "/api/v1/whatson/2").Code)
	assert.Equal(t, http.StatusNotFound, anon.Get(t, "/api/v1/whatson/99").Code)
}

func TestWhatsOnWritesRequireEditor(t *testing.T) {
	anon, _, manager, _ := newHarness(t)
	fields := map[string]string{"title": "x", "dateOfEvent": "2030-01-01"}
	for _, tc := range []struct{ method, path string }{
		{http.MethodPost, "/api/v1/whatson"},
		{http.MethodPatch, "/api/v1/whatson/1"},
		{http.MethodDelete, "/api/v1/whatson/1"},
	} {
		assert.Equal(t, http.StatusUnauthorized, anon.Multipart(t, tc.method, tc.path, fields).Code)
		assert.Equal(t, http.StatusForbidden, manager.Multipart(t, tc.method, tc.path, fields).Code)
	}
}

func TestCreateAndPatch(t *testing.T) {
	_, editor, _, store := newHarness(t)
	rec := editor.Multipart(t, http.MethodPost, "/api/v1/whatson",
		map[string]string{"title": "Quiz", "content": "<p>7pm</p>", "dateOfEvent": "2030-02-03"})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	created := apitest.Decode[whatson.Event](t, rec)
	assert.Equal(t, "2030-02-03", created.DateOfEvent.Format(web.DateLayout))

	rec = editor.Multipart(t, http.MethodPost, "/api/v1/whatson", map[string]string{"title": "No date"})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, apitest.ErrorOf(t, rec).Fields, "dateOfEvent")

	rec = editor.Multipart(t, http.MethodPatch, "/api/v1/whatson/1", map[string]string{"dateOfEvent": "03/02/2030"})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "dates must be YYYY-MM-DD")
	assert.Equal(t, "Presentation night", store.row(1).Title)
}
