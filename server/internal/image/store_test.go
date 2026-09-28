package image_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/image"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestAddImageReturnsID(t *testing.T) {
	db, _ := testdb.Open(t)
	added, err := image.NewImageRepo(db).AddImage(context.Background(), image.Image{FileName: "gallery/x.jpg"})
	require.NoError(t, err)
	require.Positive(t, added.ID)
}
