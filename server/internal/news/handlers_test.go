package news_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

type harness struct {
	anon, editor, manager *apitest.Client
	store                 *fakeStore
}

func newHarness(t *testing.T) harness {
	t.Helper()
	svc, store, _ := newService(seeded())
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	api := web.NewAPI(e, false)
	news.NewHandlers(svc).Register(api, sessions.Guards())
	c := apitest.New(e)
	return harness{
		anon:    c,
		editor:  c.As(authtest.Cookie(t, sessions, authtest.Webmaster)),
		manager: c.As(authtest.Cookie(t, sessions, authtest.Manager)),
		store:   store,
	}
}

func TestListAndGet(t *testing.T) {
	h := newHarness(t)
	rec := h.anon.Get(t, "/api/v1/news")
	require.Equal(t, http.StatusOK, rec.Code)
	list := apitest.Decode[[]news.Article](t, rec)
	require.Len(t, list, 1)
	assert.Equal(t, "https://cdn.test/news/old.jpg", list[0].ImageURL)

	rec = h.anon.Get(t, "/api/v1/news/1")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Opener", apitest.Decode[news.Article](t, rec).Title)

	rec = h.anon.Get(t, "/api/v1/news/999")
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestWritesRequireEditor(t *testing.T) {
	h := newHarness(t)
	cases := []struct {
		method, path string
	}{
		{http.MethodPost, "/api/v1/news"},
		{http.MethodPatch, "/api/v1/news/1"},
		{http.MethodDelete, "/api/v1/news/1"},
	}
	for _, tc := range cases {
		assert.Equal(t, http.StatusUnauthorized, h.anon.Multipart(t, tc.method, tc.path, map[string]string{"title": "x"}).Code, "anon %s %s", tc.method, tc.path)
		assert.Equal(t, http.StatusForbidden, h.manager.Multipart(t, tc.method, tc.path, map[string]string{"title": "x"}).Code, "manager %s %s", tc.method, tc.path)
	}
	assert.Equal(t, "Opener", h.store.row(1).Title, "nothing changed")
}

func TestCreate(t *testing.T) {
	h := newHarness(t)
	rec := h.editor.Multipart(t, http.MethodPost, "/api/v1/news",
		map[string]string{"title": "Cup run", "content": "<p>on</p>"},
		apitest.FilePart{Field: "image", Name: "c.jpg", ContentType: "image/jpeg", Body: "JPG"})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	a := apitest.Decode[news.Article](t, rec)
	assert.Positive(t, a.ID)
	assert.NotEmpty(t, a.ImageURL)

	rec = h.anon.Get(t, fmt.Sprintf("/api/v1/news/%d", a.ID))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCreateRejectsHTMLUploadOverHTTP(t *testing.T) {
	h := newHarness(t)
	rec := h.editor.Multipart(t, http.MethodPost, "/api/v1/news",
		map[string]string{"title": "x"},
		apitest.FilePart{Field: "image", Name: "x.html", ContentType: "text/html", Body: "<script>"})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, apitest.ErrorOf(t, rec).Fields, "file")
}

func TestPatchIsPartial(t *testing.T) {
	h := newHarness(t)
	rec := h.editor.Multipart(t, http.MethodPatch, "/api/v1/news/1", map[string]string{"title": "Only title"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	row := h.store.row(1)
	assert.Equal(t, "Only title", row.Title)
	assert.Equal(t, "<p>hi</p>", row.Content.String)
	assert.Equal(t, "news/old.jpg", row.FileName.String)

	rec = h.editor.Multipart(t, http.MethodPatch, "/api/v1/news/1", map[string]string{"removeImage": "true"})
	require.Equal(t, http.StatusOK, rec.Code)
	assert.False(t, h.store.row(1).FileName.Valid)
}

func TestDelete(t *testing.T) {
	h := newHarness(t)
	rec := h.editor.JSON(t, http.MethodDelete, "/api/v1/news/1", nil)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, http.StatusNotFound, h.anon.Get(t, "/api/v1/news/1").Code)
	assert.Equal(t, http.StatusNotFound, h.editor.JSON(t, http.MethodDelete, "/api/v1/news/1", nil).Code)
}
