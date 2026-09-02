package affiliation

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/utils"
)

func (s *Store) getAffiliations(ctx context.Context) ([]Affiliation, error) {
	ctx, span := tracer.Start(ctx, "affiliation.getAffiliations")
	defer span.End()
	var affiliationsDB []Affiliation
	builder := sq.Select("id", "name", "website", "file_name").
		From("affiliations").
		OrderBy("name")
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for get affiliations: %w", err))
	}
	err = s.db.SelectContext(ctx, &affiliationsDB, sql, args...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get affiliations: %w", err)
	}
	return affiliationsDB, nil
}

func (s *Store) getAffiliationsMinimal(ctx context.Context) ([]Affiliation, error) {
	ctx, span := tracer.Start(ctx, "affiliation.getAffiliationsMinimal")
	defer span.End()
	var affiliationsDB []Affiliation
	builder := sq.Select("id", "name", "website").
		From("affiliations").
		OrderBy("name")
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for get affiliations: %w", err))
	}
	err = s.db.SelectContext(ctx, &affiliationsDB, sql, args...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get affiliations: %w", err)
	}
	return affiliationsDB, nil
}

func (s *Store) getAffiliation(ctx context.Context, affiliationParam Affiliation) (Affiliation, error) {
	ctx, span := tracer.Start(ctx, "affiliation.getAffiliation")
	defer span.End()
	var affiliationDB Affiliation
	builder := utils.PSQL().Select("id", "name", "website", "file_name").
		From("affiliations").
		Where(sq.Eq{"id": affiliationParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for get affiliation: %w", err))
	}
	err = s.db.GetContext(ctx, &affiliationDB, sql, args...)
	if err != nil {
		span.RecordError(err)
		return Affiliation{}, fmt.Errorf("failed to get affiliation: %w", err)
	}
	return affiliationDB, nil
}

func (s *Store) addAffiliation(ctx context.Context, affiliationParam Affiliation) (Affiliation, error) {
	ctx, span := tracer.Start(ctx, "affiliation.addAffiliation")
	defer span.End()
	builder := utils.PSQL().Insert("affiliations").
		Columns("name", "website", "file_name").
		Values(affiliationParam.Name, affiliationParam.Website, affiliationParam.FileName)
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for addAffiliation: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		span.RecordError(err)
		return Affiliation{}, fmt.Errorf("failed to add affiliation: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return Affiliation{}, fmt.Errorf("failed to add affiliation: %w", err)
	}
	return affiliationParam, nil
}

func (s *Store) deleteAffiliation(ctx context.Context, affiliationParam Affiliation) error {
	ctx, span := tracer.Start(ctx, "affiliation.deleteAffiliation")
	defer span.End()
	builder := utils.PSQL().Delete("affiliations").
		Where(sq.Eq{"id": affiliationParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for deleteAffiliation: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete affiliation: %w", err)
	}
	return nil
}
