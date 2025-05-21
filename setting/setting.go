package setting

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type (
	// Store stores the dependencies
	Store struct {
		db *sqlx.DB
	}

	Setting struct {
		ID          string `db:"id" json:"id"`
		SettingText string `db:"setting_text" json:"settingText"`
	}
)

// NewSettingRepo stores our dependency
func NewSettingRepo(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetSettings(ctx context.Context) ([]Setting, error) {
	return s.getSettings(ctx)
}

func (s *Store) GetAffiliation(ctx context.Context, settingID string) (Setting, error) {
	return s.getSetting(ctx, settingID)
}

func (s *Store) AddSetting(ctx context.Context, settingParam Setting) (Setting, error) {
	return s.addSetting(ctx, settingParam)
}

func (s *Store) EditSetting(ctx context.Context, settingParam Setting) (Setting, error) {
	return s.editSetting(ctx, settingParam)
}

func (s *Store) DeleteSetting(ctx context.Context, settingID string) error {
	return s.deleteSetting(ctx, settingID)
}
