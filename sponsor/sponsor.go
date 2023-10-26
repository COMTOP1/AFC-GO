package sponsor

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

	Sponsor struct {
		ID       int         `db:"id" json:"id"`
		Name     string      `db:"name" json:"name"`
		Website  null.String `db:"website" json:"website"`
		Image    null.String `db:"image" json:"image"`
		FileName null.String `db:"file_name" json:"file_name"`
		Purpose  string      `db:"purpose" json:"purpose"`
		TeamID   int         `db:"team_id" json:"team_id"`
	}
)

// NewSponsorRepo stores our dependency
func NewSponsorRepo(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetSponsors(ctx context.Context) ([]Sponsor, error) {
	return s.getSponsors(ctx)
}

func (s *Store) GetSponsorsIDWebsite(ctx context.Context) ([]Sponsor, error) {
	return s.getSponsorsIDWebsite(ctx)
}

func (s *Store) GetSponsorsTeam(ctx context.Context, teamID string) ([]Sponsor, error) {
	return s.getSponsorsTeam(ctx, teamID)
}

func (s *Store) GetSponsor(ctx context.Context, s1 Sponsor) (Sponsor, error) {
	return s.getSponsor(ctx, s1)
}

func (s *Store) AddSponsor(ctx context.Context, s1 Sponsor) (Sponsor, error) {
	return s.addSponsor(ctx, s1)
}

func (s *Store) EditSponsor(ctx context.Context, s1 Sponsor) (Sponsor, error) {
	return s.editSponsor(ctx, s1)
}

func (s *Store) DeleteSponsor(ctx context.Context, s1 Sponsor) error {
	return s.deleteSponsor(ctx, s1)
}