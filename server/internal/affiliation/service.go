package affiliation

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

const uploadCategory = "affiliation"

type store interface {
	GetAffiliations(ctx context.Context) ([]Affiliation, error)
	GetAffiliationsMinimal(ctx context.Context) ([]Affiliation, error)
	GetAffiliation(ctx context.Context, affiliationParam Affiliation) (Affiliation, error)
	AddAffiliation(ctx context.Context, affiliationParam Affiliation) (Affiliation, error)
	DeleteAffiliation(ctx context.Context, affiliationParam Affiliation) error
}

// Service is the affiliation business logic.
type Service struct {
	store store
	files *upload.Files
}

func NewService(store store, files *upload.Files) *Service {
	return &Service{store: store, files: files}
}

func (s *Service) public(a Affiliation) Public {
	return Public{ID: a.ID, Name: a.Name, Website: a.Website.String, ImageURL: s.files.URL(a.FileName.String)}
}

func (s *Service) many(rows []Affiliation) []Public {
	out := make([]Public, 0, len(rows))
	for _, r := range rows {
		out = append(out, s.public(r))
	}
	return out
}

func (s *Service) List(ctx context.Context) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "affiliation.Service.List")
	defer span.End()
	rows, err := s.store.GetAffiliations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list affiliations: %w", err)
	}
	return s.many(rows), nil
}

// ListMinimal is the lighter list shown on the home page.
func (s *Service) ListMinimal(ctx context.Context) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "affiliation.Service.ListMinimal")
	defer span.End()
	rows, err := s.store.GetAffiliationsMinimal(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list affiliations: %w", err)
	}
	return s.many(rows), nil
}

func (s *Service) Create(ctx context.Context, in CreateInput, image *upload.File) (Public, error) {
	ctx, span := tracer.Start(ctx, "affiliation.Service.Create")
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
	if err := fields.Err(); err != nil {
		return Public{}, err
	}

	key, err := s.files.Save(ctx, image, uploadCategory)
	if err != nil {
		return Public{}, err
	}
	a, err := s.store.AddAffiliation(ctx, Affiliation{
		Name:     name,
		Website:  null.NewString(in.Website, in.Website != ""),
		FileName: null.StringFrom(key),
	})
	if err != nil {
		s.files.Remove(ctx, key)
		return Public{}, fmt.Errorf("failed to add affiliation: %w", err)
	}
	return s.public(a), nil
}

// Delete removes the affiliation and its logo, returning what was deleted.
func (s *Service) Delete(ctx context.Context, id int) (Public, error) {
	ctx, span := tracer.Start(ctx, "affiliation.Service.Delete")
	defer span.End()
	a, err := s.store.GetAffiliation(ctx, Affiliation{ID: id})
	if err != nil {
		return Public{}, svcerr.FromStore(err, "affiliation")
	}
	deleted := s.public(a)
	if err = s.store.DeleteAffiliation(ctx, a); err != nil {
		return Public{}, fmt.Errorf("failed to delete affiliation: %w", err)
	}
	s.files.Remove(ctx, a.FileName.String)
	return deleted, nil
}
