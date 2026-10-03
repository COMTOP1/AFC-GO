package news

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

const uploadCategory = "news"

type store interface {
	GetNews(ctx context.Context) ([]News, error)
	GetNewsLatest(ctx context.Context) (News, error)
	GetNewsArticle(ctx context.Context, newsParam News) (News, error)
	AddNews(ctx context.Context, newsParam News) (News, error)
	EditNews(ctx context.Context, newsParam News) (News, error)
	DeleteNews(ctx context.Context, newsParam News) error
}

// Service is the news business logic.
type Service struct {
	store store
	files *upload.Files
}

func NewService(store store, files *upload.Files) *Service {
	return &Service{store: store, files: files}
}

func (s *Service) article(n News) Article {
	return Article{
		ID:       n.ID,
		Title:    n.Title,
		Content:  n.Content.String,
		Date:     n.Date,
		ImageURL: s.files.URL(n.FileName.String),
	}
}

func (s *Service) List(ctx context.Context) ([]Article, error) {
	ctx, span := tracer.Start(ctx, "news.Service.List")
	defer span.End()
	rows, err := s.store.GetNews(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list news: %w", err)
	}
	out := make([]Article, 0, len(rows))
	for _, n := range rows {
		out = append(out, s.article(n))
	}
	return out, nil
}

// Latest returns the newest article; ok is false when there are none.
func (s *Service) Latest(ctx context.Context) (Article, bool, error) {
	ctx, span := tracer.Start(ctx, "news.Service.Latest")
	defer span.End()
	n, err := s.store.GetNewsLatest(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return Article{}, false, nil
	}
	if err != nil {
		return Article{}, false, fmt.Errorf("failed to get latest news: %w", err)
	}
	return s.article(n), true, nil
}

func (s *Service) Get(ctx context.Context, id int) (Article, error) {
	ctx, span := tracer.Start(ctx, "news.Service.Get")
	defer span.End()
	n, err := s.get(ctx, id)
	if err != nil {
		return Article{}, err
	}
	return s.article(n), nil
}

func (s *Service) get(ctx context.Context, id int) (News, error) {
	n, err := s.store.GetNewsArticle(ctx, News{ID: id})
	if err != nil {
		return News{}, svcerr.FromStore(err, "news article")
	}
	return n, nil
}

func (s *Service) Create(ctx context.Context, in CreateInput, image *upload.File) (Article, error) {
	ctx, span := tracer.Start(ctx, "news.Service.Create")
	defer span.End()

	title := strings.TrimSpace(in.Title)
	if title == "" {
		return Article{}, svcerr.InvalidField("title", "title is required")
	}
	content := sanitise.HTML(in.Content)

	var key string
	if image != nil {
		var err error
		if key, err = s.files.Save(ctx, image, uploadCategory); err != nil {
			return Article{}, err
		}
	}

	n, err := s.store.AddNews(ctx, News{
		Title:    title,
		Content:  null.NewString(content, content != ""),
		FileName: null.NewString(key, key != ""),
	})
	if err != nil {
		s.files.Remove(ctx, key)
		return Article{}, fmt.Errorf("failed to add news: %w", err)
	}
	return s.article(n), nil
}

func (s *Service) Update(ctx context.Context, id int, in UpdateInput, image *upload.File) (Article, error) {
	ctx, span := tracer.Start(ctx, "news.Service.Update")
	defer span.End()

	n, err := s.get(ctx, id)
	if err != nil {
		return Article{}, err
	}
	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return Article{}, svcerr.InvalidField("title", "title is required")
		}
		n.Title = title
	}
	if in.Content != nil {
		content := sanitise.HTML(*in.Content)
		n.Content = null.NewString(content, content != "")
	}

	oldKey, newKey := n.FileName.String, ""
	switch {
	case image != nil:
		if newKey, err = s.files.Save(ctx, image, uploadCategory); err != nil {
			return Article{}, err
		}
		n.FileName = null.StringFrom(newKey)
	case in.RemoveImage:
		n.FileName = null.String{}
	}

	if _, err = s.store.EditNews(ctx, n); err != nil {
		s.files.Remove(ctx, newKey)
		return Article{}, fmt.Errorf("failed to edit news: %w", err)
	}
	if oldKey != "" && oldKey != n.FileName.String {
		s.files.Remove(ctx, oldKey)
	}
	return s.article(n), nil
}

// Delete removes the article and its image, returning what was deleted.
func (s *Service) Delete(ctx context.Context, id int) (Article, error) {
	ctx, span := tracer.Start(ctx, "news.Service.Delete")
	defer span.End()

	n, err := s.get(ctx, id)
	if err != nil {
		return Article{}, err
	}
	deleted := s.article(n)
	if err = s.store.DeleteNews(ctx, n); err != nil {
		return Article{}, fmt.Errorf("failed to delete news: %w", err)
	}
	s.files.Remove(ctx, n.FileName.String)
	return deleted, nil
}
