package player

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/team"
)

type (
	// Store stores the dependencies
	Store struct {
		db *sqlx.DB
	}

	Player struct {
		ID          int         `db:"id" json:"id"`
		Name        string      `db:"name" json:"name"`
		FileName    null.String `db:"file_name" json:"file_name"`
		DateOfBirth null.Time   `db:"date_of_birth" json:"date_of_birth"`
		Position    null.String `db:"position" json:"position"`
		IsCaptain   bool        `db:"captain" json:"is_captain"`
		TeamID      int         `db:"team_id" json:"team_id"`
	}
)

// NewPlayerRepo stores our dependency
func NewPlayerRepo(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetPlayers(ctx context.Context) ([]Player, error) {
	return s.getPlayers(ctx)
}

func (s *Store) GetPlayersTeam(ctx context.Context, teamParam team.Team) ([]Player, error) {
	return s.getPlayersTeam(ctx, teamParam)
}

func (s *Store) GetPlayer(ctx context.Context, playerParam Player) (Player, error) {
	return s.getPlayer(ctx, playerParam)
}

func (s *Store) AddPlayer(ctx context.Context, playerParam Player) (Player, error) {
	if playerParam.DateOfBirth.Valid {
		playerParam.DateOfBirth = playerParam.DateOfBirth
	}
	return s.addPlayer(ctx, playerParam)
}

func (s *Store) EditPlayer(ctx context.Context, playerParam Player) (Player, error) {
	return s.editPlayer(ctx, playerParam)
}

func (s *Store) DeletePlayer(ctx context.Context, playerParam Player) error {
	return s.deletePlayer(ctx, playerParam)
}
