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
		panic(fmt.Errorf("failed to build sql for get players: %w", err))
	}
	err = s.db.SelectContext(ctx, &playersDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	return playersDB, nil
}

func (s *Store) getPlayersTeam(ctx context.Context, teamParam team.Team) ([]Player, error) {
	var playersDB []Player
	builder := utils.PSQL().Select("id", "name", "date_of_birth", "position", "captain").
		From("afc.players").
		Where(sq.Eq{"team_id": teamParam.ID}).
		OrderBy("name")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for get players team: %w", err))
	}
	err = s.db.SelectContext(ctx, &playersDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	return playersDB, nil
}

func (s *Store) getPlayer(ctx context.Context, playerParam Player) (Player, error) {
	var playerDB Player
	builder := utils.PSQL().Select("id", "name", "file_name", "date_of_birth", "position", "captain", "team_id").
		From("afc.players").
		Where(sq.Eq{"id": playerParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for get player: %w", err))
	}
	err = s.db.GetContext(ctx, &playerDB, sql, args...)
	if err != nil {
		return Player{}, fmt.Errorf("failed to get player: %w", err)
	}
	return playerDB, nil
}

func (s *Store) addPlayer(ctx context.Context, playerParam Player) (Player, error) {
	builder := utils.PSQL().Insert("afc.players").
		Columns("name", "file_name", "date_of_birth", "position", "captain", "team_id").
		Values(playerParam.Name, playerParam.FileName, playerParam.DateOfBirth, playerParam.Position, playerParam.IsCaptain, playerParam.TeamID)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for add player: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Player{}, fmt.Errorf("failed to add player: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return Player{}, fmt.Errorf("failed to add player: %w", err)
	}
	return playerParam, nil
}

func (s *Store) editPlayer(ctx context.Context, playerParam Player) (Player, error) {
	builder := utils.PSQL().Update("afc.players").
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
		panic(fmt.Errorf("failed to build sql for edit player: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Player{}, fmt.Errorf("failed to edit player: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return Player{}, fmt.Errorf("failed to edit player: %w", err)
	}
	// if rows < 1 {
	//	return Player{}, fmt.Errorf("failed to edit player: invalid rows affected: %d, this player may not exist: %d", rows, playerParam.ID)
	// }
	return playerParam, nil
}

func (s *Store) deletePlayer(ctx context.Context, playerParam Player) error {
	builder := utils.PSQL().Delete("afc.players").
		Where(sq.Eq{"id": playerParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for delete player: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete player: %w", err)
	}
	return nil
}
