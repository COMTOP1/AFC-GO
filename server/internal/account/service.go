// Package account lets a logged-in user manage their own profile photo.
package account

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

var tracer = otel.Tracer("github.com/COMTOP1/AFC-GO/server/internal/account")

// Store updates a user's photo (satisfied by *user.Store).
type Store interface {
	EditUserImage(ctx context.Context, userParam user.User) error
}

// Service changes the current user's photo.
type Service struct {
	store Store
	files *upload.Files
}

func NewService(store Store, files *upload.Files) *Service {
	return &Service{store: store, files: files}
}

// SetImage replaces u's photo; the old object is deleted after the DB write.
func (s *Service) SetImage(ctx context.Context, u user.User, image *upload.File) (user.User, error) {
	ctx, span := tracer.Start(ctx, "account.Service.SetImage")
	defer span.End()
	if image == nil {
		return user.User{}, svcerr.InvalidField("image", "image is required")
	}
	key, err := s.files.Save(ctx, image, "user")
	if err != nil {
		return user.User{}, err
	}
	old := u.FileName.String
	u.FileName = null.StringFrom(key)
	if err = s.store.EditUserImage(ctx, u); err != nil {
		s.files.Remove(ctx, key)
		return user.User{}, fmt.Errorf("failed to update image: %w", err)
	}
	s.files.Remove(ctx, old)
	return u, nil
}

// RemoveImage clears u's photo.
func (s *Service) RemoveImage(ctx context.Context, u user.User) (user.User, error) {
	ctx, span := tracer.Start(ctx, "account.Service.RemoveImage")
	defer span.End()
	old := u.FileName.String
	u.FileName = null.String{}
	if err := s.store.EditUserImage(ctx, u); err != nil {
		return user.User{}, fmt.Errorf("failed to remove image: %w", err)
	}
	s.files.Remove(ctx, old)
	return u, nil
}
