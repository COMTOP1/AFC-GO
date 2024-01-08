package templates

import (
	"context"
	"crypto/rand"
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/big"

	"github.com/COMTOP1/AFC-GO/team"
)

// tmpls are the storage of templates in the executable
//
//go:embed *.tmpl
var tmpls embed.FS

type Templater struct {
	Team *team.Store
}

type Template string

const (
	AccountTemplate        Template = "account.tmpl"
	NotFound404Template    Template = "404.tmpl"
	InfoTemplate           Template = "info.tmpl"
	HomeTemplate           Template = "home.tmpl"
	ContactTemplate        Template = "contact.tmpl"
	ResetTemplate          Template = "reset.tmpl"
	ErrorTemplate          Template = "error.tmpl"
	TeamsTemplate          Template = "teams.tmpl"
	TeamTemplate           Template = "team.tmpl"
	NewsTemplate           Template = "news.tmpl"
	NewsArticleTemplate    Template = "newsArticle.tmpl"
	WhatsOnTemplate        Template = "whatson.tmpl"
	WhatsOnArticleTemplate Template = "whatsonArticle.tmpl"
	DocumentsTemplate      Template = "documents.tmpl"
	SponsorsTemplate       Template = "sponsors.tmpl"
	UsersTemplate          Template = "users.tmpl"
	GalleryTemplate        Template = "gallery.tmpl"
	ProgrammesTemplate     Template = "programmes.tmpl"
	PlayersTemplate        Template = "players.tmpl"
	SignupEmailTemplate    Template = "signupEmail.tmpl"
	ResetEmailTemplate     Template = "resetEmail.tmpl"
)

type TemplateType int

const (
	NoNavType TemplateType = iota
	RegularType
)

// NewTemplate returns the template format to be used
func NewTemplate(team *team.Store) *Templater {
	return &Templater{
		Team: team,
	}
}

// String returns the string equivalent of Template
func (t Template) String() string {
	return string(t)
}

func (t *Templater) RenderTemplate(w io.Writer, data interface{}, mainTmpl Template, templateType TemplateType) error {
	var err error

	t1 := template.New("_base.tmpl")

	t1.Funcs(t.getFuncMaps())

	switch templateType {
	case NoNavType:
		t1, err = t1.ParseFS(tmpls, "_base.tmpl", "_topNoNav.tmpl", "_footer.tmpl", mainTmpl.String())
	case RegularType:
		t1, err = t1.ParseFS(tmpls, "_base.tmpl", "_top.tmpl", "_footer.tmpl", "_logoutModal.tmpl", "_loginModal.tmpl", mainTmpl.String())
	default:
		return fmt.Errorf("unable to parse template, invalid type: %d", templateType)
	}

	if err != nil {
		return fmt.Errorf("failed to get templates for template(RenderTemplate): %w", err)
	}

	return t1.Execute(w, data)
}

func (t *Templater) GetEmailTemplate(emailTemplate Template) (*template.Template, error) {
	return template.New(emailTemplate.String()).ParseFS(tmpls, emailTemplate.String())
}

// getFuncMaps returns all the in built functions that templates can use
func (t *Templater) getFuncMaps() template.FuncMap {
	return template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"inc": func(a int) int {
			return a + 1
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"div": func(a, b int) float64 {
			return float64(a) / float64(b)
		},
		"getTeamName": func(teamID int) string {
			t1, err := t.Team.GetTeam(context.Background(), team.Team{ID: teamID})
			if err != nil {
				log.Printf("failed to get team for getTeamName: %+v", err)
				return ""
			}
			return t1.Name
		},
		"randomImgInt": func() int64 {
			nBig, err := rand.Int(rand.Reader, big.NewInt(999999))
			if err != nil {
				panic(err)
			}
			return nBig.Int64()
		},
	}
}
