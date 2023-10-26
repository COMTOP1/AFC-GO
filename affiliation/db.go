package affiliation

import (
	"context"
	"fmt"
	"github.com/COMTOP1/AFC-GO/utils"
	sq "github.com/Masterminds/squirrel"
)

func (s *Store) getAffiliations(ctx context.Context) ([]Affiliation, error) {
	var a []Affiliation
	builder := sq.Select("id", "name", "website", "image", "file_name").
		From("afc.affiliations").
		OrderBy("name")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getAffiliations: %w", err))
	}
	err = s.db.SelectContext(ctx, &a, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get affiliations: %w", err)
	}
	return a, nil
}

func (s *Store) getAffiliationsMinimal(ctx context.Context) ([]Affiliation, error) {
	var a []Affiliation
	builder := sq.Select("id", "name", "website").
		From("afc.affiliations").
		OrderBy("name")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getAffiliations: %w", err))
	}
	err = s.db.SelectContext(ctx, &a, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get affiliations: %w", err)
	}
	return a, nil
}

func (s *Store) getAffiliation(ctx context.Context, a Affiliation) (Affiliation, error) {
	var a1 Affiliation
	builder := sq.Select("id", "name", "website", "image", "file_name").
		From("afc.affiliations").
		Where(sq.Eq{"id": a.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getAffiliation: %w", err))
	}
	err = s.db.GetContext(ctx, &a1, sql, args...)
	if err != nil {
		return Affiliation{}, fmt.Errorf("failed to get affiliation: %w", err)
	}
	return a1, nil
}

func (s *Store) addAffiliation(ctx context.Context, a Affiliation) (Affiliation, error) {
	builder := utils.MySQL().Insert("afc.affiliations").
		Columns("name", "website", "image", "file_name").
		Values(a.Name, a.Website, a.Image, a.FileName)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addAffiliation: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Affiliation{}, fmt.Errorf("failed to add affiliation: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return Affiliation{}, fmt.Errorf("failed to add affiliation: %w", err)
	}
	if rows < 1 {
		return Affiliation{}, fmt.Errorf("failed to add affiliation: invalid rows affected: %d", rows)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Affiliation{}, fmt.Errorf("failed to add affiliation: %w", err)
	}
	a.ID = int(id)
	return a, nil
}

func (s *Store) deleteAffiliation(ctx context.Context, a Affiliation) error {
	builder := utils.MySQL().Delete("afc.affiliations").
		Where(sq.Eq{"id": a.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deleteAffiliation: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete affiliation: %w", err)
	}
	return nil
}
