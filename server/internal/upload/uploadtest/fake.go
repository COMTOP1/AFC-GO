// Package uploadtest provides an in-memory upload.Storage for tests.
package uploadtest

import (
	"context"
	"io"
	"strings"
	"sync"

	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

// Storage is an in-memory object store.
type Storage struct {
	mu      sync.Mutex
	Objects map[string]string
	Deleted []string
	PutErr  error
}

func New() *Storage {
	return &Storage{Objects: map[string]string{}}
}

func (s *Storage) Put(_ context.Context, key string, body io.Reader, _ int64, _ string) error {
	if s.PutErr != nil {
		return s.PutErr
	}
	b, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Objects[key] = string(b)
	return nil
}

func (s *Storage) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Deleted = append(s.Deleted, key)
	delete(s.Objects, key)
	return nil
}

func (s *Storage) Exists(_ context.Context, key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.Objects[key]
	return ok, nil
}

func (s *Storage) PublicURL(key string) string {
	return "https://cdn.test/" + key
}

// File builds an upload.File with the given contents.
func File(name, contentType, body string) *upload.File {
	return &upload.File{
		Name:        name,
		ContentType: contentType,
		Size:        int64(len(body)),
		Open:        func() (io.ReadCloser, error) { return io.NopCloser(strings.NewReader(body)), nil },
	}
}
