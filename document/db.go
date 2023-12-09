package document

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/utils"
)

func (s *Store) getDocuments(ctx context.Context) ([]Document, error) {
	var documentsDB []Document
	builder := sq.Select("id", "name", "file_name").
		From("afc.documents").
		OrderBy("name")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getDocuments: %w", err))
	}
	err = s.db.SelectContext(ctx, &documentsDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents: %w", err)
	}
	return documentsDB, nil
}

func (s *Store) getDocument(ctx context.Context, documentParam Document) (Document, error) {
	var documentDB Document
	builder := utils.PSQL().Select("id", "name", "file_name").
		From("afc.documents").
		Where(sq.Eq{"id": documentParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getDocument: %w", err))
	}
	err = s.db.GetContext(ctx, &documentDB, sql, args...)
	if err != nil {
		return Document{}, fmt.Errorf("failed to get document: %w", err)
	}
	return documentDB, nil
}

func (s *Store) addDocument(ctx context.Context, documentParam Document) (Document, error) {
	builder := utils.PSQL().Insert("afc.documents").
		Columns("name", "file_name").
		Values(documentParam.Name, documentParam.FileName)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addDocument: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Document{}, fmt.Errorf("failed to add document: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return Document{}, fmt.Errorf("failed to add document: %w", err)
	}
	return documentParam, nil
}

func (s *Store) deleteDocument(ctx context.Context, d Document) error {
	builder := utils.PSQL().Delete("afc.documents").
		Where(sq.Eq{"id": d.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deleteDocument: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}
