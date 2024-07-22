package views

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
			data.Error = "name must contain a value"
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

		coach := c.FormValue("coach")
		physio := c.FormValue("physio")

		var isActive, isYouth bool

		tempIsActive := c.FormValue("isActive")
		if tempIsActive == "Y" {
			isActive = true
		} else if len(tempIsActive) != 0 {
			log.Printf("failed to parse isActive for teamAdd: %s", tempIsActive)
			data.Error = "failed to parse isActive for teamAdd: " + tempIsActive
			return c.JSON(http.StatusOK, data)
		}

		tempIsYouth := c.FormValue("isYouth")
		if tempIsYouth == "Y" {
			isActive = true
		} else if len(tempIsYouth) != 0 {
			log.Printf("failed to parse isYouth for teamAdd: %s", tempIsYouth)
			data.Error = "failed to parse isYouth for teamAdd: " + tempIsYouth
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

		_, err = v.team.AddTeam(c.Request().Context(), team.Team{Name: name, League: null.NewString(league, len(league) > 0), Division: null.NewString(division, len(division) > 0), LeagueTable: null.NewString(leagueTable, len(leagueTable) > 0), Fixtures: null.NewString(fixtures, len(fixtures) > 0), Coach: null.NewString(coach, len(coach) > 0), Physio: null.NewString(physio, len(physio) > 0), FileName: null.NewString(fileName, len(fileName) > 0), IsActive: isActive, IsYouth: isYouth, Ages: ages})
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
	return echo.NewHTTPError(http.StatusMethodNotAllowed, errors.New("invalid method used"))
}

