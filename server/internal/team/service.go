package team

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

const (
	uploadCategory = "team"
	// youthAgeLimit: teams for players under this age are youth teams.
	youthAgeLimit = 19
)

type store interface {
	GetTeams(ctx context.Context) ([]Team, error)
	GetTeamsActive(ctx context.Context) ([]Team, error)
	GetTeam(ctx context.Context, teamParam Team) (Team, error)
	AddTeam(ctx context.Context, teamParam Team) (Team, error)
	EditTeam(ctx context.Context, teamParam Team) (Team, error)
	DeleteTeam(ctx context.Context, teamParam Team) error
}

// Detacher unlinks records that reference a team before it is deleted
// (satisfied by the player, sponsor and user stores).
type Detacher interface {
	DetachTeam(ctx context.Context, teamID int) error
}

// Service is the team business logic shared by the API and legacy views.
type Service struct {
	store     store
	files     *upload.Files
	detachers []Detacher
}

func NewService(store store, files *upload.Files, detachers ...Detacher) *Service {
	return &Service{store: store, files: files, detachers: detachers}
}

func optional(s string) null.String {
	s = strings.TrimSpace(s)
	return null.NewString(s, s != "")
}

// setOptional applies an optional text update: nil leaves dst alone, empty clears it.
func setOptional(dst *null.String, value *string) {
	if value != nil {
		*dst = optional(*value)
	}
}

func (s *Service) public(t Team) Public {
	return Public{
		ID:             t.ID,
		Name:           t.Name,
		Description:    t.Description.String,
		League:         t.League.String,
		Division:       t.Division.String,
		LeagueTableURL: t.LeagueTable.String,
		FixturesURL:    t.Fixtures.String,
		Coach:          t.Coach.String,
		Physio:         t.Physio.String,
		ImageURL:       s.files.URL(t.FileName.String),
		IsActive:       t.IsActive,
		IsYouth:        t.IsYouth,
		Ages:           t.Ages,
	}
}

func validate(t Team) error {
	f := svcerr.Fields{}
	if strings.TrimSpace(t.Name) == "" {
		f.Add("name", "name is required")
	}
	if t.LeagueTable.Valid {
		if _, err := url.ParseRequestURI(t.LeagueTable.String); err != nil {
			f.Add("leagueTable", "league table must be a full URL")
		}
	}
	if t.Fixtures.Valid {
		if _, err := url.ParseRequestURI(t.Fixtures.String); err != nil {
			f.Add("fixtures", "fixtures must be a full URL")
		}
	}
	if t.Ages < 0 {
		f.Add("ages", "ages must be zero or more")
	}
	return f.Err()
}

// List returns active teams, or every team when includeInactive is set
// (logged-in users see inactive teams too).
func (s *Service) List(ctx context.Context, includeInactive bool) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "team.Service.List")
	defer span.End()
	var (
		rows []Team
		err  error
	)
	if includeInactive {
		rows, err = s.store.GetTeams(ctx)
	} else {
		rows, err = s.store.GetTeamsActive(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}
	out := make([]Public, 0, len(rows))
	for _, t := range rows {
		out = append(out, s.public(t))
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, id int) (Public, error) {
	ctx, span := tracer.Start(ctx, "team.Service.Get")
	defer span.End()
	t, err := s.get(ctx, id)
	if err != nil {
		return Public{}, err
	}
	return s.public(t), nil
}

func (s *Service) get(ctx context.Context, id int) (Team, error) {
	t, err := s.store.GetTeam(ctx, Team{ID: id})
	if err != nil {
		return Team{}, svcerr.FromStore(err, "team")
	}
	return t, nil
}

func (s *Service) Create(ctx context.Context, in CreateInput, image *upload.File) (Public, error) {
	ctx, span := tracer.Start(ctx, "team.Service.Create")
	defer span.End()

	t := Team{
		Name:        strings.TrimSpace(in.Name),
		Description: optional(in.Description),
		League:      optional(in.League),
		Division:    optional(in.Division),
		LeagueTable: optional(in.LeagueTable),
		Fixtures:    optional(in.Fixtures),
		Coach:       optional(in.Coach),
		Physio:      optional(in.Physio),
		IsActive:    in.IsActive,
		IsYouth:     in.IsYouth || in.Ages < youthAgeLimit,
		Ages:        in.Ages,
	}
	if err := validate(t); err != nil {
		return Public{}, err
	}
	if image != nil {
		key, err := s.files.Save(ctx, image, uploadCategory)
		if err != nil {
			return Public{}, err
		}
		t.FileName = null.StringFrom(key)
	}
	added, err := s.store.AddTeam(ctx, t)
	if err != nil {
		s.files.Remove(ctx, t.FileName.String)
		return Public{}, fmt.Errorf("failed to add team: %w", err)
	}
	return s.public(added), nil
}

func (s *Service) Update(ctx context.Context, id int, in UpdateInput, image *upload.File) (Public, error) {
	ctx, span := tracer.Start(ctx, "team.Service.Update")
	defer span.End()

	t, err := s.get(ctx, id)
	if err != nil {
		return Public{}, err
	}
	if in.Name != nil {
		t.Name = strings.TrimSpace(*in.Name)
	}
	setOptional(&t.Description, in.Description)
	setOptional(&t.League, in.League)
	setOptional(&t.Division, in.Division)
	setOptional(&t.LeagueTable, in.LeagueTable)
	setOptional(&t.Fixtures, in.Fixtures)
	setOptional(&t.Coach, in.Coach)
	setOptional(&t.Physio, in.Physio)
	if in.IsActive != nil {
		t.IsActive = *in.IsActive
	}
	if in.IsYouth != nil {
		t.IsYouth = *in.IsYouth
	}
	if in.Ages != nil {
		t.Ages = *in.Ages
	}
	if t.Ages < youthAgeLimit {
		t.IsYouth = true
	}
	if err = validate(t); err != nil {
		return Public{}, err
	}

	oldKey, newKey := t.FileName.String, ""
	switch {
	case image != nil:
		if newKey, err = s.files.Save(ctx, image, uploadCategory); err != nil {
			return Public{}, err
		}
		t.FileName = null.StringFrom(newKey)
	case in.RemoveImage:
		t.FileName = null.String{}
	}
	if _, err = s.store.EditTeam(ctx, t); err != nil {
		s.files.Remove(ctx, newKey)
		return Public{}, fmt.Errorf("failed to edit team: %w", err)
	}
	if oldKey != "" && oldKey != t.FileName.String {
		s.files.Remove(ctx, oldKey)
	}
	return s.public(t), nil
}

// Delete unlinks the team's players, sponsors and managers, then deletes the
// team and its image. If any unlink fails, the team is kept.
func (s *Service) Delete(ctx context.Context, id int) (Public, error) {
	ctx, span := tracer.Start(ctx, "team.Service.Delete")
	defer span.End()

	t, err := s.get(ctx, id)
	if err != nil {
		return Public{}, err
	}
	for _, d := range s.detachers {
		if err = d.DetachTeam(ctx, t.ID); err != nil {
			return Public{}, fmt.Errorf("failed to detach team %d: %w", t.ID, err)
		}
	}
	if err = s.store.DeleteTeam(ctx, t); err != nil {
		return Public{}, fmt.Errorf("failed to delete team: %w", err)
	}
	s.files.Remove(ctx, t.FileName.String)
	return s.public(t), nil
}
