package templates

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"time"

	role1 "github.com/COMTOP1/AFC-GO/infrastructure/role"
	"github.com/COMTOP1/AFC-GO/role"
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
	NewsTemplate           Template = "news.tmpl"
	NewsArticleTemplate    Template = "newsArticle.tmpl"
	WhatsOnTemplate        Template = "whatson.tmpl"
	WhatsOnArticleTemplate Template = "whatsonArticle.tmpl"
	DocumentsTemplate      Template = "documents.tmpl"
	SponsorsTemplate       Template = "sponsors.tmpl"
	UsersTemplate          Template = "users.tmpl"
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
		t1, err = t1.ParseFS(tmpls, "_base.tmpl", "_bodyNoNavs.tmpl", "_head.tmpl", "_footer.tmpl", mainTmpl.String())
	case RegularType:
		t1, err = t1.ParseFS(tmpls, "_base.tmpl", "_top.tmpl", "_footer.tmpl", "_logoutModal.tmpl", "_loginModal.tmpl" /*"_navbar.tmpl", "_sidebar.tmpl",*/, mainTmpl.String())
	default:
		return fmt.Errorf("unable to parse template, invalid type: %d", templateType)
	}

	if err != nil {
		log.Printf("failed to get templates for template(RenderTemplate): %+v", err)
		return err
	}

	return t1.Execute(w, data)
}

// getFuncMaps returns all the in built functions that templates can use
func (t *Templater) getFuncMaps() template.FuncMap {
	return template.FuncMap{
		"thisYear": func() int {
			return time.Now().Year()
		},
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
		"dec": func(a int) int {
			return a - 1
		},
		"even": func(a int) bool {
			return a%2 == 0
		},
		"checkPermission": func(perms []string, r string) bool {
			r1, err := role.GetRole(r)
			if err != nil {
				log.Printf("failed to parse role for checkPermission: %+v", err)
				return false
			}
			m := role1.SufficientPermissionsFor(r1)

			for _, perm := range perms {
				if m[perm] {
					return true
				}
			}
			return false
		},
		"getTeamName": func(teamID int) string {
			t1, err := t.Team.GetTeam(context.Background(), team.Team{ID: teamID})
			if err != nil {
				log.Printf("failed to get team for getTeamName: %+v", err)
				return ""
			}
			return t1.Name
		},
	}
}
