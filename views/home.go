package views

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/affiliation"
	"github.com/COMTOP1/AFC-GO/news"
	"github.com/COMTOP1/AFC-GO/sponsor"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
	"github.com/COMTOP1/AFC-GO/whatson"
)

type (
	NewsTemplate struct {
		ID      int
		Title   string
		Content string
		Date    string
	}

	WhatsOnTemplate struct {
		ID          int
		Title       string
		Content     string
		Date        string
		DateOfEvent string
	}
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
		Affiliations  []affiliation.Affiliation
		Sponsors      []sponsor.Sponsor
		NewsLatest    NewsTemplate
		WhatsOnLatest WhatsOnTemplate
		User          user.User
	}{
		Year:          year,
		Affiliations:  affiliations,
		Sponsors:      sponsors,
		NewsLatest:    DBNewsLatestToTemplateFormat(newsLatest),
		WhatsOnLatest: DBWhatsOnLatestToTemplateFormat(whatsOnLatest),
		User:          c1.User,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.HomeTemplate, templates.RegularType)
}

func DBNewsLatestToTemplateFormat(newsDB news.News) NewsTemplate {
	var n NewsTemplate
	n.ID = newsDB.ID
	n.Title = newsDB.Title
	n.Date = time.UnixMilli(newsDB.Temp).Format("2006-01-02 15:04:05 MST")
	return n
}

func DBWhatsOnLatestToTemplateFormat(whatsOnDB whatson.WhatsOn) WhatsOnTemplate {
	var w WhatsOnTemplate
	w.ID = whatsOnDB.ID
	w.Title = whatsOnDB.Title
	w.Date = time.UnixMilli(whatsOnDB.TempDate).Format("2006-01-02 15:04:05 MST")
	w.DateOfEvent = time.UnixMilli(whatsOnDB.TempDOE).Format("2006-01-02")
	return w
}
