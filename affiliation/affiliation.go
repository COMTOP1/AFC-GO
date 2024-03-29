package affiliation

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

	Affiliation struct {
		ID       int         `db:"id" json:"id"`
		Name     string      `db:"name" json:"name"`
		Website  null.String `db:"website" json:"website"`
		FileName null.String `db:"file_name" json:"file_name"`
	}
)

// NewAffiliationRepo stores our dependency
func NewAffiliationRepo(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetAffiliations(ctx context.Context) ([]Affiliation, error) {
	return s.getAffiliations(ctx)
}

func (s *Store) GetAffiliationsMinimal(ctx context.Context) ([]Affiliation, error) {
	return s.getAffiliationsMinimal(ctx)
}

func (s *Store) GetAffiliation(ctx context.Context, affiliationParam Affiliation) (Affiliation, error) {
	return s.getAffiliation(ctx, affiliationParam)
}

func (s *Store) AddAffiliation(ctx context.Context, affiliationParam Affiliation) (Affiliation, error) {
	return s.addAffiliation(ctx, affiliationParam)
}

func (s *Store) DeleteAffiliation(ctx context.Context, affiliationParam Affiliation) error {
	return s.deleteAffiliation(ctx, affiliationParam)
}
