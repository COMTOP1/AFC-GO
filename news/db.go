package news

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/utils"
)

func (s *Store) getNews(ctx context.Context) ([]News, error) {
	var newsDB []News
	builder := sq.Select("id", "title", "file_name", "content", "date").
		From("afc.news").
		OrderBy("date DESC")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for get news: %w", err))
	}
	err = s.db.SelectContext(ctx, &newsDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get news: %w", err)
	}
	return newsDB, nil
}

func (s *Store) getNewsLatest(ctx context.Context) (News, error) {
	var newsDB News
	builder := sq.Select("id", "title", "date").
		From("afc.news").
		OrderBy("date DESC").
		Limit(1)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for get news latest: %w", err))
	}
	err = s.db.GetContext(ctx, &newsDB, sql, args...)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return News{}, nil
		}
		return News{}, fmt.Errorf("failed to get news latest: %w", err)
	}
	return newsDB, nil
}

func (s *Store) getNewsArticle(ctx context.Context, newsParam News) (News, error) {
	var newsDB News
	builder := utils.PSQL().Select("id", "title", "file_name", "content", "date").
		From("afc.news").
		Where(sq.Eq{"id": newsParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for get news article: %w", err))
	}
	err = s.db.GetContext(ctx, &newsDB, sql, args...)
	if err != nil {
		return News{}, fmt.Errorf("failed to get news article: %w", err)
	}
	return newsDB, nil
}

func (s *Store) addNews(ctx context.Context, newsParam News) (News, error) {
	builder := utils.PSQL().Insert("afc.news").
		Columns("title", "file_name", "content", "date").
		Values(newsParam.Title, newsParam.FileName, newsParam.Content, newsParam.Date)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for add news: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return News{}, fmt.Errorf("failed to add news: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return News{}, fmt.Errorf("failed to add news: %w", err)
	}
	return newsParam, nil
}

func (s *Store) editNews(ctx context.Context, newsParam News) (News, error) {
	builder := utils.PSQL().Update("afc.news").
		SetMap(map[string]interface{}{
			"title":     newsParam.Title,
			"file_name": newsParam.FileName,
			"content":   newsParam.Content,
			"date":      newsParam.Date,
		}).
		Where(sq.Eq{"id": newsParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for edit news: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return News{}, fmt.Errorf("failed to edit news: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return News{}, fmt.Errorf("failed to edit news: %w", err)
	}
	return newsParam, nil
}

func (s *Store) deleteNews(ctx context.Context, newsParam News) error {
	builder := utils.PSQL().Delete("afc.news").
		Where(sq.Eq{"id": newsParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for delete news: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete news: %w", err)
	}
	return nil
}
