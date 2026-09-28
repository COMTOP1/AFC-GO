package player

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

const uploadCategory = "player"

type store interface {
	GetPlayers(ctx context.Context) ([]Player, error)
	GetPlayersTeam(ctx context.Context, teamParam team.Team) ([]Player, error)
	GetPlayer(ctx context.Context, playerParam Player) (Player, error)
	AddPlayer(ctx context.Context, playerParam Player) (Player, error)
	EditPlayer(ctx context.Context, playerParam Player) (Player, error)
	DeletePlayer(ctx context.Context, playerParam Player) error
}

// TeamGetter reads teams (satisfied by *team.Store).
type TeamGetter interface {
	GetTeam(ctx context.Context, teamParam team.Team) (team.Team, error)
	GetTeams(ctx context.Context) ([]team.Team, error)
}

// Service is the player business logic shared by the API and legacy views.
type Service struct {
	store store
	teams TeamGetter
	files *upload.Files
}

func NewService(store store, teams TeamGetter, files *upload.Files) *Service {
	return &Service{store: store, teams: teams, files: files}
}

func (s *Service) public(p Player, t *team.Team, now time.Time) Public {
	out := Public{ID: p.ID, Name: p.Name, Position: p.Position.String, IsCaptain: p.IsCaptain}
	if p.DateOfBirth.Valid {
		dob := p.DateOfBirth.Time
		out.DateOfBirth = &dob
		if age, ok := Age(dob, now); ok {
			out.Age = &age
		}
	}
	// A player whose team can't be found (e.g. team_id 0 after DetachTeam)
	// is treated as youth: we can't prove otherwise, so fail closed.
	youth := true
	if t != nil {
		out.Team = &TeamRef{ID: t.ID, Name: t.Name, IsYouth: t.IsYouth}
		youth = t.IsYouth
	}
	if PhotoVisible(p, youth, now) {
		out.ImageURL = s.files.URL(p.FileName.String)
	}
	return out
}

func (s *Service) List(ctx context.Context) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "player.Service.List")
	defer span.End()
	rows, err := s.store.GetPlayers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list players: %w", err)
	}
	teamRows, err := s.teams.GetTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}
	byID := make(map[int]team.Team, len(teamRows))
	for _, t := range teamRows {
		byID[t.ID] = t
	}
	now := time.Now()
	out := make([]Public, 0, len(rows))
	for _, p := range rows {
		var t *team.Team
		if found, ok := byID[p.TeamID]; ok {
			t = &found
		}
		out = append(out, s.public(p, t, now))
	}
	return out, nil
}

// Squad lists a team's players for its public page. The caller's t.IsYouth is
// not trusted; the team is re-read so a stale or wrong value can't expose a
// youth team's squad. A youth team's squad is hidden entirely (no names, no
// positions, not just photos), matching the legacy team page. If the team
// can't be re-read, the squad is hidden too: an unknown team can't be proven
// to be non-youth, so we fail closed.
func (s *Service) Squad(ctx context.Context, t team.Team) ([]Member, error) {
	ctx, span := tracer.Start(ctx, "player.Service.Squad")
	defer span.End()
	fresh, teamErr := s.teams.GetTeam(ctx, team.Team{ID: t.ID})
	if teamErr != nil || fresh.IsYouth {
		return []Member{}, nil
	}
	rows, err := s.store.GetPlayersTeam(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("failed to list squad: %w", err)
	}
	now := time.Now()
	out := make([]Member, 0, len(rows))
	for _, p := range rows {
		m := Member{ID: p.ID, Name: p.Name, Position: p.Position.String, IsCaptain: p.IsCaptain}
		if PhotoVisible(p, fresh.IsYouth, now) {
			m.ImageURL = s.files.URL(p.FileName.String)
		}
		out = append(out, m)
	}
	return out, nil
}

// PhotoKey returns the storage key of a player's photo, or NotFound when
// there is none or it must not be shown.
func (s *Service) PhotoKey(ctx context.Context, id int) (string, error) {
	ctx, span := tracer.Start(ctx, "player.Service.PhotoKey")
	defer span.End()
	notFound := svcerr.NotFound("player photo not found", nil)
	p, err := s.store.GetPlayer(ctx, Player{ID: id})
	if err != nil {
		// Same error as "hidden" so a caller can't tell existing IDs apart
		// from missing ones.
		return "", notFound
	}
	t, err := s.teams.GetTeam(ctx, team.Team{ID: p.TeamID})
	if err != nil {
		// Unknown team: we can't prove it isn't a youth team, so hide.
		if !errors.Is(err, sql.ErrNoRows) {
			span.RecordError(err)
		}
		return "", notFound
	}
	if !PhotoVisible(p, t.IsYouth, time.Now()) {
		return "", notFound
	}
	return p.FileName.String, nil
}

