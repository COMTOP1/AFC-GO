package controllers

import (
	"fmt"
	"github.com/COMTOP1/AFC-GO/myUtils"
	"github.com/COMTOP1/AFC-GO/structs"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/api/handler"
	"github.com/labstack/echo/v4"
	"strings"
)

type HomeRepo struct {
	controller Controller
}

func NewHomeRepo(controller Controller) *HomeRepo {
	return &HomeRepo{
		controller: controller,
	}
}

func (r *HomeRepo) Home(c echo.Context) error {
	page := structs.PageParams{
		MyUtils: myUtils.MyUtils{},
	}

	affiliations, err := r.controller.Session.ListAllAffiliations()
	if err != nil {
		fmt.Println(err)
	}

	sponsors, err := r.controller.Session.ListAllSponsorsMinimal()
	if err != nil {
		fmt.Println(err)
	}

	newsLatest, err := r.controller.Session.GetNewsLatest()
	if err != nil {
		fmt.Println(err)
	}

	whatsOnLatest, err := r.controller.Session.GetWhatsOnLatest()
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
			team, err := r.controller.Session.GetTeamById(id)
			if err != nil {
				fmt.Println(err)
				return "TEAM NOT FOUND!"
			}
			return team.Name
		},
	}

	err = r.controller.Template.RenderTemplate(c.Response().Writer, data, templates.HomeTemplate, templates.RegularType)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}
