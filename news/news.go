package news

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"
)

type (
	// Store stores the dependencies
	Store struct {
		db *sqlx.DB
	}

	News struct {
		ID       int         `db:"id" json:"id"`
		Title    string      `db:"title" json:"title"`
		FileName null.String `db:"file_name" json:"file_name"`
		Content  null.String `db:"content" json:"content"`
		Date     time.Time   `db:"date" json:"date"`
	}
)

// NewNewsRepo stores our dependency
func NewNewsRepo(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetNews(ctx context.Context) ([]News, error) {
	return s.getNews(ctx)
}

func (s *Store) GetNewsLatest(ctx context.Context) (News, error) {
	return s.getNewsLatest(ctx)
}

func (s *Store) GetNewsArticle(ctx context.Context, newsParam News) (News, error) {
	return s.getNewsArticle(ctx, newsParam)
}

func (s *Store) AddNews(ctx context.Context, newsParam News) (News, error) {
	newsParam.Date = time.Now()
	return s.addNews(ctx, newsParam)
}

func (s *Store) EditNews(ctx context.Context, newsParam News) (News, error) {
	newsDB, err := s.GetNewsArticle(ctx, newsParam)
	if err != nil {
		return News{}, fmt.Errorf("failed to get news for editNews: %w", err)
	}
	if newsDB.Title != newsParam.Title {
		newsDB.Title = newsParam.Title
	}
	if newsParam.FileName.Valid && (!newsDB.FileName.Valid || newsDB.FileName.String != newsParam.FileName.String) {
		newsDB.FileName = newsParam.FileName
	}
	if newsParam.Content.Valid && (!newsDB.Content.Valid || newsDB.Content.String != newsParam.Content.String) {
		newsDB.Content = newsParam.Content
	}
	return s.editNews(ctx, newsDB)
}

func (s *Store) DeleteNews(ctx context.Context, newsParam News) error {
	return s.deleteNews(ctx, newsParam)
}
