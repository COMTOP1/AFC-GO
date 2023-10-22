package image

import (
	"context"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"
)

type (
	// Store stores the dependencies
	Store struct {
		db *sqlx.DB
	}

	Image struct {
		ID       int         `db:"id" json:"id"`
		Image    null.String `db:"image" json:"image"`
		FileName null.String `db:"file_name" json:"file_name"`
		Caption  null.String `db:"caption" json:"caption"`
	}
)

// NewImageRepo stores our dependency
func NewImageRepo(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetImages(ctx context.Context) ([]Image, error) {
	return s.getImages(ctx)
}

func (s *Store) GetImage(ctx context.Context, i Image) (Image, error) {
	return s.getImage(ctx, i)
}

func (s *Store) AddAffiliation(ctx context.Context, i Image) (Image, error) {
	return s.addImage(ctx, i)
}

func (s *Store) DeleteAffiliation(ctx context.Context, i Image) error {
	return s.deleteImage(ctx, i)
}
