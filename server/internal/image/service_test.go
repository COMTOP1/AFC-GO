package image_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/image"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

var ctx = context.Background()

func matchDay() image.Image {
	return image.Image{ID: 1, FileName: "gallery/one.jpg", Caption: null.StringFrom("Match day")}
}

func newService(rows ...image.Image) (*image.Service, *fakeStore, *uploadtest.Storage) {
	store := newFakeStore(rows...)
	objects := uploadtest.New()
	return image.NewService(store, upload.New(objects)), store, objects
}

func TestCreateRequiresImage(t *testing.T) {
	svc, _, _ := newService()
	_, err := svc.Create(ctx, image.CreateInput{Caption: "x"}, nil)
	se, ok := svcerr.As(err)
	require.True(t, ok)
	assert.Contains(t, se.Fields, "image")
}

func TestCreateAndDelete(t *testing.T) {
	svc, store, objects := newService(matchDay())
	p, err := svc.Create(ctx, image.CreateInput{}, uploadtest.File("p.jpg", "image/jpeg", "JPG"))
	require.NoError(t, err)
	assert.Empty(t, p.Caption, "caption is optional")
	assert.Contains(t, p.ImageURL, "https://cdn.test/gallery/")

	_, err = svc.Delete(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, []string{"gallery/one.jpg"}, objects.Deleted)
	assert.Len(t, store.rows, 1)
}

func TestGalleryRoutesAllowPhotographers(t *testing.T) {
	svc, _, _ := newService(matchDay())
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	image.NewHandlers(svc).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	photographer := anon.As(authtest.Cookie(t, sessions, authtest.Photographer))
	manager := anon.As(authtest.Cookie(t, sessions, authtest.Manager))

	rec := anon.Get(t, "/api/v1/gallery")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Match day", apitest.Decode[[]image.Public](t, rec)[0].Caption)

	jpg := apitest.FilePart{Field: "image", Name: "p.jpg", ContentType: "image/jpeg", Body: "JPG"}
	assert.Equal(t, http.StatusForbidden, manager.Multipart(t, http.MethodPost, "/api/v1/gallery", nil, jpg).Code)
	assert.Equal(t, http.StatusCreated, photographer.Multipart(t, http.MethodPost, "/api/v1/gallery", map[string]string{"caption": "Goal"}, jpg).Code)
	assert.Equal(t, http.StatusForbidden, manager.JSON(t, http.MethodDelete, "/api/v1/gallery/1", nil).Code)
	assert.Equal(t, http.StatusNoContent, photographer.JSON(t, http.MethodDelete, "/api/v1/gallery/1", nil).Code)
}
