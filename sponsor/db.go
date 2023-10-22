package sponsor

import (
	"context"
	"fmt"
	"github.com/COMTOP1/AFC-GO/utils"
	sq "github.com/Masterminds/squirrel"
)

func (s *Store) getSponsors(ctx context.Context) ([]Sponsor, error) {
	var s1 []Sponsor
	builder := sq.Select("id", "name", "file_name", "date_of_sponsor", "sponsor_season_id").
		From("afc.sponsors").
		OrderBy("id")
	sql, _, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getSponsors: %w", err))
	}
	err = s.db.SelectContext(ctx, &s1, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to get sponsors: %w", err)
	}
	return s1, nil
}

func (s *Store) getSponsorsTeam(ctx context.Context, teamID string) ([]Sponsor, error) {
	var s1 []Sponsor
	builder := sq.Select("id", "name", "file_name", "date_of_sponsor", "sponsor_season_id").
		From("afc.sponsors").
		Where(sq.Eq{"sponsor_season_id": teamID}).
		OrderBy("name")
	sql, _, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getSponsorsSeason: %w", err))
	}
	err = s.db.SelectContext(ctx, &s1, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to get sponsors: %w", err)
	}
	return s1, nil
}

func (s *Store) getSponsor(ctx context.Context, s1 Sponsor) (Sponsor, error) {
	var s2 Sponsor
	builder := sq.Select("id", "name", "file_name", "date_of_sponsor", "sponsor_season_id").
		From("afc.sponsors").
		Where(sq.Eq{"id": s1.ID})
	sql, _, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getSponsor: %w", err))
	}
	err = s.db.SelectContext(ctx, &s2, sql)
	if err != nil {
		return Sponsor{}, fmt.Errorf("failed to get sponsor: %w", err)
	}
	return s2, nil
}

func (s *Store) addSponsor(ctx context.Context, s1 Sponsor) (Sponsor, error) {
	builder := utils.MySQL().Insert("afc.sponsors").
		Columns("name", "website", "image", "file_name", "purpose", "team_id").
		Values(s1.Name, s1.Website, s1.Image, s1.FileName, s1.Purpose, s1.TeamID)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addSponsor: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Sponsor{}, fmt.Errorf("failed to add sponsor: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return Sponsor{}, fmt.Errorf("failed to add sponsor: %w", err)
	}
	if rows < 1 {
		return Sponsor{}, fmt.Errorf("failed to add sponsor: invalid rows affected: %d", rows)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Sponsor{}, fmt.Errorf("failed to add sponsor: %w", err)
	}
	s1.ID = int(id)
	return s1, nil
}

func (s *Store) editSponsor(ctx context.Context, s1 Sponsor) (Sponsor, error) {
	builder := utils.MySQL().Update("afc.sponsors").
		SetMap(map[string]interface{}{
			"name":      s1.Name,
			"website":   s1.Website,
			"image":     s1.Image,
			"file_name": s1.FileName,
			"purpose":   s1.Purpose,
			"team_id":   s1.TeamID,
		}).
		Where(sq.Eq{"id": s1.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for editSponsor: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Sponsor{}, fmt.Errorf("failed to edit sponsor: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return Sponsor{}, fmt.Errorf("failed to edit sponsor: %w", err)
	}
	if rows < 1 {
		return Sponsor{}, fmt.Errorf("failed to edit sponsor: invalid rows affected: %d, this sponsor may not exist: %d", rows, s1.ID)
	}
	return s1, nil
}

func (s *Store) deleteSponsor(ctx context.Context, s1 Sponsor) error {
	builder := utils.MySQL().Delete("afc.sponsors").
		Where(sq.Eq{"id": s1.ID})
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
