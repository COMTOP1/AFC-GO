package views

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/player"
	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

func (v *Views) PlayersFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	playersDB, err := v.player.GetPlayers(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get players for players: %w", err)
	}

	teamsDB, err := v.team.GetTeams(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get teams for players: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year    int
		Players []PlayerTemplate
		Teams   []TeamTemplate
		User    user.User
		Context *Context
	}{
		Year:    year,
		Players: DBPlayersToTemplateFormat(playersDB, teamsDB),
		Teams:   DBTeamsToTemplateFormat(teamsDB),
		User:    c1.User,
		Context: c1,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.PlayersTemplate, templates.RegularType)
}

func (v *Views) PlayerAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		data := struct {
			Error string `json:"error"`
		}{}

		name := c.FormValue("name")
		if len(name) == 0 {
			log.Println("name must not be empty")
			data.Error = "name must not be empty"
			return c.JSON(http.StatusOK, data)
		}

		position := c.FormValue("position")

		teamID, err := strconv.Atoi(c.FormValue("playerTeam"))
		if err != nil {
			log.Printf("failed to parse playerTeam for playerAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to parse playerTeam for playerAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		_, err = v.team.GetTeam(c.Request().Context(), team.Team{ID: teamID})
		if err != nil {
			log.Printf("failed to get team for playerAdd: %+v, id: %d", err, teamID)
			data.Error = fmt.Sprintf("failed to get team for playerAdd: %+v, id: %d", err, teamID)
			return c.JSON(http.StatusOK, data)
		}

		dateOfBirth := c.Request().FormValue("dateOfBirth")

		parse, err := time.Parse("02/01/2006", dateOfBirth)
		if err != nil {
			log.Printf("failed to parse dateOfBirth for playerAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to parse dateOfBirth for playerAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		diff := time.Now().Compare(parse)
		if diff != 1 {
			log.Printf("dateOfBirth date be before today for playerAdd")
			data.Error = "dateOfBirth date be before today for playerAdd"
			return c.JSON(http.StatusOK, data)
		}

		var isCaptain bool

		tempIsCaptain := c.FormValue("isCaptain")
		if tempIsCaptain == "Y" {
			isCaptain = true
		} else if len(tempIsCaptain) != 0 {
			log.Printf("failed to parse isCaptain for playerAdd: %s", tempIsCaptain)
			data.Error = "failed to parse isCaptain for playerAdd: " + tempIsCaptain
			return c.JSON(http.StatusOK, data)
		}

		var fileName string
		hasUpload := true

		file, err := c.FormFile("upload")
		if err != nil {
			if !strings.Contains(err.Error(), "no such file") {
				log.Printf("failed to get file for playerAdd: %+v", err)
				data.Error = fmt.Sprintf("failed to get file for playerAdd: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			hasUpload = false
		}
		if hasUpload {
			fileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for playerAdd: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for playerAdd: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}

		_, err = v.player.AddPlayer(c.Request().Context(), player.Player{Name: name, FileName: null.NewString(fileName, len(fileName) > 0), DateOfBirth: null.TimeFrom(parse), Position: null.NewString(position, len(position) > 0), IsCaptain: isCaptain, TeamID: teamID})
		if err != nil {
			log.Printf("failed to add player for playerAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to add player for playerAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully added \"%s\"", name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for playerAdd: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, errors.New("invalid method used"))
}

func (v *Views) PlayerEditFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		playerID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to parse id for playerEdit: %w", err)
		}
		playerDB, err := v.player.GetPlayer(c.Request().Context(), player.Player{ID: playerID})
		if err != nil {
			return fmt.Errorf("failed to get player for playerEdit: %w", err)
		}

		playerDB.Name = c.FormValue("name")
		tempPosition := c.FormValue("position")
		playerDB.Position = null.NewString(tempPosition, len(tempPosition) > 0)

		data := struct {
			Error string `json:"error"`
		}{
			Error: "",
		}

		tempTeamID, err := strconv.Atoi(c.FormValue("playerTeam"))
		if err != nil {
			log.Printf("failed to parse playerTeam for playerEdit: %+v", err)
			data.Error = fmt.Sprintf("failed to parse playerTeam for playerEdit: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		_, err = v.team.GetTeam(c.Request().Context(), team.Team{ID: tempTeamID})
		if err != nil {
			log.Printf("failed to get team for playerEdit: %+v, id: %d", err, tempTeamID)
			data.Error = fmt.Sprintf("failed to get team for playerEdit: %+v, id: %d", err, tempTeamID)
			return c.JSON(http.StatusOK, data)
		}

		playerDB.TeamID = tempTeamID

		dateOfBirth := c.Request().FormValue("dateOfBirth")

		parse, err := time.Parse("02/01/2006", dateOfBirth)
		if err != nil {
			log.Printf("failed to parse dateOfBirth for playerEdit: %+v", err)
			data.Error = fmt.Sprintf("failed to parse dateOfBirth for playerEdit: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		diff := time.Now().Compare(parse)
		if diff != 1 {
			log.Printf("dateOfBirth date be before today for playerEdit")
			data.Error = "dateOfBirth date be before today for playerEdit"
			return c.JSON(http.StatusOK, data)
		}

		playerDB.DateOfBirth = null.TimeFrom(parse)

		tempIsCaptain := c.FormValue("isCaptain")
		if tempIsCaptain == "Y" {
			playerDB.IsCaptain = true
		} else if len(tempIsCaptain) != 0 {
			log.Printf("failed to parse isCaptain for playerEdit: %s", tempIsCaptain)
			data.Error = "failed to parse isCaptain for playerEdit: " + tempIsCaptain
			return c.JSON(http.StatusOK, data)
		}

		hasUpload := true

		file, err := c.FormFile("upload")
		if err != nil {
			if !strings.Contains(err.Error(), "no such file") {
				log.Printf("failed to get file for playerEdit: %+v", err)
				data.Error = fmt.Sprintf("failed to get file for playerEdit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			hasUpload = false
		}
		if hasUpload {
			var tempFileName string
			tempFileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for playerEdit: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for playerEdit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			if playerDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, playerDB.FileName.String))
				if err != nil {
					log.Printf("failed to delete old image for playerEdit: %+v", err)
				}
			}
			playerDB.FileName = null.NewString(tempFileName, len(tempFileName) > 0)
		}

		tempRemovePlayerImage := c.FormValue("removePlayerImage")
		if tempRemovePlayerImage == "Y" {
			if playerDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, playerDB.FileName.String))
				if err != nil {
					log.Printf("failed to delete image for playerEdit: %+v", err)
				}
			}
			playerDB.FileName = null.NewString("", false)
		} else if len(tempRemovePlayerImage) != 0 {
			log.Printf("failed to parse removePlayerImage for playerEdit: %s", tempRemovePlayerImage)
			data.Error = "failed to parse removePlayerImage for playerEdit: " + tempRemovePlayerImage
			return c.JSON(http.StatusOK, data)
		}

		_, err = v.player.EditPlayer(c.Request().Context(), playerDB)
		if err != nil {
			log.Printf("failed to edit player for playerEdit: %+v", err)
			data.Error = fmt.Sprintf("failed to edit player for playerEdit: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully edited \"%s\"", playerDB.Name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for playerEdit: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, errors.New("invalid method used"))
}

func (v *Views) PlayerDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for playerDelete: %w", err)
		}

		playerDB, err := v.player.GetPlayer(c.Request().Context(), player.Player{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get player for playerDelete: %w", err)
		}

		if playerDB.FileName.Valid {
			err = os.Remove(filepath.Join(v.conf.FileDir, playerDB.FileName.String))
			if err != nil {
				log.Printf("failed to delete player image for playerDelete: %+v", err)
			}
		}

		err = v.player.DeletePlayer(c.Request().Context(), playerDB)
		if err != nil {
			return fmt.Errorf("failed to delete player for playerDelete: %w", err)
		}

		c1.Message = fmt.Sprintf("successfully deleted \"%s\"", playerDB.Name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for playerDelete: %+v", err)
		}

		return c.Redirect(http.StatusFound, "/players")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, errors.New("invalid method used"))
}
