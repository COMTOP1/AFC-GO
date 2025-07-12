package views

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/affiliation"
	"github.com/COMTOP1/AFC-GO/sponsor"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

func (v *Views) HomeFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	affiliations, err := v.affiliation.GetAffiliationsMinimal(c.Request().Context())
	if err != nil {
		fmt.Println(err)
	}

	sponsors, err := v.sponsor.GetSponsorsMinimal(c.Request().Context())
	if err != nil {
		fmt.Println(err)
	}

	newsLatest, err := v.news.GetNewsLatest(c.Request().Context())
	if err != nil {
		fmt.Println(err)
	}

	whatsOnLatest, err := v.whatsOn.GetWhatsOnLatest(c.Request().Context())
	if err != nil {
		fmt.Println(err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year          int
		VisitorCount  int
		Affiliations  []affiliation.Affiliation
		Sponsors      []sponsor.Sponsor
		NewsLatest    NewsTemplate
		WhatsOnLatest WhatsOnTemplate
		User          user.User
		Context       *Context
	}{
		Year:          year,
		VisitorCount:  v.GetVisitorCount(),
		Affiliations:  affiliations,
		Sponsors:      sponsors,
		NewsLatest:    DBNewsToArticleTemplateFormat(newsLatest),
		WhatsOnLatest: DBWhatsOnToArticleTemplateFormat(whatsOnLatest),
		User:          c1.User,
		Context:       c1,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.HomeTemplate, templates.RegularType)
}
