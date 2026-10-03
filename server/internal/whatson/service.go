package whatson

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/sanitise"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

const uploadCategory = "whatson"

type store interface {
	GetWhatsOn(ctx context.Context) ([]WhatsOn, error)
	GetWhatsOnFuture(ctx context.Context) ([]WhatsOn, error)
	GetWhatsOnPast(ctx context.Context) ([]WhatsOn, error)
	GetWhatsOnLatest(ctx context.Context) (WhatsOn, error)
	GetWhatsOnArticle(ctx context.Context, whatsOnParam WhatsOn) (WhatsOn, error)
	AddWhatsOn(ctx context.Context, whatsOnParam WhatsOn) (WhatsOn, error)
	EditWhatsOn(ctx context.Context, whatsOnParam WhatsOn) (WhatsOn, error)
	DeleteWhatsOn(ctx context.Context, whatsOnParam WhatsOn) error
}

// Service is the what's-on business logic.
type Service struct {
	store store
	files *upload.Files
}

func NewService(store store, files *upload.Files) *Service {
	return &Service{store: store, files: files}
}

func (s *Service) event(w WhatsOn) Event {
	return Event{
		ID:          w.ID,
		Title:       w.Title,
		Content:     w.Content.String,
		Date:        w.Date,
		DateOfEvent: w.DateOfEvent,
		ImageURL:    s.files.URL(w.FileName.String),
	}
}

// List returns events for period; an empty period means all.
func (s *Service) List(ctx context.Context, period Period) ([]Event, error) {
	ctx, span := tracer.Start(ctx, "whatson.Service.List")
	defer span.End()

	var (
		rows []WhatsOn
		err  error
	)
	switch period {
	case PeriodAll, "":
		rows, err = s.store.GetWhatsOn(ctx)
	case PeriodFuture:
		rows, err = s.store.GetWhatsOnFuture(ctx)
	case PeriodPast:
		rows, err = s.store.GetWhatsOnPast(ctx)
	default:
		return nil, svcerr.InvalidField("period", "period must be all, future or past")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list whats on: %w", err)
	}
	out := make([]Event, 0, len(rows))
	for _, w := range rows {
		out = append(out, s.event(w))
	}
	return out, nil
}

// Next returns the soonest upcoming event; ok is false when there is none.
func (s *Service) Next(ctx context.Context) (Event, bool, error) {
	ctx, span := tracer.Start(ctx, "whatson.Service.Next")
	defer span.End()
	w, err := s.store.GetWhatsOnLatest(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return Event{}, false, nil
	}
	if err != nil {
		return Event{}, false, fmt.Errorf("failed to get next whats on: %w", err)
	}
	return s.event(w), true, nil
}

func (s *Service) Get(ctx context.Context, id int) (Event, error) {
	ctx, span := tracer.Start(ctx, "whatson.Service.Get")
	defer span.End()
	w, err := s.get(ctx, id)
	if err != nil {
		return Event{}, err
	}
	return s.event(w), nil
}

func (s *Service) get(ctx context.Context, id int) (WhatsOn, error) {
	w, err := s.store.GetWhatsOnArticle(ctx, WhatsOn{ID: id})
	if err != nil {
		return WhatsOn{}, svcerr.FromStore(err, "whats on event")
	}
	return w, nil
}

func (s *Service) Create(ctx context.Context, in CreateInput, image *upload.File) (Event, error) {
	ctx, span := tracer.Start(ctx, "whatson.Service.Create")
	defer span.End()

	title := strings.TrimSpace(in.Title)
	fields := svcerr.Fields{}
	if title == "" {
		fields.Add("title", "title is required")
	}
	if in.DateOfEvent.IsZero() {
		fields.Add("dateOfEvent", "date of event is required")
	}
	if err := fields.Err(); err != nil {
		return Event{}, err
	}
	content := sanitise.HTML(in.Content)

	var key string
	if image != nil {
		var err error
		if key, err = s.files.Save(ctx, image, uploadCategory); err != nil {
			return Event{}, err
		}
	}
	w, err := s.store.AddWhatsOn(ctx, WhatsOn{
		Title:       title,
		Content:     null.NewString(content, content != ""),
		FileName:    null.NewString(key, key != ""),
		DateOfEvent: in.DateOfEvent,
	})
	if err != nil {
		s.files.Remove(ctx, key)
		return Event{}, fmt.Errorf("failed to add whats on: %w", err)
	}
	return s.event(w), nil
}

func (s *Service) Update(ctx context.Context, id int, in UpdateInput, image *upload.File) (Event, error) {
	ctx, span := tracer.Start(ctx, "whatson.Service.Update")
	defer span.End()

	w, err := s.get(ctx, id)
	if err != nil {
		return Event{}, err
	}
	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return Event{}, svcerr.InvalidField("title", "title is required")
		}
		w.Title = title
	}
	if in.Content != nil {
		content := sanitise.HTML(*in.Content)
		w.Content = null.NewString(content, content != "")
	}
	if in.DateOfEvent != nil {
		if in.DateOfEvent.IsZero() {
			return Event{}, svcerr.InvalidField("dateOfEvent", "date of event is required")
		}
		w.DateOfEvent = *in.DateOfEvent
	}

	oldKey, newKey := w.FileName.String, ""
	switch {
	case image != nil:
		if newKey, err = s.files.Save(ctx, image, uploadCategory); err != nil {
			return Event{}, err
		}
		w.FileName = null.StringFrom(newKey)
	case in.RemoveImage:
		w.FileName = null.String{}
	}

	if _, err = s.store.EditWhatsOn(ctx, w); err != nil {
		s.files.Remove(ctx, newKey)
		return Event{}, fmt.Errorf("failed to edit whats on: %w", err)
	}
	if oldKey != "" && oldKey != w.FileName.String {
		s.files.Remove(ctx, oldKey)
	}
	return s.event(w), nil
}

// Delete removes the event and its image, returning what was deleted.
func (s *Service) Delete(ctx context.Context, id int) (Event, error) {
	ctx, span := tracer.Start(ctx, "whatson.Service.Delete")
	defer span.End()
	w, err := s.get(ctx, id)
	if err != nil {
		return Event{}, err
	}
	deleted := s.event(w)
	if err = s.store.DeleteWhatsOn(ctx, w); err != nil {
		return Event{}, fmt.Errorf("failed to delete whats on: %w", err)
	}
	s.files.Remove(ctx, w.FileName.String)
	return deleted, nil
}
