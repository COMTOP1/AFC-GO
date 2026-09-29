package programme

import (
	"context"
	"fmt"
	"strings"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

const uploadCategory = "programme"

type store interface {
	GetProgrammes(ctx context.Context) ([]Programme, error)
	GetProgrammesSeason(ctx context.Context, seasonParam Season) ([]Programme, error)
	GetProgramme(ctx context.Context, programmeParam Programme) (Programme, error)
	AddProgramme(ctx context.Context, programmeParam Programme) (Programme, error)
	EditProgramme(ctx context.Context, programmeParam Programme) (Programme, error)
	DeleteProgramme(ctx context.Context, programmeParam Programme) error
	GetSeasons(ctx context.Context) ([]Season, error)
	GetSeason(ctx context.Context, seasonParam Season) (Season, error)
	AddSeason(ctx context.Context, seasonParam Season) (Season, error)
	EditSeason(ctx context.Context, seasonParam Season) (Season, error)
	DeleteSeason(ctx context.Context, seasonParam Season) error
}

// Service is the programme and season logic shared by the API and legacy views.
type Service struct {
	store store
	files *upload.Files
}

func NewService(store store, files *upload.Files) *Service {
	return &Service{store: store, files: files}
}

func (s *Service) public(p Programme, seasons map[int]Season) Public {
	out := Public{ID: p.ID, Name: p.Name, Date: p.DateOfProgramme, FileURL: s.files.URL(p.FileName)}
	if season, ok := seasons[p.SeasonID]; ok {
		out.Season = &PublicSeason{ID: season.ID, Name: season.Season}
	}
	return out
}

func (s *Service) seasonMap(ctx context.Context) (map[int]Season, error) {
	rows, err := s.store.GetSeasons(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list seasons: %w", err)
	}
	m := make(map[int]Season, len(rows))
	for _, r := range rows {
		m[r.ID] = r
	}
	return m, nil
}

// List returns programmes, newest first; seasonID 0 lists every season.
func (s *Service) List(ctx context.Context, seasonID int) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "programme.Service.List")
	defer span.End()

	var (
		rows []Programme
		err  error
	)
	if seasonID == 0 {
		rows, err = s.store.GetProgrammes(ctx)
	} else {
		var season Season
		if season, err = s.store.GetSeason(ctx, Season{ID: seasonID}); err != nil {
			return nil, svcerr.FromStore(err, "season")
		}
		rows, err = s.store.GetProgrammesSeason(ctx, season)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list programmes: %w", err)
	}
	seasons, err := s.seasonMap(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Public, 0, len(rows))
	for _, p := range rows {
		out = append(out, s.public(p, seasons))
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, in CreateInput, file *upload.File) (Public, error) {
	ctx, span := tracer.Start(ctx, "programme.Service.Create")
	defer span.End()

	name := strings.TrimSpace(in.Name)
	fields := svcerr.Fields{}
	if name == "" {
		fields.Add("name", "name is required")
	}
	if in.Date.IsZero() {
		fields.Add("date", "date is required")
	}
	if file == nil {
		fields.Add("file", "file is required")
	}
	if in.SeasonID != 0 {
		if _, err := s.store.GetSeason(ctx, Season{ID: in.SeasonID}); err != nil {
			fields.Add("seasonId", "season does not exist")
		}
	}
	if err := fields.Err(); err != nil {
		return Public{}, err
	}

	key, err := s.files.Save(ctx, file, uploadCategory)
	if err != nil {
		return Public{}, err
	}
	p, err := s.store.AddProgramme(ctx, Programme{Name: name, FileName: key, DateOfProgramme: in.Date, SeasonID: in.SeasonID})
	if err != nil {
		s.files.Remove(ctx, key)
		return Public{}, fmt.Errorf("failed to add programme: %w", err)
	}
	seasons, err := s.seasonMap(ctx)
	if err != nil {
		return Public{}, err
	}
	return s.public(p, seasons), nil
}

// Delete removes the programme and its file, returning what was deleted.
func (s *Service) Delete(ctx context.Context, id int) (Public, error) {
	ctx, span := tracer.Start(ctx, "programme.Service.Delete")
	defer span.End()
	p, err := s.store.GetProgramme(ctx, Programme{ID: id})
	if err != nil {
		return Public{}, svcerr.FromStore(err, "programme")
	}
	if err = s.store.DeleteProgramme(ctx, p); err != nil {
		return Public{}, fmt.Errorf("failed to delete programme: %w", err)
	}
	s.files.Remove(ctx, p.FileName)
	return s.public(p, nil), nil
}

func (s *Service) Seasons(ctx context.Context) ([]PublicSeason, error) {
	ctx, span := tracer.Start(ctx, "programme.Service.Seasons")
	defer span.End()
	rows, err := s.store.GetSeasons(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list seasons: %w", err)
	}
	out := make([]PublicSeason, 0, len(rows))
	for _, r := range rows {
		out = append(out, PublicSeason{ID: r.ID, Name: r.Season})
	}
	return out, nil
}

func (s *Service) CreateSeason(ctx context.Context, name string) (PublicSeason, error) {
	ctx, span := tracer.Start(ctx, "programme.Service.CreateSeason")
	defer span.End()
	name = strings.TrimSpace(name)
	if name == "" {
		return PublicSeason{}, svcerr.InvalidField("season", "season is required")
	}
	r, err := s.store.AddSeason(ctx, Season{Season: name})
	if err != nil {
		return PublicSeason{}, fmt.Errorf("failed to add season: %w", err)
	}
	return PublicSeason{ID: r.ID, Name: r.Season}, nil
}

func (s *Service) RenameSeason(ctx context.Context, id int, name string) (PublicSeason, error) {
	ctx, span := tracer.Start(ctx, "programme.Service.RenameSeason")
	defer span.End()
	name = strings.TrimSpace(name)
	if name == "" {
		return PublicSeason{}, svcerr.InvalidField("season", "season is required")
	}
	r, err := s.store.GetSeason(ctx, Season{ID: id})
	if err != nil {
		return PublicSeason{}, svcerr.FromStore(err, "season")
	}
	r.Season = name
	if _, err = s.store.EditSeason(ctx, r); err != nil {
		return PublicSeason{}, fmt.Errorf("failed to edit season: %w", err)
	}
	return PublicSeason{ID: r.ID, Name: r.Season}, nil
}

// DeleteSeason unlinks the season's programmes (they are kept) and deletes it.
func (s *Service) DeleteSeason(ctx context.Context, id int) (PublicSeason, error) {
	ctx, span := tracer.Start(ctx, "programme.Service.DeleteSeason")
	defer span.End()
	r, err := s.store.GetSeason(ctx, Season{ID: id})
	if err != nil {
		return PublicSeason{}, svcerr.FromStore(err, "season")
	}
	linked, err := s.store.GetProgrammesSeason(ctx, r)
	if err != nil {
		return PublicSeason{}, fmt.Errorf("failed to list season programmes: %w", err)
	}
	for _, p := range linked {
		p.SeasonID = 0
		if _, err = s.store.EditProgramme(ctx, p); err != nil {
			return PublicSeason{}, fmt.Errorf("failed to unlink programme %d: %w", p.ID, err)
		}
	}
	if err = s.store.DeleteSeason(ctx, r); err != nil {
		return PublicSeason{}, fmt.Errorf("failed to delete season: %w", err)
	}
	return PublicSeason{ID: r.ID, Name: r.Season}, nil
}
