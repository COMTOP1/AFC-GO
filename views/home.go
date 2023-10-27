package views

import (
	"fmt"
	"github.com/COMTOP1/AFC-GO/affiliation"
	"github.com/COMTOP1/AFC-GO/news"
	"github.com/COMTOP1/AFC-GO/sponsor"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
	"github.com/COMTOP1/AFC-GO/whatson"
	"github.com/labstack/echo/v4"
	"time"
)

type (
	NewsTemplate struct {
		ID    int
		Title string
		Date  string
	}

	WhatsOnTemplate struct {
		ID          int
		Title       string
		Date        string
		DateOfEvent string
	}
)

func (v *Views) HomeFunc(c echo.Context) error {
	c1 := v.getSessionData(c)
	//page := structs.PageParams{
	//	MyUtils: myUtils.MyUtils{},
	//}

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

	//token, err := r.controller.Access.GetAFCToken(c.Request())
	//if err != nil {
	//	fmt.Println(err)
	//}

	//user1, err := r.controller.Session.GetUserByToken(token)
	//if err != nil {
	//	if strings.Contains(err.Error(), "invalid token") {
	//		fmt.Println(err)
	//	} else {
	//		fmt.Println(err)
	//	}
	//}

	year, _, _ := time.Now().Date()

	_ = c1.User

	data := struct {
		Year int
		//GetTime       int64
		//GetDate       func(Time int64) string
		Affiliations  []affiliation.Affiliation
		Sponsors      []sponsor.Sponsor
		NewsLatest    NewsTemplate
		WhatsOnLatest WhatsOnTemplate
		User          user.User
		//GetTeamName   func(id uint64) string
	}{
		Year: year,
		//GetTime:       page.MyUtils.GetTime,
		//GetDate:       page.MyUtils.GetDay,
		Affiliations:  affiliations,
		Sponsors:      sponsors,
		NewsLatest:    DBNewsLatestToTemplateFormat(newsLatest),
		WhatsOnLatest: DBWhatsOnLatestToTemplateFormat(whatsOnLatest),
		User:          c1.User,
		//GetTeamName: func(id uint64) string {
		//	team, err := v.team.GetTeamById(id)
		//	if err != nil {
		//		fmt.Println(err)
		//		return "TEAM NOT FOUND!"
		//	}
		//	return team.Name
		//},
	}

	err = v.template.RenderTemplate(c.Response().Writer, data, templates.HomeTemplate, templates.RegularType)
	if err != nil {
		fmt.Println(err)
	}
	return nil
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
	w.DateOfEvent = time.UnixMilli(whatsOnDB.TempDOE).Format("2006-01-02 15:04:05 MST")
	return w
}
