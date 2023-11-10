package sponsor

import (
	"context"
	"github.com/COMTOP1/AFC-GO/team"

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
		FileName null.String `db:"file_name" json:"file_name"`
		Purpose  string      `db:"purpose" json:"purpose"`
		TeamID   string      `db:"team_id" json:"team_id"`
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

func (s *Store) GetSponsorsMinimal(ctx context.Context) ([]Sponsor, error) {
	return s.getSponsorsMinimal(ctx)
}

func (s *Store) GetSponsorsTeam(ctx context.Context, teamParam team.Team) ([]Sponsor, error) {
	return s.getSponsorsTeam(ctx, teamParam)
}

func (s *Store) GetSponsor(ctx context.Context, sponsorParam Sponsor) (Sponsor, error) {
	return s.getSponsor(ctx, sponsorParam)
}

func (s *Store) AddSponsor(ctx context.Context, sponsorParam Sponsor) (Sponsor, error) {
	return s.addSponsor(ctx, sponsorParam)
}

func (s *Store) EditSponsor(ctx context.Context, sponsorParam Sponsor) (Sponsor, error) {
	return s.editSponsor(ctx, sponsorParam)
}

func (s *Store) DeleteSponsor(ctx context.Context, sponsorParam Sponsor) error {
	return s.deleteSponsor(ctx, sponsorParam)
}
