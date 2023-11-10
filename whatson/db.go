package whatson

import (
	"context"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/utils"
)

func (s *Store) getWhatsOn(ctx context.Context) ([]WhatsOn, error) {
	var whatsOnsDB []WhatsOn
	builder := sq.Select("id", "title", "file_name", "content", "date", "date_of_event").
		From("afc.whatson").
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getWhatsOn: %w", err))
	}
	err = s.db.SelectContext(ctx, &whatsOnsDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get what's on: %w", err)
	}
	return whatsOnsDB, nil
}

func (s *Store) getWhatsOnFuture(ctx context.Context) ([]WhatsOn, error) {
	var whatsOnsDB []WhatsOn
	builder := utils.MySQL().Select("id", "title", "file_name", "content", "date", "date_of_event").
		From("afc.whatson").
		Where(sq.GtOrEq{"date_of_event": time.Now().UnixMilli()}).
		OrderBy("date_of_event")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getWhatsOnFuture: %w", err))
	}
	err = s.db.SelectContext(ctx, &whatsOnsDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get what's on: %w", err)
	}
	return whatsOnsDB, nil
}

func (s *Store) getWhatsOnPast(ctx context.Context) ([]WhatsOn, error) {
	var whatsOnsDB []WhatsOn
	builder := utils.MySQL().Select("id", "title", "file_name", "content", "date", "date_of_event").
		From("afc.whatson").
		Where(sq.Lt{"date_of_event": time.Now().UnixMilli()}).
		OrderBy("date_of_event DESC")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getWhatsOnPast: %w", err))
	}
	err = s.db.SelectContext(ctx, &whatsOnsDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get what's on: %w", err)
	}
	return whatsOnsDB, nil
}

func (s *Store) getWhatsOnLatest(ctx context.Context) (WhatsOn, error) {
	var whatsOnDB WhatsOn
	builder := utils.MySQL().Select("id", "title", "date", "date_of_event").
		From("afc.whatson").
		Where(sq.GtOrEq{"date_of_event": time.Now().UnixMilli()}).
		OrderBy("id DESC").
		Limit(1)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getWhatsOnLatest: %w", err))
	}
	err = s.db.GetContext(ctx, &whatsOnDB, sql, args...)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return WhatsOn{}, nil
		}
		return WhatsOn{}, fmt.Errorf("failed to get what's on latest: %w", err)
	}

	return whatsOnDB, nil
}

func (s *Store) getWhatsOnArticle(ctx context.Context, whatsOnParam WhatsOn) (WhatsOn, error) {
	var whatsOnDB WhatsOn
	builder := utils.MySQL().Select("id", "title", "file_name", "content", "date", "date_of_event").
		From("afc.whatson").
		Where(sq.Eq{"id": whatsOnParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getWhatsOnArticle: %w", err))
	}
	err = s.db.GetContext(ctx, &whatsOnDB, sql, args...)
	if err != nil {
		return WhatsOn{}, fmt.Errorf("failed to get what's on: %w", err)
	}
	return whatsOnDB, nil
}

func (s *Store) addWhatsOn(ctx context.Context, whatsOnParam WhatsOn) (WhatsOn, error) {
	builder := utils.MySQL().Insert("afc.whatson").
		Columns("title", "file_name", "content", "date", "date_of_event").
		Values(whatsOnParam.Title, whatsOnParam.FileName, whatsOnParam.Content, whatsOnParam.Date, whatsOnParam.DateOfEvent)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addWhatsOn: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return WhatsOn{}, fmt.Errorf("failed to add what's on: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return WhatsOn{}, fmt.Errorf("failed to add what's on: %w", err)
	}
	if rows < 1 {
		return WhatsOn{}, fmt.Errorf("failed to add what's on: invalid rows affected: %d", rows)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return WhatsOn{}, fmt.Errorf("failed to add what's on: %w", err)
	}
	whatsOnParam.ID = int(id)
	return whatsOnParam, nil
}

func (s *Store) editWhatsOn(ctx context.Context, whatsOnParam WhatsOn) (WhatsOn, error) {
	builder := utils.MySQL().Update("afc.whatson").
		SetMap(map[string]interface{}{
			"title":         whatsOnParam.Title,
			"file_name":     whatsOnParam.FileName,
			"content":       whatsOnParam.Content,
			"date":          whatsOnParam.TempDate,
			"date_of_event": whatsOnParam.TempDOE,
		}).
		Where(sq.Eq{"id": whatsOnParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for editWhatsOn: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return WhatsOn{}, fmt.Errorf("failed to edit what's on: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return WhatsOn{}, fmt.Errorf("failed to edit what's on: %w", err)
	}
	if rows < 1 {
		return WhatsOn{}, fmt.Errorf("failed to edit what's on: invalid rows affected: %d, this what's on may not exist: %d", rows, whatsOnParam.ID)
	}
	return whatsOnParam, nil
}

func (s *Store) deleteWhatsOn(ctx context.Context, whatsOnParam WhatsOn) error {
	builder := utils.MySQL().Delete("afc.whatson").
		Where(sq.Eq{"id": whatsOnParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deleteWhatsOn: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete what's on: %w", err)
	}
	return nil
}
