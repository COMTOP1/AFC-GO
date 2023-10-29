package news

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/utils"
)

func (s *Store) getNewsS(ctx context.Context) ([]News, error) {
	var n []News
	builder := sq.Select("id", "title", "image", "file_name", "content", "date").
		From("afc.news").
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getNewsS: %w", err))
	}
	err = s.db.SelectContext(ctx, &n, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get news: %w", err)
	}
	return n, nil
}

func (s *Store) getNewsLatest(ctx context.Context) (News, error) {
	temp := make([]News, 1)
	builder := sq.Select("id", "title", "date").
		From("afc.news").
		OrderBy("id DESC").
		Limit(1)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getNews: %w", err))
	}
	err = s.db.GetContext(ctx, &temp, sql, args...)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return News{}, nil
		}
		return News{}, fmt.Errorf("failed to get news: %w", err)
	}

	var n1 News
	n1 = temp[0]
	return n1, nil
}

func (s *Store) getNews(ctx context.Context, n News) (News, error) {
	var n1 News
	builder := utils.MySQL().Select("id", "title", "image", "file_name", "content", "date").
		From("afc.news").
		Where(sq.Eq{"id": n.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getNews: %w", err))
	}
	err = s.db.GetContext(ctx, &n1, sql, args...)
	if err != nil {
		return News{}, fmt.Errorf("failed to get news: %w", err)
	}
	return n1, nil
}

func (s *Store) addNews(ctx context.Context, n News) (News, error) {
	builder := utils.MySQL().Insert("afc.news").
		Columns("title", "image", "file_name", "content", "date").
		Values(n.Title, n.Image, n.FileName, n.Content, n.Temp)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addNews: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return News{}, fmt.Errorf("failed to add news: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return News{}, fmt.Errorf("failed to add news: %w", err)
	}
	if rows < 1 {
		return News{}, fmt.Errorf("failed to add news: invalid rows affected: %d", rows)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return News{}, fmt.Errorf("failed to add news: %w", err)
	}
	n.ID = int(id)
	return n, nil
}

func (s *Store) editNews(ctx context.Context, n News) (News, error) {
	builder := utils.MySQL().Update("afc.news").
		SetMap(map[string]interface{}{
			"title":     n.Title,
			"image":     n.Image,
			"file_name": n.FileName,
			"content":   n.Content,
			"date":      n.Temp,
		}).
		Where(sq.Eq{"id": n.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for editNews: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return News{}, fmt.Errorf("failed to edit news: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return News{}, fmt.Errorf("failed to edit news: %w", err)
	}
	if rows < 1 {
		return News{}, fmt.Errorf("failed to edit news: invalid rows affected: %d, this news may not exist: %d", rows, n.ID)
	}
	return n, nil
}

func (s *Store) deleteNews(ctx context.Context, n News) error {
	builder := utils.MySQL().Delete("afc.news").
		Where(sq.Eq{"id": n.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deleteNews: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete news: %w", err)
	}
	return nil
}
