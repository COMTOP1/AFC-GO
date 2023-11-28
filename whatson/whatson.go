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

	// WhatsOn represents relevant whatsOn fields
	WhatsOn struct {
		ID          int         `db:"id" json:"id"`
		Title       string      `db:"title" json:"title"`
		FileName    null.String `db:"file_name" json:"file_name"`
		Content     null.String `db:"content" json:"content"`
		Date        time.Time   `db:"date" json:"date"`
		DateOfEvent time.Time   `db:"date_of_event" json:"date_of_event"`
	}
)

// NewWhatsOnRepo stores our dependency
func NewWhatsOnRepo(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetWhatsOn(ctx context.Context) ([]WhatsOn, error) {
	return s.getWhatsOn(ctx)
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

func (s *Store) GetWhatsOnArticle(ctx context.Context, whatsOnParam WhatsOn) (WhatsOn, error) {
	return s.getWhatsOnArticle(ctx, whatsOnParam)
}

func (s *Store) AddWhatsOn(ctx context.Context, whatsOnParam WhatsOn) (WhatsOn, error) {
	whatsOnParam.Date = time.Now()
	return s.addWhatsOn(ctx, whatsOnParam)
}

func (s *Store) EditWhatsOn(ctx context.Context, whatsOnParam WhatsOn) (WhatsOn, error) {
	return s.editWhatsOn(ctx, whatsOnParam)
}

func (s *Store) DeleteWhatsOn(ctx context.Context, whatsOnParam WhatsOn) error {
	return s.deleteWhatsOn(ctx, whatsOnParam)
}
