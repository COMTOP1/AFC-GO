package site

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"

	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

var tracer = otel.Tracer("github.com/COMTOP1/AFC-GO/server/internal/site")

type (
	NewsSource interface {
		Latest(ctx context.Context) (news.Article, bool, error)
	}
	WhatsOnSource interface {
		Next(ctx context.Context) (whatson.Event, bool, error)
	}
	SponsorSource interface {
		ListMinimal(ctx context.Context) ([]sponsor.Public, error)
		ForTeam(ctx context.Context, teamID int) ([]sponsor.Public, error)
	}
	AffiliationSource interface {
		ListMinimal(ctx context.Context) ([]affiliation.Public, error)
	}
	TeamSource interface {
		List(ctx context.Context, includeInactive bool) ([]team.Public, error)
		Get(ctx context.Context, id int) (team.Public, error)
	}
	SquadSource interface {
		Squad(ctx context.Context, t team.Team) ([]player.Member, error)
	}
	SettingSource interface {
		DisplayEmail(ctx context.Context) (string, error)
	}
	UserSource interface {
		GetUsersContact(ctx context.Context) ([]user.User, error)
		GetUsersManagersTeam(ctx context.Context, teamParam team.Team) ([]user.User, error)
	}
	CountSource interface {
		Count() int
	}
)

// Deps are the services site reads from.
type Deps struct {
	News         NewsSource
	WhatsOn      WhatsOnSource
	Sponsors     SponsorSource
	Affiliations AffiliationSource
	Teams        TeamSource
	Players      SquadSource
	Settings     SettingSource
	Users        UserSource
	Visitors     CountSource
	Files        *upload.Files
	Version      string
}

// Service builds the cross-domain page data.
type Service struct {
	d Deps
}

func NewService(d Deps) *Service {
	return &Service{d: d}
}

// Site returns the layout data shared by every page.
func (s *Service) Site(ctx context.Context) (Info, error) {
	ctx, span := tracer.Start(ctx, "site.Service.Site")
	defer span.End()
	teams, err := s.d.Teams.List(ctx, false)
	if err != nil {
		return Info{}, err
	}
	email, err := s.d.Settings.DisplayEmail(ctx)
	if err != nil {
		return Info{}, err
	}
	return Info{
		Year:         time.Now().Year(),
		VisitorCount: s.d.Visitors.Count(),
		DisplayEmail: email,
		Version:      s.d.Version,
		Teams:        teams,
	}, nil
}

// Home returns the home page; failed panels are logged and omitted.
func (s *Service) Home(ctx context.Context) Home {
	ctx, span := tracer.Start(ctx, "site.Service.Home")
	defer span.End()
	home := Home{Sponsors: []sponsor.Public{}, Affiliations: []affiliation.Public{}}
	soft := func(what string, err error) bool {
		if err != nil {
			span.RecordError(err)
			slog.InfoContext(ctx, fmt.Sprintf("home: failed to load %s: %+v", what, err))
			return false
		}
		return true
	}
	if a, ok, err := s.d.News.Latest(ctx); soft("latest news", err) && ok {
		home.LatestNews = &a
	}
	if e, ok, err := s.d.WhatsOn.Next(ctx); soft("next event", err) && ok {
		home.NextEvent = &e
	}
	if sponsors, err := s.d.Sponsors.ListMinimal(ctx); soft("sponsors", err) {
		home.Sponsors = sponsors
	}
	if affiliations, err := s.d.Affiliations.ListMinimal(ctx); soft("affiliations", err) {
		home.Affiliations = affiliations
	}
	return home
}

// Contact returns the club officials and public contact email.
func (s *Service) Contact(ctx context.Context) (Contact, error) {
	ctx, span := tracer.Start(ctx, "site.Service.Contact")
	defer span.End()
	people, err := s.d.Users.GetUsersContact(ctx)
	if err != nil {
		return Contact{}, fmt.Errorf("failed to get contacts: %w", err)
	}
	email, err := s.d.Settings.DisplayEmail(ctx)
	if err != nil {
		return Contact{}, err
	}
	out := Contact{DisplayEmail: email, People: make([]ContactPerson, 0, len(people))}
	for _, u := range people {
		out.People = append(out.People, ContactPerson{
			ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role.String(), ImageURL: s.d.Files.URL(u.FileName.String),
		})
	}
	return out, nil
}

// Team returns the public team page.
func (s *Service) Team(ctx context.Context, id int) (TeamDetail, error) {
	ctx, span := tracer.Start(ctx, "site.Service.Team")
	defer span.End()
	t, err := s.d.Teams.Get(ctx, id)
	if err != nil {
		return TeamDetail{}, err
	}
	ref := team.Team{ID: t.ID, IsYouth: t.IsYouth}
	managers, err := s.d.Users.GetUsersManagersTeam(ctx, ref)
	if err != nil {
		return TeamDetail{}, fmt.Errorf("failed to get managers: %w", err)
	}
	sponsors, err := s.d.Sponsors.ForTeam(ctx, t.ID)
	if err != nil {
		return TeamDetail{}, err
	}
	players, err := s.d.Players.Squad(ctx, ref)
	if err != nil {
		return TeamDetail{}, err
	}
	out := TeamDetail{Team: t, Managers: make([]Manager, 0, len(managers)), Sponsors: sponsors, Players: players}
	for _, m := range managers {
		out.Managers = append(out.Managers, Manager{Name: m.Name, Email: m.Email})
	}
	return out, nil
}