func (v *Views) TeamEditFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		data := struct {
			Error string `json:"error"`
		}{
			Error: "",
		}

		teamID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to parse id for teamEdit: %w", err)
		}
		teamDB, err := v.team.GetTeam(c.Request().Context(), team.Team{ID: teamID})
		if err != nil {
			return fmt.Errorf("failed to get team for teamEdit: %w", err)
		}

		name := c.FormValue("name")
		if len(name) == 0 {
			log.Printf("name must contain a value for teamEdit")
			data.Error = "name must contain a value"
			return c.JSON(http.StatusOK, data)
		}

		teamDB.Name = name

		tempLeague := c.FormValue("league")
		teamDB.League = null.NewString(tempLeague, len(tempLeague) > 0)
		tempDivision := c.FormValue("division")
		teamDB.Division = null.NewString(tempDivision, len(tempDivision) > 0)

		tempLeagueTable := c.FormValue("leagueTable")
		if len(tempLeagueTable) > 0 {
			_, err = url.ParseRequestURI(tempLeagueTable)
			if err != nil {
				log.Printf("failed to parse leagueTable for teamEdit: %+v", err)
				data.Error = fmt.Sprintf("failed to parse leagueTable for teamEdit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}

		teamDB.LeagueTable = null.NewString(tempLeagueTable, len(tempLeagueTable) > 0)

		tempFixtures := c.FormValue("fixtures")
		if len(tempFixtures) > 0 {
			_, err = url.ParseRequestURI(tempFixtures)
			if err != nil {
				log.Printf("failed to parse fixtures for teamEdit: %+v", err)
				data.Error = fmt.Sprintf("failed to parse fixtures for teamEdit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}

		teamDB.Fixtures = null.NewString(tempFixtures, len(tempFixtures) > 0)

		tempCoach := c.FormValue("coach")
		teamDB.Coach = null.NewString(tempCoach, len(tempCoach) > 0)
		tempPhysio := c.FormValue("physio")
		teamDB.Physio = null.NewString(tempPhysio, len(tempPhysio) > 0)

		tempIsActive := c.FormValue("isActive")
		if tempIsActive == "Y" {
			teamDB.IsActive = true
		} else if len(tempIsActive) != 0 {
			log.Printf("failed to parse isActive for teamEdit: %s", tempIsActive)
			data.Error = "failed to parse isActive for teamEdit: " + tempIsActive
			return c.JSON(http.StatusOK, data)
		}

		tempIsYouth := c.FormValue("isYouth")
		if tempIsYouth == "Y" {
			teamDB.IsYouth = true
		} else if len(tempIsYouth) != 0 {
			log.Printf("failed to parse isYouth for teamEdit: %s", tempIsYouth)
			data.Error = "failed to parse isYouth for teamEdit: " + tempIsYouth
			return c.JSON(http.StatusOK, data)
		}

		ages, err := strconv.Atoi(c.FormValue("ages"))
		if err != nil {
			log.Printf("failed to parse ages for playerAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to parse ages for playerAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		if ages < 19 {
			teamDB.IsYouth = true
		}

		teamDB.Ages = ages

		hasUpload := true

		file, err := c.FormFile("upload")
		if err != nil {
			if !strings.Contains(err.Error(), "no such file") {
				log.Printf("failed to get file for teamEdit: %+v", err)
				data.Error = fmt.Sprintf("failed to get file for teamEdit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			hasUpload = false
		}
		if hasUpload {
			var fileName string
			fileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for teamEdit: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for teamEdit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			if teamDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, teamDB.FileName.String))
				if err != nil {
					log.Printf("failed to delete old image for teamEdit: %+v", err)
				}
			}
			teamDB.FileName = null.StringFrom(fileName)
		}

		tempRemoveTeamImage := c.FormValue("removeTeamImage")
		if tempRemoveTeamImage == "Y" {
			if teamDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, teamDB.FileName.String))
				if err != nil {
					log.Printf("failed to delete image for teamEdit: %+v", err)
				}
			}
			teamDB.FileName = null.NewString("", false)
		} else if len(tempRemoveTeamImage) != 0 {
			log.Printf("failed to parse removeTeamImage for teamEdit: %s", tempRemoveTeamImage)
			data.Error = "failed to parse removeTeamImage for teamEdit: " + tempRemoveTeamImage
			return c.JSON(http.StatusOK, data)
		}

		_, err = v.team.EditTeam(c.Request().Context(), teamDB)
		if err != nil {
			log.Printf("failed to edit team for teamEdit: %+v", err)
			data.Error = fmt.Sprintf("failed to edit team for teamEdit: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully edited \"%s\"", name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for teamEdit: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, errors.New("invalid method used"))
}

func (v *Views) TeamDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for teamDelete: %w", err)
		}

		teamDB, err := v.team.GetTeam(c.Request().Context(), team.Team{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get team for teamDelete: %w", err)
		}

		playersDB, err := v.player.GetPlayersTeam(c.Request().Context(), teamDB)
		if err != nil {
			return fmt.Errorf("failed to get players for teamDelete: %w", err)
		}

		for _, playerDB := range playersDB {
			playerDB.TeamID = 0
			_, err = v.player.EditPlayer(c.Request().Context(), playerDB)
			if err != nil {
				return errors.New("failed to edit player for teamDelete")
			}
		}

		sponsorsDB, err := v.sponsor.GetSponsorsTeam(c.Request().Context(), teamDB)
		if err != nil {
			return fmt.Errorf("failed to get sponsors for teamDelete: %w", err)
		}

		for _, sponsorDB := range sponsorsDB {
			sponsorDB.TeamID = "A"
			_, err = v.sponsor.EditSponsor(c.Request().Context(), sponsorDB)
			if err != nil {
				return errors.New("failed to edit sponsor for teamDelete")
			}
		}

		usersDB, err := v.user.GetUsersManagersTeam(c.Request().Context(), teamDB)
		if err != nil {
			return fmt.Errorf("failed to get managers for teamDelete: %w", err)
		}

		for _, userDB := range usersDB {
			userDB.TeamID = 0
			_, err = v.user.EditUser(c.Request().Context(), userDB)
			if err != nil {
				return errors.New("failed to edit user for teamDelete")
			}
		}

		if teamDB.FileName.Valid {
			err = os.Remove(filepath.Join(v.conf.FileDir, teamDB.FileName.String))
			if err != nil {
				log.Printf("failed to delete team file for teamDelete: %+v", err)
			}
		}

		err = v.team.DeleteTeam(c.Request().Context(), teamDB)
		if err != nil {
			return fmt.Errorf("failed to delete team for teamDelete: %w", err)
		}

		c1.Message = fmt.Sprintf("successfully deleted \"%s\"", teamDB.Name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for teamDelete: %+v", err)
		}

		return c.Redirect(http.StatusFound, "/teams")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, errors.New("invalid method used"))
}
