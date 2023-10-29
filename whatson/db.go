package whatson

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/utils"
)

func (s *Store) getWhatsOnS(ctx context.Context) ([]WhatsOn, error) {
	var w []WhatsOn
	builder := sq.Select("id", "title", "image", "file_name", "content", "date", "date_of_event").
		From("afc.whatson").
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getWhatsOns: %w", err))
	}
	err = s.db.SelectContext(ctx, &w, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get what's on: %w", err)
	}
	return w, nil
}

func (s *Store) getWhatsOnFuture(ctx context.Context) ([]WhatsOn, error) {
	var w []WhatsOn
	builder := sq.Select("id", "title", "image", "file_name", "content", "date", "date_of_event").
		From("afc.whatson").
		Where(sq.GtOrEq{"date_of_event": time.Now().UnixMilli()}).
		OrderBy("name")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getWhatsOnsSeason: %w", err))
	}
	err = s.db.SelectContext(ctx, &w, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get what's on: %w", err)
	}
	return w, nil
}

func (s *Store) getWhatsOnPast(ctx context.Context) ([]WhatsOn, error) {
	var w []WhatsOn
	builder := sq.Select("id", "title", "image", "file_name", "content", "date", "date_of_event").
		From("afc.whatson").
		Where(sq.Lt{"date_of_event": time.Now().UnixMilli()}).
		OrderBy("name")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getWhatsOnsSeason: %w", err))
	}
	err = s.db.SelectContext(ctx, &w, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get what's on: %w", err)
	}
	return w, nil
}

func (s *Store) getWhatsOnLatest(ctx context.Context) (WhatsOn, error) {
	var w WhatsOn
	builder := sq.Select("id", "title", "date", "date_of_event").
		From("afc.whatson").
		Where(sq.GtOrEq{"date_of_event": time.Now().UnixMilli()}).
		OrderBy("id DESC").
		Limit(1)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getWhatsOn: %w", err))
	}
	err = s.db.GetContext(ctx, &w, sql, args...)
	if err != nil {
		return WhatsOn{}, fmt.Errorf("failed to get what's on latest: %w", err)
	}

	return w, nil
}

func (s *Store) getWhatsOn(ctx context.Context, w WhatsOn) (WhatsOn, error) {
	var w1 WhatsOn
	builder := sq.Select("id", "title", "image", "file_name", "content", "date", "date_of_event").
		From("afc.whatson").
		Where(sq.Eq{"id": w.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getWhatsOn: %w", err))
	}
	err = s.db.GetContext(ctx, &w1, sql, args...)
	if err != nil {
		return WhatsOn{}, fmt.Errorf("failed to get what's on: %w", err)
	}
	return w1, nil
}

func (s *Store) addWhatsOn(ctx context.Context, w WhatsOn) (WhatsOn, error) {
	builder := utils.MySQL().Insert("afc.whatson").
		Columns("title", "image", "file_name", "content", "date", "date_of_event").
		Values(w.Title, w.Image, w.FileName, w.Content, w.Date, w.DateOfEvent)
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
	w.ID = int(id)
	return w, nil
}

func (s *Store) editWhatsOn(ctx context.Context, w WhatsOn) (WhatsOn, error) {
	builder := utils.MySQL().Update("afc.whatson").
		SetMap(map[string]interface{}{
			"title":         w.Title,
			"image":         w.Image,
			"file_name":     w.FileName,
			"content":       w.Content,
			"date":          w.TempDate,
			"date_of_event": w.TempDOE,
		}).
		Where(sq.Eq{"id": w.ID})
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
		return WhatsOn{}, fmt.Errorf("failed to edit what's on: invalid rows affected: %d, this what's on may not exist: %d", rows, w.ID)
	}
	return w, nil
}

func (s *Store) deleteWhatsOn(ctx context.Context, w WhatsOn) error {
	builder := utils.MySQL().Delete("afc.whatson").
		Where(sq.Eq{"id": w.ID})
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
