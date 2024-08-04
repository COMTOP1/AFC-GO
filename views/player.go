package views

import (
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
		return fmt.Errorf("failed to get players for players, error: %w", err)
	}

	teamsDB, err := v.team.GetTeams(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get teams for players, error: %w", err)
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
			log.Println("name must not be empty for player add")
			data.Error = "name must not be empty"
			return c.JSON(http.StatusOK, data)
		}

		position := c.FormValue("position")

		teamID, err := strconv.Atoi(c.FormValue("playerTeam"))
		if err != nil {
			log.Printf("failed to parse playerTeam for player add, error: %+v", err)
			data.Error = fmt.Sprintf("failed to parse playerTeam for player add: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		_, err = v.team.GetTeam(c.Request().Context(), team.Team{ID: teamID})
		if err != nil {
			log.Printf("failed to get team for player add, team id: %d, error: %+v", teamID, err)
			data.Error = fmt.Sprintf("failed to get team for player add, team id: %d: %+v", teamID, err)
			return c.JSON(http.StatusOK, data)
		}

		dateOfBirth := c.Request().FormValue("dateOfBirth")

		parse, err := time.Parse("02/01/2006", dateOfBirth)
		if err != nil {
			log.Printf("failed to parse dateOfBirth for player add, error: %+v", err)
			data.Error = fmt.Sprintf("failed to parse dateOfBirth for player add: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		diff := time.Now().Compare(parse)
		if diff != 1 {
			log.Printf("dateOfBirth date be before today for player add")
			data.Error = "dateOfBirth date be before today for player add"
			return c.JSON(http.StatusOK, data)
		}

		var isCaptain bool

		tempIsCaptain := c.FormValue("isCaptain")
		if tempIsCaptain == "Y" {
			isCaptain = true
		} else if len(tempIsCaptain) != 0 {
			log.Printf("failed to parse isCaptain for player add, value: %s", tempIsCaptain)
			data.Error = "failed to parse isCaptain for player add, value: " + tempIsCaptain
			return c.JSON(http.StatusOK, data)
		}

		var fileName string
		hasUpload := true

		file, err := c.FormFile("upload")
		if err != nil {
			if !strings.Contains(err.Error(), "no such file") {
				log.Printf("failed to get file for player add, error: %+v", err)
				data.Error = fmt.Sprintf("failed to get file for player add: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			hasUpload = false
		}
		if hasUpload {
			fileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for player add, error: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for player add: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}

		_, err = v.player.AddPlayer(c.Request().Context(), player.Player{Name: name, FileName: null.NewString(fileName, len(fileName) > 0), DateOfBirth: null.TimeFrom(parse), Position: null.NewString(position, len(position) > 0), IsCaptain: isCaptain, TeamID: teamID})
		if err != nil {
			log.Printf("failed to add player for player add, error: %+v", err)
			data.Error = fmt.Sprintf("failed to add player for player add: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully added \"%s\"", name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for player add, error: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return v.invalidMethodUsed(c)
}

func (v *Views) PlayerEditFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		playerID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to parse id for player edit, error: %w", err)
		}

		playerDB, err := v.player.GetPlayer(c.Request().Context(), player.Player{ID: playerID})
		if err != nil {
			return fmt.Errorf("failed to get player for player edit, player id: %d, error: %w", playerID, err)
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
			log.Printf("failed to parse playerTeam for player edit, player id: %d, error: %+v", playerID, err)
			data.Error = fmt.Sprintf("failed to parse playerTeam for player edit: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		_, err = v.team.GetTeam(c.Request().Context(), team.Team{ID: tempTeamID})
		if err != nil {
			log.Printf("failed to get team for player edit, player id: %d, team id: %d, error: %+v", playerID, tempTeamID, err)
			data.Error = fmt.Sprintf("failed to get team for player edit, team id: %d: %+v", tempTeamID, err)
			return c.JSON(http.StatusOK, data)
		}

		playerDB.TeamID = tempTeamID

		dateOfBirth := c.Request().FormValue("dateOfBirth")

		parse, err := time.Parse("02/01/2006", dateOfBirth)
		if err != nil {
			log.Printf("failed to parse dateOfBirth for player edit, player id: %d, error: %+v", playerID, err)
			data.Error = fmt.Sprintf("failed to parse dateOfBirth for player edit: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		diff := time.Now().Compare(parse)
		if diff != 1 {
			log.Printf("dateOfBirth date be before today for player edit, player id: %d", playerID)
			data.Error = "dateOfBirth date be before today for player edit"
			return c.JSON(http.StatusOK, data)
		}

		playerDB.DateOfBirth = null.TimeFrom(parse)

		tempIsCaptain := c.FormValue("isCaptain")
		if tempIsCaptain == "Y" {
			playerDB.IsCaptain = true
		} else if len(tempIsCaptain) != 0 {
			log.Printf("failed to parse isCaptain for player edit, player id: %d, value: %s", playerID, tempIsCaptain)
			data.Error = "failed to parse isCaptain for player edit, value: " + tempIsCaptain
			return c.JSON(http.StatusOK, data)
		}

		hasUpload := true

		file, err := c.FormFile("upload")
		if err != nil {
			if !strings.Contains(err.Error(), "no such file") {
				log.Printf("failed to get file for player edit, player id: %d, error: %+v", playerID, err)
				data.Error = fmt.Sprintf("failed to get file for player edit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			hasUpload = false
		}
		if hasUpload {
			var tempFileName string
			tempFileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for player edit, player id: %d, error: %+v", playerID, err)
				data.Error = fmt.Sprintf("failed to upload file for player edit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			if playerDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, playerDB.FileName.String))
				if err != nil {
					log.Printf("failed to delete old image for player edit, player id: %d, error: %+v", playerID, err)
				}
			}
			playerDB.FileName = null.NewString(tempFileName, len(tempFileName) > 0)
		}

		tempRemovePlayerImage := c.FormValue("removePlayerImage")
		if tempRemovePlayerImage == "Y" {
			if playerDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, playerDB.FileName.String))
				if err != nil {
					log.Printf("failed to delete image for player edit, player id: %d, error: %+v", playerID, err)
				}
			}
			playerDB.FileName = null.NewString("", false)
		} else if len(tempRemovePlayerImage) != 0 {
			log.Printf("failed to parse removePlayerImage for player edit, player id: %d, error: %s", playerID, tempRemovePlayerImage)
			data.Error = "failed to parse removePlayerImage for player edit: " + tempRemovePlayerImage
			return c.JSON(http.StatusOK, data)
		}

		_, err = v.player.EditPlayer(c.Request().Context(), playerDB)
		if err != nil {
			log.Printf("failed to edit player for player edit, player id: %d, error: %+v", playerID, err)
			data.Error = fmt.Sprintf("failed to edit player for player edit: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully edited \"%s\"", playerDB.Name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for player edit, player id: %d, error: %+v", playerID, err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return v.invalidMethodUsed(c)
}

func (v *Views) PlayerDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for player delete, error: %w", err)
		}

		playerDB, err := v.player.GetPlayer(c.Request().Context(), player.Player{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get player for player delete, player id: %d, error: %w", id, err)
		}

		if playerDB.FileName.Valid {
			err = os.Remove(filepath.Join(v.conf.FileDir, playerDB.FileName.String))
			if err != nil {
				log.Printf("failed to delete player image for player delete, player id: %d, error: %+v", id, err)
			}
		}

		err = v.player.DeletePlayer(c.Request().Context(), playerDB)
		if err != nil {
			return fmt.Errorf("failed to delete player for player delete, player id: %d, error: %w", id, err)
		}

		c1.Message = fmt.Sprintf("successfully deleted \"%s\"", playerDB.Name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for player delete, player id: %d, error: %+v", id, err)
		}

		return c.Redirect(http.StatusFound, "/players")
	}
	return v.invalidMethodUsed(c)
}
