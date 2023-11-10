package programme

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
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

func (s *Store) GetProgrammesSeason(ctx context.Context, seasonParam Season) ([]Programme, error) {
	return s.getProgrammesSeason(ctx, seasonParam)
}

func (s *Store) GetProgramme(ctx context.Context, programmeParam Programme) (Programme, error) {
	return s.getProgramme(ctx, programmeParam)
}

func (s *Store) AddProgramme(ctx context.Context, programmeParam Programme) (Programme, error) {
	return s.addProgramme(ctx, programmeParam)
}

func (s *Store) EditProgramme(ctx context.Context, programmeParam Programme) (Programme, error) {
	return s.editProgramme(ctx, programmeParam)
}

func (s *Store) DeleteProgramme(ctx context.Context, programmeParam Programme) error {
	return s.deleteProgramme(ctx, programmeParam)
}

func (s *Store) GetSeasons(ctx context.Context) ([]Season, error) {
	return s.getSeasons(ctx)
}

func (s *Store) GetSeason(ctx context.Context, seasonParam Season) (Season, error) {
	return s.getSeason(ctx, seasonParam)
}

func (s *Store) AddSeason(ctx context.Context, seasonParam Season) (Season, error) {
	return s.addSeason(ctx, seasonParam)
}

func (s *Store) EditSeason(ctx context.Context, seasonParam Season) (Season, error) {
	return s.editSeason(ctx, seasonParam)
}

func (s *Store) DeleteSeason(ctx context.Context, seasonParam Season) error {
	return s.deleteSeason(ctx, seasonParam)
}
