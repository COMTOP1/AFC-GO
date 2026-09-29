package document_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth/authtest"
	"github.com/COMTOP1/AFC-GO/server/internal/document"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

var ctx = context.Background()

func rules() document.Document {
	return document.Document{ID: 1, Name: "Club Rules", FileName: "document/rules.pdf"}
}

func newService(rows ...document.Document) (*document.Service, *fakeStore, *uploadtest.Storage) {
	store := newFakeStore(rows...)
	objects := uploadtest.New()
	return document.NewService(store, upload.New(objects)), store, objects
}

func TestCreateValidates(t *testing.T) {
	svc, _, _ := newService()
	_, err := svc.Create(ctx, document.CreateInput{}, uploadtest.File("r.pdf", "application/pdf", "PDF"))
	se, ok := svcerr.As(err)
	require.True(t, ok)
	assert.Contains(t, se.Fields, "name")

	_, err = svc.Create(ctx, document.CreateInput{Name: "x"}, nil)
	se, ok = svcerr.As(err)
	require.True(t, ok)
	assert.Contains(t, se.Fields, "file")
}

func TestCreateAndDelete(t *testing.T) {
	svc, store, objects := newService(rules())
	d, err := svc.Create(ctx, document.CreateInput{Name: "Minutes"}, uploadtest.File("m.pdf", "application/pdf", "PDF"))
	require.NoError(t, err)
	assert.Contains(t, d.FileURL, "https://cdn.test/document/")

	deleted, err := svc.Delete(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "Club Rules", deleted.Name)
	assert.Equal(t, []string{"document/rules.pdf"}, objects.Deleted)
	assert.Len(t, store.rows, 1)
}

func TestDocumentRoutes(t *testing.T) {
	svc, _, _ := newService(rules())
	sessions := authtest.Everyone()
	e := apitest.NewEcho()
	document.NewHandlers(svc).Register(web.NewAPI(e, false), sessions.Guards())
	anon := apitest.New(e)
	editor := anon.As(authtest.Cookie(t, sessions, authtest.Webmaster))
	photographer := anon.As(authtest.Cookie(t, sessions, authtest.Photographer))

	rec := anon.Get(t, "/api/v1/documents")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Len(t, apitest.Decode[[]document.Public](t, rec), 1)

	pdf := apitest.FilePart{Field: "file", Name: "m.pdf", ContentType: "application/pdf", Body: "PDF"}
	assert.Equal(t, http.StatusForbidden, photographer.Multipart(t, http.MethodPost, "/api/v1/documents", map[string]string{"name": "x"}, pdf).Code)
	assert.Equal(t, http.StatusCreated, editor.Multipart(t, http.MethodPost, "/api/v1/documents", map[string]string{"name": "x"}, pdf).Code)
	assert.Equal(t, http.StatusNoContent, editor.JSON(t, http.MethodDelete, "/api/v1/documents/1", nil).Code)
}
