package programme

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/utils"
)

func (s *Store) getProgrammes(ctx context.Context) ([]Programme, error) {
	var programmesDB []Programme
	builder2 := utils.PSQL1().
		Select("id", "name", "file_name", "date_of_programme", "programme_season_id").
		From("programmes").
		Where(sq.Lt{"date_of_programme": time.Now().Format("2006-01-02")}).
		OrderBy("date_of_programme DESC")
	builder1 := utils.PSQL1().
		Select("id", "name", "file_name", "date_of_programme", "programme_season_id").
		From("programmes").
		Where(sq.GtOrEq{"date_of_programme": time.Now().Format("2006-01-02")}).
		OrderBy("date_of_programme").
		UnionAll(builder2)
	sql, args, err := builder1.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for get programmes: %w", err))
	}
	err = s.db.SelectContext(ctx, &programmesDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get programmes: %w", err)
	}
	return programmesDB, nil
}

func (s *Store) getProgrammesSeason(ctx context.Context, seasonParam Season) ([]Programme, error) {
	var programmesDB []Programme
	builder2 := utils.PSQL1().
		Select("id", "name", "file_name", "date_of_programme", "programme_season_id").
		From("programmes").
		Where(sq.And{sq.Eq{"programme_season_id": seasonParam.ID}, sq.Lt{"date_of_programme": time.Now().Format("2006-01-02")}}).
		OrderBy("date_of_programme DESC")
	builder1 := utils.PSQL1().
		Select("id", "name", "file_name", "date_of_programme", "programme_season_id").
		From("programmes").
		Where(sq.And{sq.Eq{"programme_season_id": seasonParam.ID}, sq.GtOrEq{"date_of_programme": time.Now().Format("2006-01-02")}}).
		OrderBy("date_of_programme").
		UnionAll(builder2)
	sql, args, err := builder1.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for get programmes season: %w", err))
	}
	err = s.db.SelectContext(ctx, &programmesDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get programmes: %w", err)
	}
	return programmesDB, nil
}

func (s *Store) getProgramme(ctx context.Context, programmeParam Programme) (Programme, error) {
	var programmeDB Programme
	builder := utils.PSQL().Select("id", "name", "file_name", "date_of_programme", "programme_season_id").
		From("afc.programmes").
		Where(sq.Eq{"id": programmeParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for get programme: %w", err))
	}
	err = s.db.GetContext(ctx, &programmeDB, sql, args...)
	if err != nil {
		return Programme{}, fmt.Errorf("failed to get programme: %w", err)
	}
	return programmeDB, nil
}

func (s *Store) addProgramme(ctx context.Context, programmeParam Programme) (Programme, error) {
	builder := utils.PSQL().Insert("afc.programmes").
		Columns("name", "file_name", "date_of_programme", "programme_season_id").
		Values(programmeParam.Name, programmeParam.FileName, programmeParam.DateOfProgramme, programmeParam.SeasonID)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for add programme: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Programme{}, fmt.Errorf("failed to add programme: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return Programme{}, fmt.Errorf("failed to add programme: %w", err)
	}
	return programmeParam, nil
}

func (s *Store) editProgramme(ctx context.Context, programmeParam Programme) (Programme, error) {
	builder := utils.PSQL().Update("afc.programmes").
		SetMap(map[string]interface{}{
			"name":                programmeParam.Name,
			"file_name":           programmeParam.FileName,
			"date_of_programme":   programmeParam.DateOfProgramme,
			"programme_season_id": programmeParam.SeasonID,
		}).
		Where(sq.Eq{"id": programmeParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for edit programme: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Programme{}, fmt.Errorf("failed to edit programme: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return Programme{}, fmt.Errorf("failed to edit programme: %w", err)
	}
	return programmeParam, nil
}

func (s *Store) deleteProgramme(ctx context.Context, programmeParam Programme) error {
	builder := utils.PSQL().Delete("afc.programmes").
		Where(sq.Eq{"id": programmeParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for delete programme: %w", err))
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
		panic(fmt.Errorf("failed to build sql for get seasons: %w", err))
	}
	err = s.db.SelectContext(ctx, &seasonsDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get seasons: %w", err)
	}
	return seasonsDB, nil
}

func (s *Store) getSeason(ctx context.Context, seasonParam Season) (Season, error) {
	var seasonDB Season
	builder := utils.PSQL().Select("id", "season").
		From("afc.programme_seasons").
		Where(sq.Eq{"id": seasonParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for get season: %w", err))
	}
	err = s.db.GetContext(ctx, &seasonDB, sql, args...)
	if err != nil {
		return Season{}, fmt.Errorf("failed to get season: %w", err)
	}
	return seasonDB, nil
}

func (s *Store) addSeason(ctx context.Context, seasonParam Season) (Season, error) {
	builder := utils.PSQL().Insert("afc.programme_seasons").
		Columns("season").
		Values(seasonParam.Season)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for add season: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Season{}, fmt.Errorf("failed to add season: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return Season{}, fmt.Errorf("failed to add season: %w", err)
	}
	return seasonParam, nil
}

func (s *Store) editSeason(ctx context.Context, seasonParam Season) (Season, error) {
	builder := utils.PSQL().Update("afc.programme_seasons").
		SetMap(map[string]interface{}{
			"season": seasonParam.Season,
		}).
		Where(sq.Eq{"id": seasonParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for edit season: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Season{}, fmt.Errorf("failed to edit season: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return Season{}, fmt.Errorf("failed to edit season: %w", err)
	}
	return seasonParam, nil
}

func (s *Store) deleteSeason(ctx context.Context, seasonParam Season) error {
	builder := utils.PSQL().Delete("afc.programme_seasons").
		Where(sq.Eq{"id": seasonParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for delete season: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete season: %w", err)
	}
	return nil
}
