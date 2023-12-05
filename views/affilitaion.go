package views

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/affiliation"
)

func (v *Views) AffiliationAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		name := c.FormValue("name")

		data := struct {
			Error string `json:"error"`
		}{
			Error: "",
		}

		website := c.FormValue("website")
		if len(website) > 0 {
			_, err := url.ParseRequestURI(website)
			if err != nil {
				log.Printf("failed to parse website for affiliationAdd: %+v", err)
				data.Error = fmt.Sprintf("failed to parse website for affiliationAdd: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}

		file, err := c.FormFile("upload")
		if err != nil {
			log.Printf("failed to get file for affiliationAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to get file for affiliationAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}
		fileName, err := v.fileUpload(file)
		if err != nil {
			log.Printf("failed to upload file for affiliationAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to upload file for affiliationAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		_, err = v.affiliation.AddAffiliation(c.Request().Context(), affiliation.Affiliation{Name: name, Website: null.NewString(website, len(website) > 0), FileName: null.NewString(fileName, len(fileName) > 0)})
		if err != nil {
			log.Printf("failed to add affiliation for affiliationAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to add affiliation for affiliationAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully added \"%s\"", name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for affiliationAdd: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

func (v *Views) AffiliationDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for affiliationDelete: %w", err)
		}

		affiliationDB, err := v.affiliation.GetAffiliation(c.Request().Context(), affiliation.Affiliation{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get affiliaiton for affiliationDelete: %w", err)
		}

		if affiliationDB.FileName.Valid {
			err = os.Remove(filepath.Join(v.conf.FileDir, affiliationDB.FileName.String))
			if err != nil {
				return fmt.Errorf("failed to delete affiliation image for affiliationDelete: %w", err)
			}
		}

		err = v.affiliation.DeleteAffiliation(c.Request().Context(), affiliationDB)
		if err != nil {
			return fmt.Errorf("failed to delete affiliaiton for affiliationDelete: %w", err)
		}
		return c.Redirect(http.StatusFound, "/")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}
