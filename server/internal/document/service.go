package document

import (
	"context"
	"fmt"
	"strings"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

const uploadCategory = "document"

type store interface {
	GetDocuments(ctx context.Context) ([]Document, error)
	GetDocument(ctx context.Context, documentParam Document) (Document, error)
	AddDocument(ctx context.Context, documentParam Document) (Document, error)
	DeleteDocument(ctx context.Context, documentParam Document) error
}

// Service is the document business logic.
type Service struct {
	store store
	files *upload.Files
}

func NewService(store store, files *upload.Files) *Service {
	return &Service{store: store, files: files}
}

func (s *Service) public(d Document) Public {
	return Public{ID: d.ID, Name: d.Name, FileURL: s.files.URL(d.FileName)}
}

func (s *Service) List(ctx context.Context) ([]Public, error) {
	ctx, span := tracer.Start(ctx, "document.Service.List")
	defer span.End()
	rows, err := s.store.GetDocuments(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	out := make([]Public, 0, len(rows))
	for _, d := range rows {
		out = append(out, s.public(d))
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, in CreateInput, file *upload.File) (Public, error) {
	ctx, span := tracer.Start(ctx, "document.Service.Create")
	defer span.End()

	name := strings.TrimSpace(in.Name)
	fields := svcerr.Fields{}
	if name == "" {
		fields.Add("name", "name is required")
	}
	if file == nil {
		fields.Add("file", "file is required")
	}
	if err := fields.Err(); err != nil {
		return Public{}, err
	}
	key, err := s.files.Save(ctx, file, uploadCategory)
	if err != nil {
		return Public{}, err
	}
	d, err := s.store.AddDocument(ctx, Document{Name: name, FileName: key})
	if err != nil {
		s.files.Remove(ctx, key)
		return Public{}, fmt.Errorf("failed to add document: %w", err)
	}
	return s.public(d), nil
}

// Delete removes the document and its file, returning what was deleted.
func (s *Service) Delete(ctx context.Context, id int) (Public, error) {
	ctx, span := tracer.Start(ctx, "document.Service.Delete")
	defer span.End()
	d, err := s.store.GetDocument(ctx, Document{ID: id})
	if err != nil {
		return Public{}, svcerr.FromStore(err, "document")
	}
	deleted := s.public(d)
	if err = s.store.DeleteDocument(ctx, d); err != nil {
		return Public{}, fmt.Errorf("failed to delete document: %w", err)
	}
	s.files.Remove(ctx, d.FileName)
	return deleted, nil
}
