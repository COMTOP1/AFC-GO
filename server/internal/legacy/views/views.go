package views

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/gorilla/sessions"
	"go.opentelemetry.io/otel"

	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	"github.com/COMTOP1/AFC-GO/server/internal/document"
	"github.com/COMTOP1/AFC-GO/server/internal/image"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/server/internal/legacy/templates"
	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/programme"
	"github.com/COMTOP1/AFC-GO/server/internal/role"
	"github.com/COMTOP1/AFC-GO/server/internal/setting"
	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/visitors"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

var tracer = otel.Tracer("github.com/COMTOP1/AFC-GO/server/internal/legacy/views")

type (
	// Config is what the legacy views still need from the environment.
	Config struct {
		DomainName        string
		SessionCookieName string
		Security          SecurityConfig
	}

	// SecurityConfig is kept as an alias so existing v.conf.Security.* call
	// sites don't change.
	SecurityConfig = auth.PasswordConfig

	// Views encapsulates our view dependencies
	Views struct {
		affiliation    *affiliation.Store
		affiliationSvc *affiliation.Service
		conf           *Config
		cookie         *sessions.CookieStore
		document       *document.Store
		documentSvc    *document.Service
		image          *image.Store
		mailer         *mail.MailerInit
		news           *news.Store
		newsSvc        *news.Service
		player         *player.Store
		programme      *programme.Store
		setting        *setting.Store
		sponsor        *sponsor.Store
		sponsorSvc     *sponsor.Service
		storage        upload.Storage
		team           *team.Store
		template       *templates.Templater
		user           *user.Store
		whatsOn        *whatson.Store
		whatsOnSvc     *whatson.Service

		tokens   *auth.Tokens
		visitors *visitors.Counter
	}

	TemplateHelper struct {
		UserPermissions []role.Role
		ActivePage      string
		Assumed         bool
	}

	// Deps are the shared services the legacy views use.
	Deps struct {
		Conf     *Config
		Sessions *sessions.CookieStore
		Storage  upload.Storage
		Mailer   *mail.MailerInit
		Visitors *visitors.Counter
		Tokens   *auth.Tokens

		Affiliation        *affiliation.Store
		AffiliationService *affiliation.Service
		Document           *document.Store
		DocumentService    *document.Service
		Image              *image.Store
		News               *news.Store
		NewsService        *news.Service
		Player             *player.Store
		Programme          *programme.Store
		Setting            *setting.Store
		Sponsor            *sponsor.Store
		SponsorService     *sponsor.Service
		Team               *team.Store
		User               *user.Store
		WhatsOn            *whatson.Store
		WhatsOnService     *whatson.Service
	}
)

// New builds the legacy views from shared dependencies. It connects to nothing.
func New(d Deps) *Views {
	// InternalContext carries flash messages in the session; user.User is
	// registered by the auth package.
	gob.Register(InternalContext{})
	return &Views{
		affiliation:    d.Affiliation,
		affiliationSvc: d.AffiliationService,
		conf:           d.Conf,
		cookie:         d.Sessions,
		document:       d.Document,
		documentSvc:    d.DocumentService,
		image:          d.Image,
		mailer:         d.Mailer,
		news:           d.News,
		newsSvc:        d.NewsService,
		player:         d.Player,
		programme:      d.Programme,
		setting:        d.Setting,
		sponsor:        d.Sponsor,
		sponsorSvc:     d.SponsorService,
		storage:        d.Storage,
		team:           d.Team,
		template:       templates.NewTemplate(d.Team, d.Storage),
		tokens:         d.Tokens,
		user:           d.User,
		visitors:       d.Visitors,
		whatsOn:        d.WhatsOn,
		whatsOnSvc:     d.WhatsOnService,
	}
}

// GetVisitorCount is the site-wide visitor total shown in the footer.
func (v *Views) GetVisitorCount() int { return v.visitors.Count() }

// SetResetToken, GetResetToken and DeleteResetToken keep the existing view
// call sites unchanged while the tokens live in auth.
func (v *Views) SetResetToken(ctx context.Context, token string, userID int, ttl time.Duration) error {
	return v.tokens.Set(ctx, token, userID, ttl)
}

func (v *Views) GetResetToken(ctx context.Context, token string) (int, bool) {
	return v.tokens.Get(ctx, token)
}

func (v *Views) DeleteResetToken(ctx context.Context, token string) {
	v.tokens.Delete(ctx, token)
}
