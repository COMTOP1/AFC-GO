package news

import (
	"context"
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
		Temp     int64       `db:"date" json:"date"`
		Date     time.Time
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
	return s.addNews(ctx, newsParam)
}

func (s *Store) EditNews(ctx context.Context, newsParam News) (News, error) {
	return s.editNews(ctx, newsParam)
}

func (s *Store) DeleteNews(ctx context.Context, newsParam News) error {
	return s.deleteNews(ctx, newsParam)
}
