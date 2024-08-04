package team

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/utils"
)

func (s *Store) getTeams(ctx context.Context) ([]Team, error) {
	var teamsDB []Team
	builder := sq.Select("id", "name", "description", "league", "division", "league_table", "fixtures", "coach", "physio", "file_name", "active", "youth", "ages").
		From("afc.teams").
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for get teams: %w", err))
	}
	err = s.db.SelectContext(ctx, &teamsDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}
	return teamsDB, nil
}

func (s *Store) getTeamsActive(ctx context.Context) ([]Team, error) {
	var teamsDB []Team
	builder := utils.PSQL().Select("id", "name", "description", "league", "division", "league_table", "fixtures", "coach", "physio", "file_name", "active", "youth", "ages").
		From("afc.teams").
		Where(sq.Eq{"active": true}).
		OrderBy("name")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for get teams season: %w", err))
	}
	err = s.db.SelectContext(ctx, &teamsDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}
	return teamsDB, nil
}

func (s *Store) getTeam(ctx context.Context, teamParam Team) (Team, error) {
	var teamDB Team
	builder := utils.PSQL().Select("id", "name", "description", "league", "division", "league_table", "fixtures", "coach", "physio", "file_name", "active", "youth", "ages").
		From("afc.teams").
		Where(sq.Eq{"id": teamParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for get team: %w", err))
	}
	err = s.db.GetContext(ctx, &teamDB, sql, args...)
	if err != nil {
		return Team{}, fmt.Errorf("failed to get team: %w", err)
	}
	return teamDB, nil
}

func (s *Store) addTeam(ctx context.Context, teamParam Team) (Team, error) {
	builder := utils.PSQL().Insert("afc.teams").
		Columns("name", "description", "league", "division", "league_table", "fixtures", "coach", "physio", "file_name", "active", "youth", "ages").
		Values(teamParam.Name, teamParam.Description, teamParam.League, teamParam.Division, teamParam.LeagueTable, teamParam.Fixtures, teamParam.Coach, teamParam.Physio, teamParam.FileName, teamParam.IsActive, teamParam.IsYouth, teamParam.Ages)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for add team: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Team{}, fmt.Errorf("failed to add team: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return Team{}, fmt.Errorf("failed to add team: %w", err)
	}
	return teamParam, nil
}

func (s *Store) editTeam(ctx context.Context, teamParam Team) (Team, error) {
	builder := utils.PSQL().Update("afc.teams").
		SetMap(map[string]interface{}{
			"name":         teamParam.Name,
			"description":  teamParam.Description,
			"league":       teamParam.League,
			"division":     teamParam.Division,
			"league_table": teamParam.LeagueTable,
			"fixtures":     teamParam.Fixtures,
			"coach":        teamParam.Coach,
			"physio":       teamParam.Physio,
			"file_name":    teamParam.FileName,
			"active":       teamParam.IsActive,
			"youth":        teamParam.IsYouth,
			"ages":         teamParam.Ages,
		}).
		Where(sq.Eq{"id": teamParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for edit team: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Team{}, fmt.Errorf("failed to edit team: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return Team{}, fmt.Errorf("failed to edit team: %w", err)
	}
	return teamParam, nil
}

func (s *Store) deleteTeam(ctx context.Context, teamParam Team) error {
	builder := utils.PSQL().Delete("afc.teams").
		Where(sq.Eq{"id": teamParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for delete team: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}
	return nil
}
