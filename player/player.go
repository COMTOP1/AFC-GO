package player

import (
	"context"
	"fmt"

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
	return s.addPlayer(ctx, playerParam)
}

func (s *Store) EditPlayer(ctx context.Context, playerParam Player) (Player, error) {
	playerDB, err := s.GetPlayer(ctx, playerParam)
	if err != nil {
		return Player{}, fmt.Errorf("failed to get player for editPlayer: %w", err)
	}
	if playerDB.Name != playerParam.Name {
		playerDB.Name = playerParam.Name
	}
	if playerParam.FileName.Valid && (!playerDB.FileName.Valid || playerDB.FileName.String != playerParam.FileName.String) {
		playerDB.FileName = playerParam.FileName
	}
	if playerParam.DateOfBirth.Valid && playerDB.DateOfBirth.Time != playerParam.DateOfBirth.Time {
		playerDB.DateOfBirth = playerParam.DateOfBirth
	}
	if playerParam.Position.Valid && (!playerDB.Position.Valid || playerDB.Position.String != playerParam.Position.String) {
		playerDB.Position = playerParam.Position
	}
	if playerDB.IsCaptain != playerParam.IsCaptain {
		playerDB.IsCaptain = playerParam.IsCaptain
	}
	if playerDB.TeamID != playerParam.TeamID {
		playerDB.TeamID = playerParam.TeamID
	}
	return s.editPlayer(ctx, playerDB)
}

func (s *Store) DeletePlayer(ctx context.Context, playerParam Player) error {
	return s.deletePlayer(ctx, playerParam)
}
