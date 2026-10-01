package files_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/files"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func TestDownloadKeepsOldLinksWorking(t *testing.T) {
	e := apitest.NewEcho()
	files.NewHandlers(newService()).RegisterDownload(e)
	client := apitest.New(e)

	rec := client.Get(t, "/download?s=d&id=1")
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "document/rules.pdf")
	assert.Equal(t, "public, max-age=31536000, immutable", rec.Header().Get("Cache-Control"))

	for _, bad := range []string{"/download?s=d&id=0", "/download?s=d&id=x", "/download?s=zz&id=1", "/download"} {
		rec = client.Get(t, bad)
		assert.Equal(t, http.StatusBadRequest, rec.Code, bad)
	}

	rec = client.Get(t, "/download?s=d&id=99")
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
