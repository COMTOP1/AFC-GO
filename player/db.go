package player

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/utils"
)

func (s *Store) getPlayers(ctx context.Context) ([]Player, error) {
	var p []Player
	builder := sq.Select("id", "name", "image", "file_name", "date_of_birth", "position", "captain", "team_id").
		From("afc.players").
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getPlayers: %w", err))
	}
	err = s.db.SelectContext(ctx, &p, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	return p, nil
}

func (s *Store) getPlayersTeam(ctx context.Context, teamID int) ([]Player, error) {
	var p []Player
	builder := sq.Select("id", "name", "image", "file_name", "date_of_birth", "position", "captain", "team_id").
		From("afc.players").
		Where(sq.Eq{"team_id": teamID}).
		OrderBy("name")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getPlayersTeam: %w", err))
	}
	err = s.db.SelectContext(ctx, &p, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	return p, nil
}

func (s *Store) getPlayer(ctx context.Context, p Player) (Player, error) {
	var p1 Player
	builder := sq.Select("id", "name", "image", "file_name", "date_of_birth", "position", "captain", "team_id").
		From("afc.players").
		Where(sq.Eq{"id": p.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getPlayer: %w", err))
	}
	err = s.db.GetContext(ctx, &p1, sql, args...)
	if err != nil {
		return Player{}, fmt.Errorf("failed to get player: %w", err)
	}
	return p1, nil
}

func (s *Store) addPlayer(ctx context.Context, p Player) (Player, error) {
	builder := utils.MySQL().Insert("afc.players").
		Columns("name", "image", "file_name", "date_of_birth", "position", "captain", "team_id").
		Values(p.Name, p.Image, p.FileName, p.TempDOB, p.Position, p.IsCaptain, p.TeamID)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addPlayer: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Player{}, fmt.Errorf("failed to add player: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return Player{}, fmt.Errorf("failed to add player: %w", err)
	}
	if rows < 1 {
		return Player{}, fmt.Errorf("failed to add player: invalid rows affected: %d", rows)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Player{}, fmt.Errorf("failed to add player: %w", err)
	}
	p.ID = int(id)
	return p, nil
}

func (s *Store) editPlayer(ctx context.Context, p Player) (Player, error) {
	builder := utils.MySQL().Update("afc.players").
		SetMap(map[string]interface{}{
			"name":          p.Name,
			"image":         p.Image,
			"file_name":     p.FileName,
			"date_of_birth": p.TempDOB,
			"position":      p.Position,
			"captain":       p.IsCaptain,
			"team_id":       p.TeamID,
		}).
		Where(sq.Eq{"id": p.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for editPlayer: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Player{}, fmt.Errorf("failed to edit player: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return Player{}, fmt.Errorf("failed to edit player: %w", err)
	}
	if rows < 1 {
		return Player{}, fmt.Errorf("failed to edit player: invalid rows affected: %d, this player may not exist: %d", rows, p.ID)
	}
	return p, nil
}

func (s *Store) deletePlayer(ctx context.Context, p Player) error {
	builder := utils.MySQL().Delete("afc.players").
		Where(sq.Eq{"id": p.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deletePlayer: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete player: %w", err)
	}
	return nil
}
