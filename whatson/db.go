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
	ctx, span := tracer.Start(ctx, "whatson.getWhatsOn")
	defer span.End()
	var whatsOnsDB []WhatsOn
	builder := sq.Select("id", "title", "file_name", "content", "date", "date_of_event").
		From("whatson").
		OrderBy("date_of_event")
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for get whats on: %w", err))
	}
	err = s.db.SelectContext(ctx, &whatsOnsDB, sql, args...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get whats on: %w", err)
	}
	return whatsOnsDB, nil
}

func (s *Store) getWhatsOnFuture(ctx context.Context) ([]WhatsOn, error) {
	ctx, span := tracer.Start(ctx, "whatson.getWhatsOnFuture")
	defer span.End()
	var whatsOnsDB []WhatsOn
	builder := utils.PSQL().Select("id", "title", "file_name", "content", "date", "date_of_event").
		From("whatson").
		Where(sq.GtOrEq{"date_of_event": time.Now().Format("2006-01-02")}).
		OrderBy("date_of_event")
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for get whats on future: %w", err))
	}
	err = s.db.SelectContext(ctx, &whatsOnsDB, sql, args...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get whats on: %w", err)
	}
	return whatsOnsDB, nil
}

func (s *Store) getWhatsOnPast(ctx context.Context) ([]WhatsOn, error) {
	ctx, span := tracer.Start(ctx, "whatson.getWhatsOnPast")
	defer span.End()
	var whatsOnsDB []WhatsOn
	builder := utils.PSQL().Select("id", "title", "file_name", "content", "date", "date_of_event").
		From("whatson").
		Where(sq.Lt{"date_of_event": time.Now().Format("2006-01-02")}).
		OrderBy("date_of_event DESC")
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for get whats on past: %w", err))
	}
	err = s.db.SelectContext(ctx, &whatsOnsDB, sql, args...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get whats on: %w", err)
	}
	return whatsOnsDB, nil
}

func (s *Store) getWhatsOnLatest(ctx context.Context) (WhatsOn, error) {
	ctx, span := tracer.Start(ctx, "whatson.getWhatsOnLatest")
	defer span.End()
	var whatsOnDB WhatsOn
	builder := utils.PSQL().Select("id", "title", "date", "date_of_event").
		From("whatson").
		Where(sq.GtOrEq{"date_of_event": time.Now().Format("2006-01-02")}).
		OrderBy("date_of_event ASC").
		Limit(1)
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for get whats on latest: %w", err))
	}
	err = s.db.GetContext(ctx, &whatsOnDB, sql, args...)
	if err != nil {
		span.RecordError(err)
		if strings.Contains(err.Error(), "no rows in result set") {
			return WhatsOn{}, nil
		}
		return WhatsOn{}, fmt.Errorf("failed to get whats on latest: %w", err)
	}

	return whatsOnDB, nil
}

func (s *Store) getWhatsOnArticle(ctx context.Context, whatsOnParam WhatsOn) (WhatsOn, error) {
	ctx, span := tracer.Start(ctx, "whatson.getWhatsOnArticle")
	defer span.End()
	var whatsOnDB WhatsOn
	builder := utils.PSQL().Select("id", "title", "file_name", "content", "date", "date_of_event").
		From("whatson").
		Where(sq.Eq{"id": whatsOnParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for get whats on article: %w", err))
	}
	err = s.db.GetContext(ctx, &whatsOnDB, sql, args...)
	if err != nil {
		span.RecordError(err)
		return WhatsOn{}, fmt.Errorf("failed to get whats on: %w", err)
	}
	return whatsOnDB, nil
}

func (s *Store) addWhatsOn(ctx context.Context, whatsOnParam WhatsOn) (WhatsOn, error) {
	ctx, span := tracer.Start(ctx, "whatson.addWhatsOn")
	defer span.End()
	builder := utils.PSQL().Insert("whatson").
		Columns("title", "file_name", "content", "date", "date_of_event").
		Values(whatsOnParam.Title, whatsOnParam.FileName, whatsOnParam.Content, whatsOnParam.Date, whatsOnParam.DateOfEvent)
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for add whats on: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		span.RecordError(err)
		return WhatsOn{}, fmt.Errorf("failed to add whats on: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return WhatsOn{}, fmt.Errorf("failed to add whats on: %w", err)
	}
	return whatsOnParam, nil
}

func (s *Store) editWhatsOn(ctx context.Context, whatsOnParam WhatsOn) (WhatsOn, error) {
	ctx, span := tracer.Start(ctx, "whatson.editWhatsOn")
	defer span.End()
	builder := utils.PSQL().Update("whatson").
		SetMap(map[string]interface{}{
			"title":         whatsOnParam.Title,
			"file_name":     whatsOnParam.FileName,
			"content":       whatsOnParam.Content,
			"date":          whatsOnParam.Date,
			"date_of_event": whatsOnParam.DateOfEvent,
		}).
		Where(sq.Eq{"id": whatsOnParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for edit whats on: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		span.RecordError(err)
		return WhatsOn{}, fmt.Errorf("failed to edit whats on: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return WhatsOn{}, fmt.Errorf("failed to edit whats on: %w", err)
	}
	return whatsOnParam, nil
}

func (s *Store) deleteWhatsOn(ctx context.Context, whatsOnParam WhatsOn) error {
	ctx, span := tracer.Start(ctx, "whatson.deleteWhatsOn")
	defer span.End()
	builder := utils.PSQL().Delete("whatson").
		Where(sq.Eq{"id": whatsOnParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for delete whats on: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete whats on: %w", err)
	}
	return nil
}
