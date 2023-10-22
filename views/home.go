package views

import (
	"fmt"
	"github.com/COMTOP1/AFC-GO/myUtils"
	"github.com/COMTOP1/AFC-GO/structs"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/api/handler"
	"github.com/labstack/echo/v4"
	"strings"
)

func (v *Views) Home(c echo.Context) error {
	page := structs.PageParams{
		MyUtils: myUtils.MyUtils{},
	}

	affiliations, err := v.affiliation.ListAllAffiliations()
	if err != nil {
		fmt.Println(err)
	}

	sponsors, err := v.sponsor.ListAllSponsorsMinimal()
	if err != nil {
		fmt.Println(err)
	}

	newsLatest, err := v.news.GetNewsLatest()
	if err != nil {
		fmt.Println(err)
	}

	whatsOnLatest, err := v.whatsOn.GetWhatsOnLatest()
	if err != nil {
		fmt.Println(err)
	}

	token, err := r.controller.Access.GetAFCToken(c.Request())
	if err != nil {
		fmt.Println(err)
	}

	user, err := r.controller.Session.GetUserByToken(token)
	if err != nil {
		if strings.Contains(err.Error(), "invalid token") {
			fmt.Println(err)
		} else {
			fmt.Println(err)
		}
	}

	data := struct {
		MyUtils struct {
			GetYear func() int
		}
		GetTime       func() int64
		GetDate       func(Time int64) string
		Affiliations  []handler.Affiliation
		Sponsors      []handler.Sponsor
		NewsLatest    handler.News
		WhatsOnLatest handler.WhatsOn
		User          handler.User
		GetTeamName   func(id uint64) string
	}{
		MyUtils: struct {
			GetYear func() int
		}{
			GetYear: page.MyUtils.GetYear,
		},
		GetTime:       page.MyUtils.GetTime,
		GetDate:       page.MyUtils.GetDay,
		Affiliations:  affiliations,
		Sponsors:      sponsors,
		NewsLatest:    newsLatest,
		WhatsOnLatest: whatsOnLatest,
		User:          user,
		GetTeamName: func(id uint64) string {
			team, err := v.team.GetTeamById(id)
			if err != nil {
				fmt.Println(err)
				return "TEAM NOT FOUND!"
			}
			return team.Name
		},
	}

	err = v.template.RenderTemplate(c.Response().Writer, data, templates.HomeTemplate, templates.RegularType)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}
