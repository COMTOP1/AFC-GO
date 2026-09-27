package document_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/document"
	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

func TestAddDocumentReturnsID(t *testing.T) {
	db, _ := testdb.Open(t)
	added, err := document.NewDocumentRepo(db).AddDocument(context.Background(), document.Document{Name: "Policy", FileName: "document/p.pdf"})
	require.NoError(t, err)
	require.Positive(t, added.ID)
}
