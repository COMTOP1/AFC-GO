package programme

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/utils"
)

func (s *Store) getProgrammes(ctx context.Context) ([]Programme, error) {
	var programmesDB []Programme
	builder := sq.Select("id", "name", "file_name", "date_of_programme", "programme_season_id").
		From("afc.programmes").
		OrderBy("date_of_programme DESC")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getProgrammes: %w", err))
	}
	err = s.db.SelectContext(ctx, &programmesDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get programmes: %w", err)
	}
	return programmesDB, nil
}

func (s *Store) getProgrammesSeason(ctx context.Context, seasonParam Season) ([]Programme, error) {
	var programmesDB []Programme
	builder := utils.MySQL().Select("id", "name", "file_name", "date_of_programme", "programme_season_id").
		From("afc.programmes").
		Where(sq.Eq{"programme_season_id": seasonParam.ID}).
		OrderBy("name")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getProgrammesSeason: %w", err))
	}
	err = s.db.SelectContext(ctx, &programmesDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get programmes: %w", err)
	}
	return programmesDB, nil
}

func (s *Store) getProgramme(ctx context.Context, programmeParam Programme) (Programme, error) {
	var programmeDB Programme
	builder := utils.MySQL().Select("id", "name", "file_name", "date_of_programme", "programme_season_id").
		From("afc.programmes").
		Where(sq.Eq{"id": programmeParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getProgramme: %w", err))
	}
	err = s.db.GetContext(ctx, &programmeDB, sql, args...)
	if err != nil {
		return Programme{}, fmt.Errorf("failed to get programme: %w", err)
	}
	return programmeDB, nil
}

func (s *Store) addProgramme(ctx context.Context, programmeParam Programme) (Programme, error) {
	builder := utils.MySQL().Insert("afc.programmes").
		Columns("name", "file_name", "date_of_programme", "programme_season_id").
		Values(programmeParam.Name, programmeParam.FileName, programmeParam.DateOfProgramme, programmeParam.SeasonID)
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
	programmeParam.ID = int(id)
	return programmeParam, nil
}

func (s *Store) editProgramme(ctx context.Context, programmeParam Programme) (Programme, error) {
	builder := utils.MySQL().Update("afc.programmes").
		SetMap(map[string]interface{}{
			"name":                programmeParam.Name,
			"file_name":           programmeParam.FileName,
			"date_of_programme":   programmeParam.DateOfProgramme,
			"programme_season_id": programmeParam.SeasonID,
		}).
		Where(sq.Eq{"id": programmeParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for editProgramme: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Programme{}, fmt.Errorf("failed to edit programme: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return Programme{}, fmt.Errorf("failed to edit programme: %w", err)
	}
	//if rows < 1 {
	//	return Programme{}, fmt.Errorf("failed to edit programme: invalid rows affected: %d, this programme may not exist: %d", rows, programmeParam.ID)
	//}
	return programmeParam, nil
}

func (s *Store) deleteProgramme(ctx context.Context, programmeParam Programme) error {
	builder := utils.MySQL().Delete("afc.programmes").
		Where(sq.Eq{"id": programmeParam.ID})
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
	var seasonsDB []Season
	builder := sq.Select("id", "season").
		From("afc.programme_seasons").
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getSeasons: %w", err))
	}
	err = s.db.SelectContext(ctx, &seasonsDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get seasons: %w", err)
	}
	return seasonsDB, nil
}

func (s *Store) getSeason(ctx context.Context, seasonParam Season) (Season, error) {
	var seasonDB Season
	builder := utils.MySQL().Select("id", "season").
		From("afc.programme_seasons").
		Where(sq.Eq{"id": seasonParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getSeason: %w", err))
	}
	err = s.db.GetContext(ctx, &seasonDB, sql, args...)
	if err != nil {
		return Season{}, fmt.Errorf("failed to get season: %w", err)
	}
	return seasonDB, nil
}

func (s *Store) addSeason(ctx context.Context, seasonParam Season) (Season, error) {
	builder := utils.MySQL().Insert("afc.programme_seasons").
		Columns("season").
		Values(seasonParam.Season)
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
	seasonParam.ID = int(id)
	return seasonParam, nil
}

func (s *Store) editSeason(ctx context.Context, seasonParam Season) (Season, error) {
	builder := utils.MySQL().Update("afc.programme_seasons").
		SetMap(map[string]interface{}{
			"season": seasonParam.Season,
		}).
		Where(sq.Eq{"id": seasonParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for editSeason: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Season{}, fmt.Errorf("failed to edit season: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return Season{}, fmt.Errorf("failed to edit season: %w", err)
	}
	//if rows < 1 {
	//	return Season{}, fmt.Errorf("failed to edit season: invalid rows affected: %d, this season may not exist: %d", rows, seasonParam.ID)
	//}
	return seasonParam, nil
}

func (s *Store) deleteSeason(ctx context.Context, seasonParam Season) error {
	builder := utils.MySQL().Delete("afc.programme_seasons").
		Where(sq.Eq{"id": seasonParam.ID})
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
