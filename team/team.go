package team

import (
	"context"
	"fmt"

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
		Description null.String `db:"description" json:"description"`
		League      null.String `db:"league" json:"league"`
		Division    null.String `db:"division" json:"division"`
		LeagueTable null.String `db:"league_table" json:"league_table"`
		Fixtures    null.String `db:"fixtures" json:"fixtures"`
		Coach       null.String `db:"coach" json:"coach"`
		Physio      null.String `db:"physio" json:"physio"`
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

func (s *Store) GetTeam(ctx context.Context, teamParam Team) (Team, error) {
	return s.getTeam(ctx, teamParam)
}

func (s *Store) AddTeam(ctx context.Context, teamParam Team) (Team, error) {
	return s.addTeam(ctx, teamParam)
}

func (s *Store) EditTeam(ctx context.Context, teamParam Team) (Team, error) {
	teamDB, err := s.GetTeam(ctx, teamParam)
	if err != nil {
		return Team{}, fmt.Errorf("failed to get team for editTeam: %w", err)
	}
	if teamDB.Name != teamParam.Name {
		teamDB.Name = teamParam.Name
	}
	if teamParam.Description.Valid && (!teamDB.Description.Valid || teamDB.Description.String != teamParam.Description.String) {
		teamDB.Description = teamParam.Description
	}
	if teamParam.League.Valid && (!teamDB.League.Valid || teamDB.League.String != teamParam.League.String) {
		teamDB.League = teamParam.League
	}
	if teamParam.Division.Valid && (!teamDB.Division.Valid || teamDB.Division.String != teamParam.Division.String) {
		teamDB.Division = teamParam.Division
	}
	if teamParam.LeagueTable.Valid && (!teamDB.LeagueTable.Valid || teamDB.LeagueTable.String != teamParam.LeagueTable.String) {
		teamDB.LeagueTable = teamParam.LeagueTable
	}
	if teamParam.Fixtures.Valid && (!teamDB.Fixtures.Valid || teamDB.Fixtures.String != teamParam.Fixtures.String) {
		teamDB.Fixtures = teamParam.Fixtures
	}
	if teamParam.Coach.Valid && (!teamDB.Coach.Valid || teamDB.Coach.String != teamParam.Coach.String) {
		teamDB.Coach = teamParam.Coach
	}
	if teamParam.Physio.Valid && (!teamDB.Physio.Valid || teamDB.Physio.String != teamParam.Physio.String) {
		teamDB.Physio = teamParam.Physio
	}
	if teamParam.FileName.Valid && (!teamDB.FileName.Valid || teamDB.FileName.String != teamParam.FileName.String) {
		teamDB.FileName = teamParam.FileName
	}
	if teamDB.IsActive != teamParam.IsActive {
		teamDB.IsActive = teamParam.IsActive
	}
	if teamDB.IsYouth != teamParam.IsYouth {
		teamDB.IsYouth = teamParam.IsYouth
	}
	if teamDB.Ages != teamParam.Ages {
		teamDB.Ages = teamParam.Ages
	}
	return s.editTeam(ctx, teamParam)
}

func (s *Store) DeleteTeam(ctx context.Context, teamParam Team) error {
	return s.deleteTeam(ctx, teamParam)
}
