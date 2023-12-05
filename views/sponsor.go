package views

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/sponsor"
	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

func (v *Views) SponsorsFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	sponsorsDB, err := v.sponsor.GetSponsors(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get sponsors for sponsors: %w", err)
	}

	teamsDB, err := v.team.GetTeams(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get teams for sponsors: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year     int
		Sponsors []SponsorTemplate
		Teams    []TeamTemplate
		User     user.User
		Context  *Context
	}{
		Year:     year,
		Sponsors: DBSponsorsToTemplateFormat(sponsorsDB),
		Teams:    DBTeamsToTemplateFormat(teamsDB),
		User:     c1.User,
		Context:  c1,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.SponsorsTemplate, templates.RegularType)
}

func (v *Views) SponsorAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		name := c.FormValue("name")
		purpose := c.FormValue("purpose")

		data := struct {
			Error string `json:"error"`
		}{
			Error: "",
		}

		website := c.FormValue("website")
		if len(website) > 0 {
			_, err := url.ParseRequestURI(website)
			if err != nil {
				log.Printf("failed to parse website for sponsorAdd: %+v", err)
				data.Error = fmt.Sprintf("failed to parse website for sponsorAdd: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}

		var teamActual string

		teamValue := c.FormValue("team")
		if teamValue == "A" || teamValue == "O" || teamValue == "Y" {
			teamActual = teamValue
		} else if len(teamValue) > 0 {
			teamID, err := strconv.Atoi(teamValue)
			if err != nil {
				log.Printf("failed to parse team for sponsorAdd: %+v", err)
				data.Error = fmt.Sprintf("failed to parse team for sponsorAdd: %+v", err)
				return c.JSON(http.StatusOK, data)
			}

			_, err = v.team.GetTeam(c.Request().Context(), team.Team{ID: teamID})
			if err != nil {
				log.Printf("failed to get team for sponsorAdd: %+v", err)
				data.Error = fmt.Sprintf("failed to get team for sponsorAdd: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			teamActual = teamValue
		}

		file, err := c.FormFile("upload")
		if err != nil {
			log.Printf("failed to get file for sponsorAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to get file for sponsorAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}
		fileName, err := v.fileUpload(file)
		if err != nil {
			log.Printf("failed to upload file for sponsorAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to upload file for sponsorAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		_, err = v.sponsor.AddSponsor(c.Request().Context(), sponsor.Sponsor{Name: name, Website: null.NewString(website, len(website) > 0), FileName: null.NewString(fileName, len(fileName) > 0), TeamID: teamActual, Purpose: null.NewString(purpose, len(purpose) > 0)})
		if err != nil {
			log.Printf("failed to add sponsor for sponsorAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to add sponsor for sponsorAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully added \"%s\"", name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for sponsorAdd: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

func (v *Views) SponsorDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for sponsorDelete: %w", err)
		}

		sponsorDB, err := v.sponsor.GetSponsor(c.Request().Context(), sponsor.Sponsor{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get sponsor for sponsorDelete: %w", err)
		}

		if sponsorDB.FileName.Valid {
			err = os.Remove(filepath.Join(v.conf.FileDir, sponsorDB.FileName.String))
			if err != nil {
				log.Printf("failed to delete sponsor image for sponsorDelete: %+v", err)
			}
		}

		err = v.sponsor.DeleteSponsor(c.Request().Context(), sponsorDB)
		if err != nil {
			return fmt.Errorf("failed to delete sponsor for sponsorDelete: %w", err)
		}

		c1.Message = fmt.Sprintf("successfully deleted \"%s\"", sponsorDB.Name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for sponsorDelete: %+v", err)
		}

		return c.Redirect(http.StatusFound, "/sponsors")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}
