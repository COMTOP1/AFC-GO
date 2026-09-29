package upload_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
)

func TestSaveStoresUnderCategoryWithExtension(t *testing.T) {
	store := uploadtest.New()
	files := upload.New(store)

	key, err := files.Save(context.Background(), uploadtest.File("photo.png", "image/png", "PNGDATA"), "news")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(key, "news/"), key)
	assert.True(t, strings.HasSuffix(key, ".png"), key)
	assert.Equal(t, "PNGDATA", store.Objects[key])
	assert.Equal(t, "https://cdn.test/"+key, files.URL(key))
}

func TestSaveRejectsUnsupportedType(t *testing.T) {
	for _, ct := range []string{"text/html", "application/x-msdownload", "application/octet-stream", ""} {
		t.Run(ct, func(t *testing.T) {
			store := uploadtest.New()
			_, err := upload.New(store).Save(context.Background(), uploadtest.File("x", ct, "<script>"), "news")
			se, ok := svcerr.As(err)
			require.True(t, ok, "want svcerr, got %v", err)
			assert.Equal(t, svcerr.KindInvalid, se.Kind)
			assert.Contains(t, se.Fields, "file")
			assert.Empty(t, store.Objects)
		})
	}
}

func TestRemoveIgnoresEmptyKeyAndLogsFailures(t *testing.T) {
	store := uploadtest.New()
	files := upload.New(store)
	files.Remove(context.Background(), "")
	assert.Empty(t, store.Deleted)
	files.Remove(context.Background(), "news/a.png")
	assert.Equal(t, []string{"news/a.png"}, store.Deleted)
	assert.Empty(t, files.URL(""))
}
