package team

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

	Team struct {
		ID          int         `db:"id" json:"id"`
		Name        string      `db:"name" json:"name"`
		League      string      `db:"league" json:"league"`
		Division    string      `db:"division" json:"division"`
		LeagueTable null.String `db:"league_table" json:"league_table"`
		Fixtures    null.String `db:"fixtures" json:"fixtures"`
		Coach       string      `db:"coach" json:"coach"`
		Physio      null.String `db:"physio" json:"physio"`
		Image       null.String `db:"team_photo" json:"image"`
		FileName    null.String `db:"file_name" json:"file_name"`
		IsActive    bool        `db:"active" json:"is_active"`
		IsYouth     bool        `db:"youth" json:"is_youth"`
		Ages        int         `db:"ages" json:"ages"`
	}
)

// NewTeamRepo stores our dependency
func NewTeamRepo(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetTeams(ctx context.Context) ([]Team, error) {
	return s.getTeams(ctx)
}

func (s *Store) GetTeamsActive(ctx context.Context) ([]Team, error) {
	return s.getTeamsActive(ctx)
}

func (s *Store) GetTeam(ctx context.Context, t Team) (Team, error) {
	return s.getTeam(ctx, t)
}

func (s *Store) AddTeam(ctx context.Context, t Team) (Team, error) {
	return s.addTeam(ctx, t)
}

func (s *Store) EditTeam(ctx context.Context, t Team) (Team, error) {
	return s.editTeam(ctx, t)
}

func (s *Store) DeleteTeam(ctx context.Context, t Team) error {
	return s.deleteTeam(ctx, t)
}
