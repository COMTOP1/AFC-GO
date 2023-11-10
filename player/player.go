package player

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

	Player struct {
		ID          int         `db:"id" json:"id"`
		Name        string      `db:"name" json:"name"`
		FileName    null.String `json:"file_name"`
		TempDOB     int64       `db:"date_of_birth" json:"date_of_birth"`
		DateOfBirth null.Time
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

func (s *Store) GetPlayersTeam(ctx context.Context, teamID int) ([]Player, error) {
	return s.getPlayersTeam(ctx, teamID)
}

func (s *Store) GetPlayer(ctx context.Context, i Player) (Player, error) {
	return s.getPlayer(ctx, i)
}

func (s *Store) AddPlayer(ctx context.Context, i Player) (Player, error) {
	return s.addPlayer(ctx, i)
}

func (s *Store) EditPlayer(ctx context.Context, i Player) (Player, error) {
	return s.editPlayer(ctx, i)
}

func (s *Store) DeletePlayer(ctx context.Context, i Player) error {
	return s.deletePlayer(ctx, i)
}
