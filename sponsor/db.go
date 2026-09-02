package sponsor

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/utils"
)

func (s *Store) getSponsors(ctx context.Context) ([]Sponsor, error) {
	ctx, span := tracer.Start(ctx, "sponsor.getSponsors")
	defer span.End()
	var sponsorsDB []Sponsor
	builder := sq.Select("id", "name", "website", "file_name", "purpose", "team_id").
		From("sponsors").
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for get sponsors: %w", err))
	}
	err = s.db.SelectContext(ctx, &sponsorsDB, sql, args...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get sponsors: %w", err)
	}
	return sponsorsDB, nil
}

func (s *Store) getSponsorsMinimal(ctx context.Context) ([]Sponsor, error) {
	ctx, span := tracer.Start(ctx, "sponsor.getSponsorsMinimal")
	defer span.End()
	var sponsorsDB []Sponsor
	builder := sq.Select("id", "website").
		From("sponsors").
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for get sponsors: %w", err))
	}
	err = s.db.SelectContext(ctx, &sponsorsDB, sql, args...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get sponsors: %w", err)
	}
	return sponsorsDB, nil
}

func (s *Store) getSponsorsTeam(ctx context.Context, teamParam team.Team) ([]Sponsor, error) {
	ctx, span := tracer.Start(ctx, "sponsor.getSponsorsTeam")
	defer span.End()
	var sponsorsDB []Sponsor
	builder := utils.PSQL().Select("id", "name", "website", "purpose").
		From("sponsors").
		Where(sq.Eq{"team_id": teamParam.ID}).
		OrderBy("name")
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for get sponsors team: %w", err))
	}
	err = s.db.SelectContext(ctx, &sponsorsDB, sql, args...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get sponsors: %w", err)
	}
	return sponsorsDB, nil
}

func (s *Store) getSponsor(ctx context.Context, sponsorParam Sponsor) (Sponsor, error) {
	ctx, span := tracer.Start(ctx, "sponsor.getSponsor")
	defer span.End()
	var seasonDB Sponsor
	builder := utils.PSQL().Select("id", "name", "website", "file_name", "purpose", "team_id").
		From("sponsors").
		Where(sq.Eq{"id": sponsorParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for get sponsor: %w", err))
	}
	err = s.db.GetContext(ctx, &seasonDB, sql, args...)
	if err != nil {
		span.RecordError(err)
		return Sponsor{}, fmt.Errorf("failed to get sponsor: %w", err)
	}
	return seasonDB, nil
}

func (s *Store) addSponsor(ctx context.Context, sponsorParam Sponsor) (Sponsor, error) {
	ctx, span := tracer.Start(ctx, "sponsor.addSponsor")
	defer span.End()
	builder := utils.PSQL().Insert("sponsors").
		Columns("name", "website", "file_name", "purpose", "team_id").
		Values(sponsorParam.Name, sponsorParam.Website, sponsorParam.FileName, sponsorParam.Purpose, sponsorParam.TeamID)
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for add sponsor: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		span.RecordError(err)
		return Sponsor{}, fmt.Errorf("failed to add sponsor: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return Sponsor{}, fmt.Errorf("failed to add sponsor: %w", err)
	}
	return sponsorParam, nil
}

func (s *Store) editSponsor(ctx context.Context, sponsorParam Sponsor) (Sponsor, error) {
	ctx, span := tracer.Start(ctx, "sponsor.editSponsor")
	defer span.End()
	builder := utils.PSQL().Update("sponsors").
		SetMap(map[string]interface{}{
			"name":      sponsorParam.Name,
			"website":   sponsorParam.Website,
			"file_name": sponsorParam.FileName,
			"purpose":   sponsorParam.Purpose,
			"team_id":   sponsorParam.TeamID,
		}).
		Where(sq.Eq{"id": sponsorParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for edit sponsor: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		span.RecordError(err)
		return Sponsor{}, fmt.Errorf("failed to edit sponsor: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return Sponsor{}, fmt.Errorf("failed to edit sponsor: %w", err)
	}
	return sponsorParam, nil
}

func (s *Store) deleteSponsor(ctx context.Context, sponsorParam Sponsor) error {
	ctx, span := tracer.Start(ctx, "sponsor.deleteSponsor")
	defer span.End()
	builder := utils.PSQL().Delete("sponsors").
		Where(sq.Eq{"id": sponsorParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for delete sponsor: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete sponsor: %w", err)
	}
	return nil
}
