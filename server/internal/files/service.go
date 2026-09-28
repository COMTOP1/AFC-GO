// Package files redirects to stored files by resource kind and ID, applying
// per-kind access rules (player photos respect the safeguarding rule).
package files

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"

	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
	"github.com/COMTOP1/AFC-GO/server/internal/document"
	"github.com/COMTOP1/AFC-GO/server/internal/image"
	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/programme"
	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

var tracer = otel.Tracer("github.com/COMTOP1/AFC-GO/server/internal/files")

// Resolver returns the storage key for a resource ID ("" when it has no file).
type Resolver func(ctx context.Context, id int) (string, error)

// PhotoKeyer resolves player photos, enforcing the safeguarding rule
// (satisfied by *player.Service).
type PhotoKeyer interface {
	PhotoKey(ctx context.Context, id int) (string, error)
}

// Sources are the lookups behind each kind (satisfied by the domain stores).
type Sources struct {
	Affiliation interface {
		GetAffiliation(ctx context.Context, a affiliation.Affiliation) (affiliation.Affiliation, error)
	}
	Document interface {
		GetDocument(ctx context.Context, d document.Document) (document.Document, error)
	}
	Image interface {
		GetImage(ctx context.Context, i image.Image) (image.Image, error)
	}
	News interface {
		GetNewsArticle(ctx context.Context, n news.News) (news.News, error)
	}
	Programme interface {
		GetProgramme(ctx context.Context, p programme.Programme) (programme.Programme, error)
	}
	Sponsor interface {
		GetSponsor(ctx context.Context, s sponsor.Sponsor) (sponsor.Sponsor, error)
	}
	Team interface {
		GetTeam(ctx context.Context, t team.Team) (team.Team, error)
	}
	User interface {
		GetUser(ctx context.Context, u user.User) (user.User, error)
	}
	WhatsOn interface {
		GetWhatsOnArticle(ctx context.Context, w whatson.WhatsOn) (whatson.WhatsOn, error)
	}
	Players PhotoKeyer
}

// Resolvers builds the kind → lookup table.
func Resolvers(src Sources) map[string]Resolver {
	return map[string]Resolver{
		"affiliation": func(ctx context.Context, id int) (string, error) {
			a, err := src.Affiliation.GetAffiliation(ctx, affiliation.Affiliation{ID: id})
			return a.FileName.String, svcerr.FromStore(err, "affiliation")
		},
		"document": func(ctx context.Context, id int) (string, error) {
			d, err := src.Document.GetDocument(ctx, document.Document{ID: id})
			return d.FileName, svcerr.FromStore(err, "document")
		},
		"gallery": func(ctx context.Context, id int) (string, error) {
			i, err := src.Image.GetImage(ctx, image.Image{ID: id})
			return i.FileName, svcerr.FromStore(err, "gallery image")
		},
		"news": func(ctx context.Context, id int) (string, error) {
			n, err := src.News.GetNewsArticle(ctx, news.News{ID: id})
			return n.FileName.String, svcerr.FromStore(err, "news article")
		},
		"player": src.Players.PhotoKey,
		"programme": func(ctx context.Context, id int) (string, error) {
			p, err := src.Programme.GetProgramme(ctx, programme.Programme{ID: id})
			return p.FileName, svcerr.FromStore(err, "programme")
		},
		"sponsor": func(ctx context.Context, id int) (string, error) {
			s, err := src.Sponsor.GetSponsor(ctx, sponsor.Sponsor{ID: id})
			return s.FileName.String, svcerr.FromStore(err, "sponsor")
		},
		"team": func(ctx context.Context, id int) (string, error) {
			t, err := src.Team.GetTeam(ctx, team.Team{ID: id})
			return t.FileName.String, svcerr.FromStore(err, "team")
		},
		"user": func(ctx context.Context, id int) (string, error) {
			u, err := src.User.GetUser(ctx, user.User{ID: id})
			if err == nil && u.ID != id {
				// GetUser matches email OR id; never hand out someone else's photo.
				return "", svcerr.NotFound("user not found", nil)
			}
			return u.FileName.String, svcerr.FromStore(err, "user")
		},
		"whatson": func(ctx context.Context, id int) (string, error) {
			w, err := src.WhatsOn.GetWhatsOnArticle(ctx, whatson.WhatsOn{ID: id})
			return w.FileName.String, svcerr.FromStore(err, "whats on event")
		},
	}
}

var legacyKinds = map[string]string{
	"a": "affiliation", "d": "document", "g": "gallery", "l": "player", "n": "news",
	"p": "programme", "s": "sponsor", "t": "team", "u": "user", "w": "whatson",
}

// LegacyKind maps the old /download?s=<code> codes to kinds.
func LegacyKind(code string) (string, bool) {
	kind, ok := legacyKinds[code]
	return kind, ok
}

// Service resolves files to public URLs.
type Service struct {
	files     *upload.Files
	resolvers map[string]Resolver
}

func NewService(files *upload.Files, resolvers map[string]Resolver) *Service {
	return &Service{files: files, resolvers: resolvers}
}

// URL returns the CDN URL for the file behind (kind, id), or NotFound.
func (s *Service) URL(ctx context.Context, kind string, id int) (string, error) {
	ctx, span := tracer.Start(ctx, "files.Service.URL")
	defer span.End()
	resolve, ok := s.resolvers[kind]
	if !ok {
		return "", svcerr.NotFound("unknown file kind", nil)
	}
	key, err := resolve(ctx, id)
	if err != nil {
		// Resolvers built by Resolvers() already return a *svcerr.Error (or
		// nil); this only converts a resolver that returned a raw store
		// error (e.g. sql.ErrNoRows) without translating it first.
		if _, ok := svcerr.As(err); !ok {
			err = svcerr.FromStore(err, kind)
		}
		return "", err
	}
	if key == "" {
		return "", svcerr.NotFound(kind+" has no file", nil)
	}
	exists, err := s.files.Exists(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to check %s file: %w", kind, err)
	}
	if !exists {
		return "", svcerr.NotFound(kind+" file not found", nil)
	}
	return s.files.URL(key), nil
}
