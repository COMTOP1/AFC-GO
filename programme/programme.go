package programme

import (
	"context"
	"github.com/jmoiron/sqlx"
	"time"
)

type (
	// Store stores the dependencies
	Store struct {
		db *sqlx.DB
	}

	Programme struct {
		ID              int    `db:"id" json:"id"`
		Name            string `db:"name" json:"name"`
		FileName        string `db:"file_name" json:"file_name"`
		TempDOP         int64  `db:"date_of_programme" json:"date_of_programme"`
		DateOfProgramme time.Time
		SeasonID        int `db:"programme_season_id" json:"season_id"`
	}

	Season struct {
		ID     int    `db:"id" json:"id"`
		Season string `db:"season" json:"season"`
	}
)

// NewProgrammeRepo stores our dependency
func NewProgrammeRepo(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetProgrammes(ctx context.Context) ([]Programme, error) {
	return s.getProgrammes(ctx)
}

func (s *Store) GetProgrammesSeason(ctx context.Context, seasonID int) ([]Programme, error) {
	return s.getProgrammesSeason(ctx, seasonID)
}

func (s *Store) GetProgramme(ctx context.Context, p Programme) (Programme, error) {
	return s.getProgramme(ctx, p)
}

func (s *Store) AddProgramme(ctx context.Context, p Programme) (Programme, error) {
	return s.addProgramme(ctx, p)
}

func (s *Store) EditProgramme(ctx context.Context, p Programme) (Programme, error) {
	return s.editProgramme(ctx, p)
}

func (s *Store) DeleteProgramme(ctx context.Context, p Programme) error {
	return s.deleteProgramme(ctx, p)
}

func (s *Store) GetSeasons(ctx context.Context) ([]Season, error) {
	return s.getSeasons(ctx)
}

func (s *Store) GetSeason(ctx context.Context, s1 Season) (Season, error) {
	return s.getSeason(ctx, s1)
}

func (s *Store) AddSeason(ctx context.Context, s1 Season) (Season, error) {
	return s.addSeason(ctx, s1)
}

func (s *Store) EditSeason(ctx context.Context, s1 Season) (Season, error) {
	return s.editSeason(ctx, s1)
}

func (s *Store) DeleteSeason(ctx context.Context, s1 Season) error {
	return s.deleteSeason(ctx, s1)
}