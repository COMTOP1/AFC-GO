package views

//import (
//	//"fmt"
//	//"github.com/COMTOP1/AFC-GO/structs"
//	//"github.com/COMTOP1/api/handler"
//	"fmt"
//	"github.com/COMTOP1/api/handler"
//	"github.com/COMTOP1/api/utils"
//	"github.com/labstack/echo/v4"
//	"net/http"
//	"strconv"
//	"strings"
//	"unicode"
//
//	//"net/http"
//	//"strconv"
//	//"strings"
//	//"unicode"
//)
//
//func (v *Views) TeamsFunc(c echo.Context) error {
//	_ = c
//	page := structs.PageParams{
//		//MyUtils: myUtils.MyUtils{},
//	}
//
//	token, err := v.controller.Access.GetAFCToken(c.Request())
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	user, err := v.controller.Session.GetUserByToken(token)
//	if err != nil {
//		if strings.Contains(err.Error(), "invalid token") {
//			fmt.Println(err)
//		} else {
//			fmt.Println(err)
//		}
//	}
//
//	var teams []handler.Team
//
//	if user.Id != 0 {
//		teams, err = v.controller.Session.ListAllTeams(token)
//		if err != nil {
//			fmt.Println(err)
//		}
//	} else {
//		teams, err = v.controller.Session.ListActiveTeams()
//		if err != nil {
//			fmt.Println(err)
//		}
//	}
//
//	data := struct {
//		MyUtils struct {
//			GetYear func() int
//		}
//		GetTime func() int64
//		GetDate func(Time int64) string
//		Teams   []handler.Team
//		User    handler.User
//		//GetTeamName   func(id uint64) string
//	}{
//		MyUtils: struct {
//			GetYear func() int
//		}{
//			GetYear: page.MyUtils.GetYear,
//		},
//		GetTime: page.MyUtils.GetTime,
//		GetDate: page.MyUtils.GetDay,
//		Teams:   teams,
//		User:    user,
//		//GetTeamName: func(id uint64) string {
//		//	team, err := v.controller.Session.GetTeamById(id)
//		//	if err != nil {
//		//		fmt.Println(err)
//		//		return "TEAM NOT FOUND!"
//		//	}
//		//	return team.Name
//		//},
//	}
//	_ = data
//
//	//err = v.controller.Template.RenderTemplate(c.Response().Writer, page, data, "teams.tmpl")
//	if err != nil {
//		fmt.Println(err)
//	}
//	if err != nil {
//		fmt.Println(err)
//	}
//	return nil
//}
//
//func (v *Views) TeamFunc(c echo.Context) error {
//	_ = c
//	fmt.Println(c)
//	temp := c.QueryParam("id")
//	temp1 := []rune(temp)
//	for _, r2 := range temp1 {
//		if !unicode.IsNumber(r2) {
//			return echo.NewHTTPError(http.StatusBadRequest, utils.Error{Error: "id expects a positive number, the provided is not a positive number"})
//		}
//	}
//	id, err := strconv.ParseUint(temp, 10, 64)
//	if err != nil {
//		err = fmt.Errorf("team failed to get id: %w", err)
//		//return echo.NewHTTPError(http.StatusInternalServerError, utils.Error{Error: err.Error()})
//		return c.Redirect(http.StatusSeeOther, "/teams")
//	}
//
//	team, err := r.controller.Session.GetTeamById(id)
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	if !team.Active {
//		return c.Redirect(http.StatusSeeOther, "/teams")
//	}
//
//	if team.Id == 0 {
//		return c.JSON(http.StatusInternalServerError, utils.Error{Error: "team not found"})
//	}
//
//	managers, err := v.controller.Session.ListTeamManagersUsers(id)
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	sponsors, err := v.controller.Session.ListAllSponsorsByTeamId(id)
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	page := structs.PageParams{
//		//MyUtils: myUtils.MyUtils{},
//	}
//
//	token, err := v.controller.Access.GetAFCToken(c.Request())
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	user, err := v.controller.Session.GetUserByToken(token)
//	if err != nil {
//		if strings.Contains(err.Error(), "invalid token") {
//			fmt.Println(err)
//		} else {
//			fmt.Println(err)
//		}
//	}
//
//	players, err := v.controller.Session.ListAllPlayersByTeam(id)
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	data := struct {
//		MyUtils struct {
//			GetYear func() int
//		}
//		GetTime  func() int64
//		GetDate  func(Time int64) string
//		Team     handler.Team
//		Managers []handler.User
//		Sponsors []handler.Sponsor
//		Players  []handler.Player
//		User     handler.User
//	}{
//		MyUtils: struct {
//			GetYear func() int
//		}{
//			GetYear: page.MyUtils.GetYear,
//		},
//		GetTime:  page.MyUtils.GetTime,
//		GetDate:  page.MyUtils.GetDay,
//		Team:     team,
//		Managers: managers,
//		Sponsors: sponsors,
//		Players:  players,
//		User:     user,
//	}
//
//	_ = data
//
//	//err = r.controller.Template.RenderTemplate(c.Response().Writer, page, data, "team.tmpl")
//	if err != nil {
//		fmt.Println(err)
//	}
//	return nil
//}
