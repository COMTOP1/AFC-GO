package sponsor

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

const uploadCategory = "sponsor"

type store interface {
	GetSponsors(ctx context.Context) ([]Sponsor, error)
	GetSponsorsMinimal(ctx context.Context) ([]Sponsor, error)
	GetSponsorsTeam(ctx context.Context, teamParam team.Team) ([]Sponsor, error)
	GetSponsor(ctx context.Context, sponsorParam Sponsor) (Sponsor, error)
	AddSponsor(ctx context.Context, sponsorParam Sponsor) (Sponsor, error)
	DeleteSponsor(ctx context.Context, sponsorParam Sponsor) error
}

// TeamGetter checks team IDs (satisfied by *team.Store).
type TeamGetter interface {
	GetTeam(ctx context.Context, teamParam team.Team) (team.Team, error)
}

// Service is the sponsor business logic shared by the API and legacy views.
type Service struct {
	store store
	teams TeamGetter
	files *upload.Files
}

func NewService(store store, teams TeamGetter, files *upload.Files) *Service {
	return &Service{store: store, teams: teams, files: files}
}

func (s *Service) sponsor(x Sponsor) Public {
	return Public{
		ID:       x.ID,
		Name:     x.Name,
		Website:  x.Website.String,
		Purpose:  x.Purpose.String,
		Team:     x.TeamID,
		ImageURL: s.files.URL(x.FileName.String),
	}
}

func (s *Service) many(rows []Sponsor) []Public {
	out := make([]Public, 0, len(rows))
	for _, r := range rows {
		out = append(out, s.sponsor(r))
	}
	return out
}

func (s *Service) List(ctx context.Context) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "sponsor.Service.List")
	defer span.End()
	rows, err := s.store.GetSponsors(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list sponsors: %w", err)
	}
	return s.many(rows), nil
}

// ListMinimal is the lighter list shown on the home page.
func (s *Service) ListMinimal(ctx context.Context) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "sponsor.Service.ListMinimal")
	defer span.End()
	rows, err := s.store.GetSponsorsMinimal(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list sponsors: %w", err)
	}
	return s.many(rows), nil
}

// ForTeam lists the sponsors attached to one team.
func (s *Service) ForTeam(ctx context.Context, teamID int) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "sponsor.Service.ForTeam")
	defer span.End()
	rows, err := s.store.GetSponsorsTeam(ctx, team.Team{ID: teamID})
	if err != nil {
		return nil, fmt.Errorf("failed to list team sponsors: %w", err)
	}
	return s.many(rows), nil
}

func (s *Service) Create(ctx context.Context, in CreateInput, image *upload.File) (Public, error) {
	ctx, span := tracer.Start(ctx, "sponsor.Service.Create")
	defer span.End()

	name := strings.TrimSpace(in.Name)
	fields := svcerr.Fields{}
	if name == "" {
		fields.Add("name", "name is required")
	}
	if image == nil {
		fields.Add("image", "image is required")
	}
	if in.Website != "" {
		if _, err := url.ParseRequestURI(in.Website); err != nil {
			fields.Add("website", "website must be a full URL, e.g. https://example.com")
		}
	}
	if msg := s.checkTeam(ctx, in.Team); msg != "" {
		fields.Add("team", msg)
	}
	if err := fields.Err(); err != nil {
		return Public{}, err
	}

	key, err := s.files.Save(ctx, image, uploadCategory)
	if err != nil {
		return Public{}, err
	}
	x, err := s.store.AddSponsor(ctx, Sponsor{
		Name:     name,
		Website:  null.NewString(in.Website, in.Website != ""),
		Purpose:  null.NewString(in.Purpose, in.Purpose != ""),
		FileName: null.StringFrom(key),
		TeamID:   in.Team,
	})
	if err != nil {
		s.files.Remove(ctx, key)
		return Public{}, fmt.Errorf("failed to add sponsor: %w", err)
	}
	return s.sponsor(x), nil
}

// checkTeam returns a validation message, or "" when value is acceptable.
func (s *Service) checkTeam(ctx context.Context, value string) string {
	switch value {
	case "", "A", "O", "Y":
		return ""
	}
	id, err := strconv.Atoi(value)
	if err != nil {
		return `team must be "A", "O", "Y" or a team ID`
	}
	if _, err = s.teams.GetTeam(ctx, team.Team{ID: id}); err != nil {
		return "team does not exist"
	}
	return ""
}

// Delete removes the sponsor and its logo, returning what was deleted.
func (s *Service) Delete(ctx context.Context, id int) (Public, error) {
	ctx, span := tracer.Start(ctx, "sponsor.Service.Delete")
	defer span.End()
	x, err := s.store.GetSponsor(ctx, Sponsor{ID: id})
	if err != nil {
		return Public{}, svcerr.FromStore(err, "sponsor")
	}
	deleted := s.sponsor(x)
	if err = s.store.DeleteSponsor(ctx, x); err != nil {
		return Public{}, fmt.Errorf("failed to delete sponsor: %w", err)
	}
	s.files.Remove(ctx, x.FileName.String)
	return deleted, nil
}
