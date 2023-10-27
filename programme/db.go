package programme

import (
	"context"
	"fmt"
	"github.com/COMTOP1/AFC-GO/utils"
	sq "github.com/Masterminds/squirrel"
)

func (s *Store) getProgrammes(ctx context.Context) ([]Programme, error) {
	var p []Programme
	builder := sq.Select("id", "name", "file_name", "date_of_programme", "programme_season_id").
		From("afc.programmes").
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getProgrammes: %w", err))
	}
	err = s.db.SelectContext(ctx, &p, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get programmes: %w", err)
	}
	return p, nil
}

func (s *Store) getProgrammesSeason(ctx context.Context, seasonID int) ([]Programme, error) {
	var p []Programme
	builder := sq.Select("id", "name", "file_name", "date_of_programme", "programme_season_id").
		From("afc.programmes").
		Where(sq.Eq{"programme_season_id": seasonID}).
		OrderBy("name")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getProgrammesSeason: %w", err))
	}
	err = s.db.SelectContext(ctx, &p, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get programmes: %w", err)
	}
	return p, nil
}

func (s *Store) getProgramme(ctx context.Context, p Programme) (Programme, error) {
	var p1 Programme
	builder := sq.Select("id", "name", "file_name", "date_of_programme", "programme_season_id").
		From("afc.programmes").
		Where(sq.Eq{"id": p.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getProgramme: %w", err))
	}
	err = s.db.GetContext(ctx, &p1, sql, args...)
	if err != nil {
		return Programme{}, fmt.Errorf("failed to get programme: %w", err)
	}
	return p1, nil
}

func (s *Store) addProgramme(ctx context.Context, p Programme) (Programme, error) {
	builder := utils.MySQL().Insert("afc.programmes").
		Columns("name", "file_name", "date_of_programme", "programme_season_id").
		Values(p.Name, p.FileName, p.TempDOP, p.SeasonID)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addProgramme: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Programme{}, fmt.Errorf("failed to add programme: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return Programme{}, fmt.Errorf("failed to add programme: %w", err)
	}
	if rows < 1 {
		return Programme{}, fmt.Errorf("failed to add programme: invalid rows affected: %d", rows)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Programme{}, fmt.Errorf("failed to add programme: %w", err)
	}
	p.ID = int(id)
	return p, nil
}

func (s *Store) editProgramme(ctx context.Context, p Programme) (Programme, error) {
	builder := utils.MySQL().Update("afc.programmes").
		SetMap(map[string]interface{}{
			"name":                p.Name,
			"file_name":           p.FileName,
			"date_of_programme":   p.TempDOP,
			"programme_season_id": p.SeasonID,
		}).
		Where(sq.Eq{"id": p.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for editProgramme: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Programme{}, fmt.Errorf("failed to edit programme: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return Programme{}, fmt.Errorf("failed to edit programme: %w", err)
	}
	if rows < 1 {
		return Programme{}, fmt.Errorf("failed to edit programme: invalid rows affected: %d, this programme may not exist: %d", rows, p.ID)
	}
	return p, nil
}

func (s *Store) deleteProgramme(ctx context.Context, p Programme) error {
	builder := utils.MySQL().Delete("afc.programmes").
		Where(sq.Eq{"id": p.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deleteProgramme: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete programme: %w", err)
	}
	return nil
}

func (s *Store) getSeasons(ctx context.Context) ([]Season, error) {
	var s1 []Season
	builder := sq.Select("id", "name", "file_name", "date_of_programme", "programme_season_id").
		From("afc.programme_seasons").
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getSeasons: %w", err))
	}
	err = s.db.SelectContext(ctx, &s1, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get seasons: %w", err)
	}
	return s1, nil
}

func (s *Store) getSeason(ctx context.Context, s1 Season) (Season, error) {
	var s2 Season
	builder := sq.Select("id", "name", "file_name", "date_of_programme", "programme_season_id").
		From("afc.programme_seasons").
		Where(sq.Eq{"id": s1.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getSeason: %w", err))
	}
	err = s.db.GetContext(ctx, &s2, sql, args...)
	if err != nil {
		return Season{}, fmt.Errorf("failed to get season: %w", err)
	}
	return s2, nil
}

func (s *Store) addSeason(ctx context.Context, s1 Season) (Season, error) {
	builder := utils.MySQL().Insert("afc.programme_seasons").
		Columns("season").
		Values(s1.Season)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addSeason: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Season{}, fmt.Errorf("failed to add season: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return Season{}, fmt.Errorf("failed to add season: %w", err)
	}
	if rows < 1 {
		return Season{}, fmt.Errorf("failed to add season: invalid rows affected: %d", rows)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Season{}, fmt.Errorf("failed to add season: %w", err)
	}
	s1.ID = int(id)
	return s1, nil
}

func (s *Store) editSeason(ctx context.Context, s1 Season) (Season, error) {
	builder := utils.MySQL().Update("afc.programme_seasons").
		SetMap(map[string]interface{}{
			"season": s1.Season,
		}).
		Where(sq.Eq{"id": s1.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for editSeason: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Season{}, fmt.Errorf("failed to edit season: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return Season{}, fmt.Errorf("failed to edit season: %w", err)
	}
	if rows < 1 {
		return Season{}, fmt.Errorf("failed to edit season: invalid rows affected: %d, this season may not exist: %d", rows, s1.ID)
	}
	return s1, nil
}

func (s *Store) deleteSeason(ctx context.Context, s1 Season) error {
	builder := utils.MySQL().Delete("afc.programme_seasons").
		Where(sq.Eq{"id": s1.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deleteSeason: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete season: %w", err)
	}
	return nil
}
