package image

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

const uploadCategory = "gallery"

type store interface {
	GetImages(ctx context.Context) ([]Image, error)
	GetImage(ctx context.Context, imageParam Image) (Image, error)
	AddImage(ctx context.Context, imageParam Image) (Image, error)
	DeleteImage(ctx context.Context, imageParam Image) error
}

// Service is the gallery business logic.
type Service struct {
	store store
	files *upload.Files
}

func NewService(store store, files *upload.Files) *Service {
	return &Service{store: store, files: files}
}

func (s *Service) public(i Image) Public {
	return Public{ID: i.ID, Caption: i.Caption.String, ImageURL: s.files.URL(i.FileName)}
}

func (s *Service) List(ctx context.Context) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "image.Service.List")
	defer span.End()
	rows, err := s.store.GetImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list gallery: %w", err)
	}
	out := make([]Public, 0, len(rows))
	for _, i := range rows {
		out = append(out, s.public(i))
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, in CreateInput, photo *upload.File) (Public, error) {
	ctx, span := tracer.Start(ctx, "image.Service.Create")
	defer span.End()
	if photo == nil {
		return Public{}, svcerr.InvalidField("image", "image is required")
	}
	key, err := s.files.Save(ctx, photo, uploadCategory)
	if err != nil {
		return Public{}, err
	}
	caption := strings.TrimSpace(in.Caption)
	i, err := s.store.AddImage(ctx, Image{FileName: key, Caption: null.NewString(caption, caption != "")})
	if err != nil {
		s.files.Remove(ctx, key)
		return Public{}, fmt.Errorf("failed to add gallery image: %w", err)
	}
	return s.public(i), nil
}

// Delete removes the photo and its file, returning what was deleted.
func (s *Service) Delete(ctx context.Context, id int) (Public, error) {
	ctx, span := tracer.Start(ctx, "image.Service.Delete")
	defer span.End()
	i, err := s.store.GetImage(ctx, Image{ID: id})
	if err != nil {
		return Public{}, svcerr.FromStore(err, "gallery image")
	}
	deleted := s.public(i)
	if err = s.store.DeleteImage(ctx, i); err != nil {
		return Public{}, fmt.Errorf("failed to delete gallery image: %w", err)
	}
	s.files.Remove(ctx, i.FileName)
	return deleted, nil
}
