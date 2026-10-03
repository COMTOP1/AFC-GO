package setting

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	emailverifier "github.com/AfterShip/email-verifier"

	"github.com/COMTOP1/AFC-GO/server/internal/sanitise"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
)

const (
	infoContentID  = "infoContent"
	displayEmailID = "displayEmail"
)

type store interface {
	GetSetting(ctx context.Context, settingID string) (Setting, error)
	AddSetting(ctx context.Context, settingParam Setting) (Setting, error)
	EditSetting(ctx context.Context, settingParam Setting) (Setting, error)
	DeleteSetting(ctx context.Context, settingID string) error
}

// Service manages the editable site settings.
type Service struct {
	store store
}

func NewService(store store) *Service {
	return &Service{store: store}
}

func (s *Service) get(ctx context.Context, id string) (string, error) {
	v, err := s.store.GetSetting(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get setting %s: %w", id, err)
	}
	return v.SettingText, nil
}

// put inserts or updates a setting.
func (s *Service) put(ctx context.Context, id, text string) error {
	_, err := s.store.GetSetting(ctx, id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		_, err = s.store.AddSetting(ctx, Setting{ID: id, SettingText: text})
	case err == nil:
		_, err = s.store.EditSetting(ctx, Setting{ID: id, SettingText: text})
	}
	if err != nil {
		return fmt.Errorf("failed to save setting %s: %w", id, err)
	}
	return nil
}

// Info returns the "about the club" page HTML ("" when never set).
func (s *Service) Info(ctx context.Context) (string, error) {
	ctx, span := tracer.Start(ctx, "setting.Service.Info")
	defer span.End()
	return s.get(ctx, infoContentID)
}

// SetInfo sanitises and saves the info page HTML, returning what was stored.
func (s *Service) SetInfo(ctx context.Context, html string) (string, error) {
	ctx, span := tracer.Start(ctx, "setting.Service.SetInfo")
	defer span.End()
	safe := sanitise.HTML(html)
	if err := s.put(ctx, infoContentID, safe); err != nil {
		return "", err
	}
	return safe, nil
}

// DisplayEmail returns the public contact email ("" when not set).
func (s *Service) DisplayEmail(ctx context.Context) (string, error) {
	ctx, span := tracer.Start(ctx, "setting.Service.DisplayEmail")
	defer span.End()
	return s.get(ctx, displayEmailID)
}

// SetDisplayEmail validates and saves the public contact email; an empty
// email removes it.
func (s *Service) SetDisplayEmail(ctx context.Context, email string) (string, error) {
	ctx, span := tracer.Start(ctx, "setting.Service.SetDisplayEmail")
	defer span.End()
	email = strings.TrimSpace(email)
	if email == "" {
		if err := s.store.DeleteSetting(ctx, displayEmailID); err != nil {
			return "", fmt.Errorf("failed to delete display email: %w", err)
		}
		return "", nil
	}
	if !emailverifier.IsAddressValid(email) {
		return "", svcerr.InvalidField("email", "email address is not valid")
	}
	if err := s.put(ctx, displayEmailID, email); err != nil {
		return "", err
	}
	return email, nil
}
