package document

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/utils"
)

func (s *Store) getDocuments(ctx context.Context) ([]Document, error) {
	ctx, span := tracer.Start(ctx, "document.getDocuments")
	defer span.End()
	var documentsDB []Document
	builder := sq.Select("id", "name", "file_name").
		From("documents").
		OrderBy("name")
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for get documents: %w", err))
	}
	err = s.db.SelectContext(ctx, &documentsDB, sql, args...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get documents: %w", err)
	}
	return documentsDB, nil
}

func (s *Store) getDocument(ctx context.Context, documentParam Document) (Document, error) {
	ctx, span := tracer.Start(ctx, "document.getDocument")
	defer span.End()
	var documentDB Document
	builder := utils.PSQL().Select("id", "name", "file_name").
		From("documents").
		Where(sq.Eq{"id": documentParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for get document: %w", err))
	}
	err = s.db.GetContext(ctx, &documentDB, sql, args...)
	if err != nil {
		span.RecordError(err)
		return Document{}, fmt.Errorf("failed to get document: %w", err)
	}
	return documentDB, nil
}

func (s *Store) addDocument(ctx context.Context, documentParam Document) (Document, error) {
	ctx, span := tracer.Start(ctx, "document.addDocument")
	defer span.End()
	builder := utils.PSQL().Insert("documents").
		Columns("name", "file_name").
		Values(documentParam.Name, documentParam.FileName)
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for add document: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		span.RecordError(err)
		return Document{}, fmt.Errorf("failed to add document: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return Document{}, fmt.Errorf("failed to add document: %w", err)
	}
	return documentParam, nil
}

func (s *Store) deleteDocument(ctx context.Context, d Document) error {
	ctx, span := tracer.Start(ctx, "document.deleteDocument")
	defer span.End()
	builder := utils.PSQL().Delete("documents").
		Where(sq.Eq{"id": d.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		panic(fmt.Errorf("failed to build sql for delete document: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}
