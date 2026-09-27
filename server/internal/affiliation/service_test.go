package affiliation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

var ctx = context.Background()

func countyFA() affiliation.Affiliation {
	return affiliation.Affiliation{ID: 1, Name: "County FA", Website: null.StringFrom("https://fa.example.test"),
		FileName: null.StringFrom("affiliation/fa.png")}
}

func newService(rows ...affiliation.Affiliation) (*affiliation.Service, *fakeStore, *uploadtest.Storage) {
	store := newFakeStore(rows...)
	objects := uploadtest.New()
	return affiliation.NewService(store, upload.New(objects)), store, objects
}

func TestCreateValidates(t *testing.T) {
	svc, _, objects := newService()
	img := uploadtest.File("a.png", "image/png", "PNG")
	for name, tc := range map[string]struct {
		in    affiliation.CreateInput
		image *upload.File
		field string
	}{
		"name":    {affiliation.CreateInput{}, img, "name"},
		"image":   {affiliation.CreateInput{Name: "x"}, nil, "image"},
		"website": {affiliation.CreateInput{Name: "x", Website: "fa dot com"}, img, "website"},
	} {
		_, err := svc.Create(ctx, tc.in, tc.image)
		se, ok := svcerr.As(err)
		require.True(t, ok, name)
		assert.Contains(t, se.Fields, tc.field, name)
	}
	assert.Empty(t, objects.Objects)
}

func TestCreateAndDelete(t *testing.T) {
	svc, store, objects := newService(countyFA())
	a, err := svc.Create(ctx, affiliation.CreateInput{Name: "League", Website: "https://l.example.test"},
		uploadtest.File("l.png", "image/png", "PNG"))
	require.NoError(t, err)
	assert.NotEmpty(t, a.ImageURL)

	deleted, err := svc.Delete(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "County FA", deleted.Name)
	assert.Equal(t, []string{"affiliation/fa.png"}, objects.Deleted)
	assert.Len(t, store.rows, 1)
}

func TestAffiliationRoutes(t *testing.T) {
	svc, _, _ := newService(countyFA())
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	affiliation.NewHandlers(svc).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	editor := anon.As(authtest.Cookie(t, sessions, authtest.Webmaster))
	manager := anon.As(authtest.Cookie(t, sessions, authtest.Manager))

	rec := anon.Get(t, "/api/v1/affiliations")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Len(t, apitest.Decode[[]affiliation.Public](t, rec), 1)

	logo := apitest.FilePart{Field: "image", Name: "l.png", ContentType: "image/png", Body: "PNG"}
	assert.Equal(t, http.StatusForbidden, manager.Multipart(t, http.MethodPost, "/api/v1/affiliations", map[string]string{"name": "x"}, logo).Code)
	assert.Equal(t, http.StatusCreated, editor.Multipart(t, http.MethodPost, "/api/v1/affiliations", map[string]string{"name": "x"}, logo).Code)
	assert.Equal(t, http.StatusNoContent, editor.JSON(t, http.MethodDelete, "/api/v1/affiliations/1", nil).Code)
	assert.Equal(t, http.StatusNotFound, editor.JSON(t, http.MethodDelete, "/api/v1/affiliations/1", nil).Code)
}
