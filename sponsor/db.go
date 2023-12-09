package sponsor

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/utils"
)

func (s *Store) getSponsors(ctx context.Context) ([]Sponsor, error) {
	var sponsorsDB []Sponsor
	builder := sq.Select("id", "name", "website", "file_name", "purpose", "team_id").
		From("afc.sponsors").
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getSponsors: %w", err))
	}
	err = s.db.SelectContext(ctx, &sponsorsDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get sponsors: %w", err)
	}
	return sponsorsDB, nil
}

func (s *Store) getSponsorsMinimal(ctx context.Context) ([]Sponsor, error) {
	var sponsorsDB []Sponsor
	builder := sq.Select("id", "website").
		From("afc.sponsors").
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getSponsors: %w", err))
	}
	err = s.db.SelectContext(ctx, &sponsorsDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get sponsors: %w", err)
	}
	return sponsorsDB, nil
}

func (s *Store) getSponsorsTeam(ctx context.Context, teamParam team.Team) ([]Sponsor, error) {
	var sponsorsDB []Sponsor
	builder := utils.PSQL().Select("id", "name", "website", "purpose").
		From("afc.sponsors").
		Where(sq.Eq{"team_id": teamParam.ID}).
		OrderBy("name")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getSponsorsSeason: %w", err))
	}
	err = s.db.SelectContext(ctx, &sponsorsDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get sponsors: %w", err)
	}
	return sponsorsDB, nil
}

func (s *Store) getSponsor(ctx context.Context, seasonParam Sponsor) (Sponsor, error) {
	var seasonDB Sponsor
	builder := utils.PSQL().Select("id", "name", "website", "file_name", "purpose", "team_id").
		From("afc.sponsors").
		Where(sq.Eq{"id": seasonParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getSponsor: %w", err))
	}
	err = s.db.GetContext(ctx, &seasonDB, sql, args...)
	if err != nil {
		return Sponsor{}, fmt.Errorf("failed to get sponsor: %w", err)
	}
	return seasonDB, nil
}

func (s *Store) addSponsor(ctx context.Context, seasonParam Sponsor) (Sponsor, error) {
	builder := utils.PSQL().Insert("afc.sponsors").
		Columns("name", "website", "file_name", "purpose", "team_id").
		Values(seasonParam.Name, seasonParam.Website, seasonParam.FileName, seasonParam.Purpose, seasonParam.TeamID)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addSponsor: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Sponsor{}, fmt.Errorf("failed to add sponsor: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return Sponsor{}, fmt.Errorf("failed to add sponsor: %w", err)
	}
	return seasonParam, nil
}

func (s *Store) editSponsor(ctx context.Context, seasonParam Sponsor) (Sponsor, error) {
	builder := utils.PSQL().Update("afc.sponsors").
		SetMap(map[string]interface{}{
			"name":      seasonParam.Name,
			"website":   seasonParam.Website,
			"file_name": seasonParam.FileName,
			"purpose":   seasonParam.Purpose,
			"team_id":   seasonParam.TeamID,
		}).
		Where(sq.Eq{"id": seasonParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for editSponsor: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Sponsor{}, fmt.Errorf("failed to edit sponsor: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return Sponsor{}, fmt.Errorf("failed to edit sponsor: %w", err)
	}
	return seasonParam, nil
}

func (s *Store) deleteSponsor(ctx context.Context, seasonParam Sponsor) error {
	builder := utils.PSQL().Delete("afc.sponsors").
		Where(sq.Eq{"id": seasonParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deleteSponsor: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete sponsor: %w", err)
	}
	return nil
}
