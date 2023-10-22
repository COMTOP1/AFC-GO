package controllers

import (
	"fmt"
	"github.com/COMTOP1/AFC-GO/utils"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"unicode"
)

type DownloadRepo struct {
	controller Controller
}

func NewDownloadRepo(controller Controller) *DownloadRepo {
	return &DownloadRepo{
		controller: controller,
	}
}

func (r *DownloadRepo) Download(c echo.Context) error {
	source := c.QueryParam("s")
	temp := c.QueryParam("id")
	temp1 := []rune(temp)
	for _, r2 := range temp1 {
		if !unicode.IsNumber(r2) {
			return echo.NewHTTPError(http.StatusBadRequest, utils.Error{Error: "id expects a positive number, the provided is not a positive number"})
		}
	}
	id, err := strconv.ParseUint(temp, 10, 64)
	if err != nil {
		err = fmt.Errorf("download failed to get id: %w", err)
		return echo.NewHTTPError(http.StatusInternalServerError, utils.Error{Error: err.Error()})
	}
	if len(source) != 1 {
		err = fmt.Errorf("download failed to get source: source format does not conform")
		return echo.NewHTTPError(http.StatusBadRequest, utils.Error{Error: err.Error()})
	}
	switch source {
	case "a": // Affiliation
		affiliation, err := r.controller.Session.GetAffiliationById(id)
		if err != nil {
			err = fmt.Errorf("download failed to get affiliation: %w", err)
			return echo.NewHTTPError(http.StatusInternalServerError, utils.Error{Error: err.Error()})
		}
		if len(affiliation.FileName) == 0 {
			err = fmt.Errorf("download failed to get affiliation file name: no file name is present")
			return echo.NewHTTPError(http.StatusInternalServerError, utils.Error{Error: err.Error()})
		}
		return c.Inline("files/"+affiliation.FileName, affiliation.FileName)
	case "d": //Document
		break
	case "g": // Gallery
		break
	case "l": // Player
		break
	case "n": // News
		break
	case "p": // Programme
		programme, err := r.controller.Session.GetProgrammeById(id)
		if err != nil {
			err = fmt.Errorf("download failed to get programme: %w", err)
			return echo.NewHTTPError(http.StatusInternalServerError, utils.Error{Error: err.Error()})
		}
		if len(programme.FileName) == 0 {
			err = fmt.Errorf("download failed to get programme file name: no file name is present")
			return echo.NewHTTPError(http.StatusInternalServerError, utils.Error{Error: err.Error()})
		}
		return c.Inline("files/"+programme.FileName, programme.FileName)
	case "s": // Sponsor
		sponsor, err := r.controller.Session.GetSponsorById(id)
		if err != nil {
			err = fmt.Errorf("download failed to get sponsor: %w", err)
			return echo.NewHTTPError(http.StatusInternalServerError, utils.Error{Error: err.Error()})
		}
		if len(sponsor.FileName) == 0 {
			err = fmt.Errorf("download failed to get sponsor file name: no file name is present")
			return echo.NewHTTPError(http.StatusInternalServerError, utils.Error{Error: err.Error()})
		}
		return c.Inline("files/"+sponsor.FileName, sponsor.FileName)
	case "t": // Team
		team, err := r.controller.Session.GetTeamById(id)
		if err != nil {
			err = fmt.Errorf("download failed to get team: %w", err)
			return echo.NewHTTPError(http.StatusInternalServerError, utils.Error{Error: err.Error()})
		}
		if len(team.FileName) == 0 {
			err = fmt.Errorf("download failed to get team file name: no file name is present")
			return echo.NewHTTPError(http.StatusInternalServerError, utils.Error{Error: err.Error()})
		}
		return c.Inline("files/"+team.FileName, team.FileName)
	case "u": // User
		break
	case "w": // WhatsOn
		break
	default:
		err = fmt.Errorf("download failed to get source: source format does not conform")
		return echo.NewHTTPError(http.StatusBadRequest, utils.Error{Error: err.Error()})
	}
	return nil
}
