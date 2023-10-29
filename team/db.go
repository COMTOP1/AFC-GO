package team

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/utils"
)

func (s *Store) getTeams(ctx context.Context) ([]Team, error) {
	var t []Team
	builder := sq.Select("id", "name", "league", "division", "league_table", "fixtures", "coach", "physio", "image", "file_name", "active", "youth", "ages").
		From("afc.teams").
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getTeams: %w", err))
	}
	err = s.db.SelectContext(ctx, &t, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}
	return t, nil
}

func (s *Store) getTeamsActive(ctx context.Context) ([]Team, error) {
	var t []Team
	builder := utils.MySQL().Select("id", "name", "league", "division", "league_table", "fixtures", "coach", "physio", "image", "file_name", "active", "youth", "ages").
		From("afc.teams").
		Where(sq.Eq{"active": true}).
		OrderBy("name")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getTeamsSeason: %w", err))
	}
	err = s.db.SelectContext(ctx, &t, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}
	return t, nil
}

func (s *Store) getTeam(ctx context.Context, t Team) (Team, error) {
	var t1 Team
	builder := utils.MySQL().Select("id", "name", "league", "division", "league_table", "fixtures", "coach", "physio", "image", "file_name", "active", "youth", "ages").
		From("afc.teams").
		Where(sq.Eq{"id": t.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getTeam: %w", err))
	}
	err = s.db.GetContext(ctx, &t1, sql, args...)
	if err != nil {
		return Team{}, fmt.Errorf("failed to get team: %w", err)
	}
	return t1, nil
}

func (s *Store) addTeam(ctx context.Context, t Team) (Team, error) {
	builder := utils.MySQL().Insert("afc.teams").
		Columns("name", "league", "division", "league_table", "fixtures", "coach", "physio", "image", "file_name", "active", "youth", "ages").
		Values(t.Name, t.League, t.Division, t.LeagueTable, t.Fixtures, t.Coach, t.Physio, t.Image, t.FileName, t.IsActive, t.IsYouth, t.Ages)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addTeam: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Team{}, fmt.Errorf("failed to add team: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return Team{}, fmt.Errorf("failed to add team: %w", err)
	}
	if rows < 1 {
		return Team{}, fmt.Errorf("failed to add team: invalid rows affected: %d", rows)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Team{}, fmt.Errorf("failed to add team: %w", err)
	}
	t.ID = int(id)
	return t, nil
}

func (s *Store) editTeam(ctx context.Context, t Team) (Team, error) {
	builder := utils.MySQL().Update("afc.teams").
		SetMap(map[string]interface{}{
			"name":         t.Name,
			"league":       t.League,
			"division":     t.Division,
			"league_table": t.LeagueTable,
			"fixtures":     t.Fixtures,
			"coach":        t.Coach,
			"physio":       t.Physio,
			"image":        t.Image,
			"file_name":    t.FileName,
			"active":       t.IsActive,
			"youth":        t.IsYouth,
			"ages":         t.Ages,
		}).
		Where(sq.Eq{"id": t.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for editTeam: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Team{}, fmt.Errorf("failed to edit team: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return Team{}, fmt.Errorf("failed to edit team: %w", err)
	}
	if rows < 1 {
		return Team{}, fmt.Errorf("failed to edit team: invalid rows affected: %d, this team may not exist: %d", rows, t.ID)
	}
	return t, nil
}

func (s *Store) deleteTeam(ctx context.Context, t Team) error {
	builder := utils.MySQL().Delete("afc.teams").
		Where(sq.Eq{"id": t.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deleteTeam: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}
	return nil
}
