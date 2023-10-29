package document

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type (
	// Store stores the dependencies
	Store struct {
		db *sqlx.DB
	}

	Document struct {
		ID       int    `db:"id" json:"id"`
		Name     string `db:"name" json:"name"`
		FileName string `db:"file_name" json:"file_name"`
	}
)

// NewDocumentRepo stores our dependency
func NewDocumentRepo(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetDocuments(ctx context.Context) ([]Document, error) {
	return s.getDocuments(ctx)
}

func (s *Store) GetDocument(ctx context.Context, d Document) (Document, error) {
	return s.getDocument(ctx, d)
}

func (s *Store) AddDocument(ctx context.Context, d Document) (Document, error) {
	return s.addDocument(ctx, d)
}

func (s *Store) DeleteDocument(ctx context.Context, d Document) error {
	return s.deleteDocument(ctx, d)
}
