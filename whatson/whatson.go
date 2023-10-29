package whatson

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

	WhatsOn struct {
		ID          int         `db:"id" json:"id"`
		Title       string      `db:"title" json:"title"`
		Image       null.String `db:"image" json:"image"`
		FileName    string      `db:"file_name" json:"file_name"`
		Content     string      `db:"content" json:"content"`
		TempDate    int64       `db:"date" json:"date"`
		Date        time.Time
		TempDOE     int64 `db:"date_of_event" json:"date_of_event"`
		DateOfEvent time.Time
	}
)

// NewWhatsOnRepo stores our dependency
func NewWhatsOnRepo(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetWhatsOnS(ctx context.Context) ([]WhatsOn, error) {
	return s.getWhatsOnS(ctx)
}

func (s *Store) GetWhatsOnFuture(ctx context.Context) ([]WhatsOn, error) {
	return s.getWhatsOnFuture(ctx)
}

func (s *Store) GetWhatsOnPast(ctx context.Context) ([]WhatsOn, error) {
	return s.getWhatsOnPast(ctx)
}

func (s *Store) GetWhatsOnLatest(ctx context.Context) (WhatsOn, error) {
	return s.getWhatsOnLatest(ctx)
}

func (s *Store) GetWhatsOn(ctx context.Context, i WhatsOn) (WhatsOn, error) {
	return s.getWhatsOn(ctx, i)
}

func (s *Store) AddWhatsOn(ctx context.Context, i WhatsOn) (WhatsOn, error) {
	return s.addWhatsOn(ctx, i)
}

func (s *Store) EditWhatsOn(ctx context.Context, i WhatsOn) (WhatsOn, error) {
	return s.editWhatsOn(ctx, i)
}

func (s *Store) DeleteWhatsOn(ctx context.Context, i WhatsOn) error {
	return s.deleteWhatsOn(ctx, i)
}
