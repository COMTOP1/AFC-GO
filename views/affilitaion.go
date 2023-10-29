package views

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/affiliation"
)

func (v *Views) AffiliationAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		name := c.FormValue("name")
		website := c.FormValue("website")

		file, err := c.FormFile("image")
		if err != nil {
			return fmt.Errorf("failed to get file for affiliationAdd: %w", err)
		}
		fileName, err := v.fileUpload(file)
		if err != nil {
			return fmt.Errorf("failed to upload file for affiliationAdd: %w", err)
		}

		_, err = v.affiliation.AddAffiliation(c.Request().Context(), affiliation.Affiliation{Name: name, Website: null.StringFrom(website), FileName: null.StringFrom(fileName)})
		if err != nil {
			return fmt.Errorf("failed to add affiliation for affiliationAdd: %w", err)
		}

		return c.Redirect(http.StatusFound, "/")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

func (v *Views) AffiliationDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for affiliationDelete: %w", err)
		}

		affiliation1, err := v.affiliation.GetAffiliation(c.Request().Context(), affiliation.Affiliation{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get user for affiliationDelete: %w", err)
		}

		err = v.affiliation.DeleteAffiliation(c.Request().Context(), affiliation1)
		if err != nil {
			return fmt.Errorf("failed to delete user for affiliationDelete: %w", err)
		}
		return c.Redirect(http.StatusFound, "/")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}
