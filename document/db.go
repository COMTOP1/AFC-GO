package document

import (
	"context"
	"fmt"
	"github.com/COMTOP1/AFC-GO/utils"
	sq "github.com/Masterminds/squirrel"
)

func (s *Store) getDocuments(ctx context.Context) ([]Document, error) {
	var d []Document
	builder := sq.Select("id", "name", "file_name").
		From("afc.documents").
		OrderBy("name")
	sql, _, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getDocuments: %w", err))
	}
	err = s.db.SelectContext(ctx, &d, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents: %w", err)
	}
	return d, nil
}

func (s *Store) getDocument(ctx context.Context, d Document) (Document, error) {
	var d1 Document
	builder := sq.Select("id", "name", "file_name").
		From("afc.documents").
		Where(sq.Eq{"id": d.ID})
	sql, _, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getDocument: %w", err))
	}
	err = s.db.SelectContext(ctx, &d1, sql)
	if err != nil {
		return Document{}, fmt.Errorf("failed to get document: %w", err)
	}
	return d1, nil
}

func (s *Store) addDocument(ctx context.Context, d Document) (Document, error) {
	builder := utils.MySQL().Insert("afc.documents").
		Columns("name", "file_name").
		Values(d.Name, d.FileName)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addDocument: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Document{}, fmt.Errorf("failed to add document: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return Document{}, fmt.Errorf("failed to add document: %w", err)
	}
	if rows < 1 {
		return Document{}, fmt.Errorf("failed to add document: invalid rows affected: %d", rows)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Document{}, fmt.Errorf("failed to add document: %w", err)
	}
	d.ID = int(id)
	return d, nil
}

func (s *Store) deleteDocument(ctx context.Context, d Document) error {
	builder := utils.MySQL().Delete("afc.documents").
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
