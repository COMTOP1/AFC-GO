package player

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/utils"
)

func (s *Store) getPlayers(ctx context.Context) ([]Player, error) {
	var playersDB []Player
	builder := sq.Select("id", "name", "file_name", "date_of_birth", "position", "captain", "team_id").
		From("afc.players").
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getPlayers: %w", err))
	}
	err = s.db.SelectContext(ctx, &playersDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	return playersDB, nil
}

func (s *Store) getPlayersTeam(ctx context.Context, teamParam team.Team) ([]Player, error) {
	var playersDB []Player
	builder := utils.MySQL().Select("id", "name", "date_of_birth", "position", "captain").
		From("afc.players").
		Where(sq.Eq{"team_id": teamParam.ID}).
		OrderBy("name")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getPlayersTeam: %w", err))
	}
	err = s.db.SelectContext(ctx, &playersDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	return playersDB, nil
}

func (s *Store) getPlayer(ctx context.Context, playerParam Player) (Player, error) {
	var playerDB Player
	builder := utils.MySQL().Select("id", "name", "file_name", "date_of_birth", "position", "captain", "team_id").
		From("afc.players").
		Where(sq.Eq{"id": playerParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getPlayer: %w", err))
	}
	err = s.db.GetContext(ctx, &playerDB, sql, args...)
	if err != nil {
		return Player{}, fmt.Errorf("failed to get player: %w", err)
	}
	return playerDB, nil
}

func (s *Store) addPlayer(ctx context.Context, playerParam Player) (Player, error) {
	builder := utils.MySQL().Insert("afc.players").
		Columns("name", "file_name", "date_of_birth", "position", "captain", "team_id").
		Values(playerParam.Name, playerParam.FileName, playerParam.DateOfBirth, playerParam.Position, playerParam.IsCaptain, playerParam.TeamID)
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
	playerParam.ID = int(id)
	return playerParam, nil
}

func (s *Store) editPlayer(ctx context.Context, playerParam Player) (Player, error) {
	builder := utils.MySQL().Update("afc.players").
		SetMap(map[string]interface{}{
			"name":          playerParam.Name,
			"file_name":     playerParam.FileName,
			"date_of_birth": playerParam.DateOfBirth,
			"position":      playerParam.Position,
			"captain":       playerParam.IsCaptain,
			"team_id":       playerParam.TeamID,
		}).
		Where(sq.Eq{"id": playerParam.ID})
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
		return Player{}, fmt.Errorf("failed to edit player: invalid rows affected: %d, this player may not exist: %d", rows, playerParam.ID)
	}
	return playerParam, nil
}

func (s *Store) deletePlayer(ctx context.Context, playerParam Player) error {
	builder := utils.MySQL().Delete("afc.players").
		Where(sq.Eq{"id": playerParam.ID})
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
