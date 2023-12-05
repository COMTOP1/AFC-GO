package views

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

func (v *Views) TeamsFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	var teams []team.Team
	var err error

	if c1.User.ID != 0 {
		teams, err = v.team.GetTeams(c.Request().Context())
	} else {
		teams, err = v.team.GetTeamsActive(c.Request().Context())
	}
	if err != nil {
		return fmt.Errorf("failed to get teams for teams: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year    int
		Teams   []TeamTemplate
		User    user.User
		Context *Context
	}{
		Year:    year,
		Teams:   DBTeamsToTemplateFormat(teams),
		User:    c1.User,
		Context: c1,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.TeamsTemplate, templates.RegularType)
}

func (v *Views) TeamFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse id for team: %w", err))
	}
	teamDB, err := v.team.GetTeam(c.Request().Context(), team.Team{ID: teamID})
	if err != nil {
		return fmt.Errorf("failed to get team for team: %w", err)
	}

	managersDB, err := v.user.GetUsersManagersTeam(c.Request().Context(), teamDB)
	if err != nil {
		return fmt.Errorf("failed to get managers for team: %w, id: %d", err, teamID)
	}

	sponsorsDB, err := v.sponsor.GetSponsorsTeam(c.Request().Context(), teamDB)
	if err != nil {
		return fmt.Errorf("failed to get sponsors for team: %w, id: %d", err, teamID)
	}

	playersDB, err := v.player.GetPlayersTeam(c.Request().Context(), teamDB)
	if err != nil {
		return fmt.Errorf("failed to get players for team: %w, id: %d", err, teamID)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year     int
		Team     team.Team
		Managers []string
		Sponsors []SponsorTemplate
		Players  []PlayerTemplate
		User     user.User
		Context  *Context
	}{
		Year:     year,
		Team:     teamDB,
		Managers: DBManagersToTemplateFormat(managersDB),
		Sponsors: DBSponsorsToTemplateFormat(sponsorsDB),
		Players:  DBPlayersTeamToTemplateFormat(playersDB),
		User:     c1.User,
		Context:  c1,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.TeamTemplate, templates.RegularType)
}

func (v *Views) TeamAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		data := struct {
			Error string `json:"error"`
		}{
			Error: "",
		}

		name := c.FormValue("name")
		if len(name) == 0 {
			log.Printf("name must contain a value for teamAdd")
			data.Error = fmt.Sprintf("name must contain a value")
			return c.JSON(http.StatusOK, data)
		}

		league := c.FormValue("league")
		division := c.FormValue("division")

		leagueTable := c.FormValue("leagueTable")
		if len(leagueTable) > 0 {
			_, err := url.ParseRequestURI(leagueTable)
			if err != nil {
				log.Printf("failed to parse leagueTable for teamAdd: %+v", err)
				data.Error = fmt.Sprintf("failed to parse leagueTable for teamAdd: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}

		fixtures := c.FormValue("fixtures")
		if len(fixtures) > 0 {
			_, err := url.ParseRequestURI(fixtures)
			if err != nil {
				log.Printf("failed to parse fixtures for teamAdd: %+v", err)
				data.Error = fmt.Sprintf("failed to parse fixtures for teamAdd: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}

		tempCoach := c.FormValue("tempCoach")
		coach := null.StringFrom(tempCoach)
		if len(tempCoach) == 0 {
			coach.Valid = false
		}

		tempPhysio := c.FormValue("physio")
		physio := null.StringFrom(tempPhysio)
		if len(tempPhysio) == 0 {
			physio.Valid = false
		}

		var isActive, isYouth bool

		tempIsActive := c.FormValue("isActive")
		if tempIsActive == "Y" {
			isActive = true
		} else if len(tempIsActive) != 0 {
			log.Printf("failed to parse isActive for teamAdd: %s", tempIsActive)
			data.Error = fmt.Sprintf("failed to parse isActive for teamAdd: %s", tempIsActive)
			return c.JSON(http.StatusOK, data)
		}

		tempIsYouth := c.FormValue("isYouth")
		if tempIsYouth == "Y" {
			isActive = true
		} else if len(tempIsYouth) != 0 {
			log.Printf("failed to parse isYouth for teamAdd: %s", tempIsYouth)
			data.Error = fmt.Sprintf("failed to parse isYouth for teamAdd: %s", tempIsYouth)
			return c.JSON(http.StatusOK, data)
		}

		ages, err := strconv.Atoi(c.FormValue("ages"))
		if err != nil {
			log.Printf("failed to parse ages for playerAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to parse ages for playerAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		if ages < 19 {
			isYouth = true
		}

		var fileName string
		hasUpload := true

		file, err := c.FormFile("upload")
		if err != nil {
			if !strings.Contains(err.Error(), "no such file") {
				log.Printf("failed to get file for teamAdd: %+v", err)
				data.Error = fmt.Sprintf("failed to get file for teamAdd: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			hasUpload = false
		}
		if hasUpload {
			fileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for teamAdd: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for teamAdd: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}

		_, err = v.team.AddTeam(c.Request().Context(), team.Team{Name: name, League: null.StringFrom(league), Division: null.StringFrom(division), LeagueTable: null.StringFrom(leagueTable), Fixtures: null.StringFrom(fixtures), Coach: coach, Physio: physio, FileName: null.StringFrom(fileName), IsActive: isActive, IsYouth: isYouth, Ages: ages})
		if err != nil {
			log.Printf("failed to add team for teamAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to add team for teamAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully added \"%s\"", name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for teamAdd: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

func (v *Views) TeamEditFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}

func (v *Views) TeamDeleteFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}