func (s *Service) get(ctx context.Context, id int) (Player, error) {
	p, err := s.store.GetPlayer(ctx, Player{ID: id})
	if err != nil {
		return Player{}, svcerr.FromStore(err, "player")
	}
	return p, nil
}

func (s *Service) teamFor(ctx context.Context, id int) (*team.Team, error) {
	t, err := s.teams.GetTeam(ctx, team.Team{ID: id})
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Service) validate(ctx context.Context, p Player) error {
	f := svcerr.Fields{}
	if strings.TrimSpace(p.Name) == "" {
		f.Add("name", "name is required")
	}
	if _, err := s.teamFor(ctx, p.TeamID); err != nil {
		f.Add("teamId", "team does not exist")
	}
	if !p.DateOfBirth.Valid || p.DateOfBirth.Time.IsZero() {
		f.Add("dateOfBirth", "date of birth is required")
	} else if !p.DateOfBirth.Time.Before(time.Now()) {
		f.Add("dateOfBirth", "date of birth must be in the past")
	}
	return f.Err()
}

func (s *Service) Create(ctx context.Context, in CreateInput, image *upload.File) (Public, error) {
	ctx, span := tracer.Start(ctx, "player.Service.Create")
	defer span.End()
	position := strings.TrimSpace(in.Position)
	p := Player{
		Name:        strings.TrimSpace(in.Name),
		Position:    null.NewString(position, position != ""),
		TeamID:      in.TeamID,
		IsCaptain:   in.IsCaptain,
		DateOfBirth: null.NewTime(in.DateOfBirth, !in.DateOfBirth.IsZero()),
	}
	if err := s.validate(ctx, p); err != nil {
		return Public{}, err
	}
	if image != nil {
		key, err := s.files.Save(ctx, image, uploadCategory)
		if err != nil {
			return Public{}, err
		}
		p.FileName = null.StringFrom(key)
	}
	added, err := s.store.AddPlayer(ctx, p)
	if err != nil {
		s.files.Remove(ctx, p.FileName.String)
		return Public{}, fmt.Errorf("failed to add player: %w", err)
	}
	t, _ := s.teamFor(ctx, added.TeamID) //nolint:errcheck // validated above
	return s.public(added, t, time.Now()), nil
}

func (s *Service) Update(ctx context.Context, id int, in UpdateInput, image *upload.File) (Public, error) {
	ctx, span := tracer.Start(ctx, "player.Service.Update")
	defer span.End()
	p, err := s.get(ctx, id)
	if err != nil {
		return Public{}, err
	}
	if in.Name != nil {
		p.Name = strings.TrimSpace(*in.Name)
	}
	if in.Position != nil {
		position := strings.TrimSpace(*in.Position)
		p.Position = null.NewString(position, position != "")
	}
	if in.TeamID != nil {
		p.TeamID = *in.TeamID
	}
	if in.DateOfBirth != nil {
		p.DateOfBirth = null.NewTime(*in.DateOfBirth, !in.DateOfBirth.IsZero())
	}
	if in.IsCaptain != nil {
		p.IsCaptain = *in.IsCaptain
	}
	if err = s.validate(ctx, p); err != nil {
		return Public{}, err
	}

	oldKey, newKey := p.FileName.String, ""
	switch {
	case image != nil:
		if newKey, err = s.files.Save(ctx, image, uploadCategory); err != nil {
			return Public{}, err
		}
		p.FileName = null.StringFrom(newKey)
	case in.RemoveImage:
		p.FileName = null.String{}
	}
	if _, err = s.store.EditPlayer(ctx, p); err != nil {
		s.files.Remove(ctx, newKey)
		return Public{}, fmt.Errorf("failed to edit player: %w", err)
	}
	if oldKey != "" && oldKey != p.FileName.String {
		s.files.Remove(ctx, oldKey)
	}
	t, _ := s.teamFor(ctx, p.TeamID) //nolint:errcheck // validated above
	return s.public(p, t, time.Now()), nil
}

// Delete removes the player and their photo, returning what was deleted.
func (s *Service) Delete(ctx context.Context, id int) (Public, error) {
	ctx, span := tracer.Start(ctx, "player.Service.Delete")
	defer span.End()
	p, err := s.get(ctx, id)
	if err != nil {
		return Public{}, err
	}
	t, _ := s.teamFor(ctx, p.TeamID) //nolint:errcheck // a missing team only affects display
	deleted := s.public(p, t, time.Now())
	if err = s.store.DeletePlayer(ctx, p); err != nil {
		return Public{}, fmt.Errorf("failed to delete player: %w", err)
	}
	s.files.Remove(ctx, p.FileName.String)
	return deleted, nil
}
