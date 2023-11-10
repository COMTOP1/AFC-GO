package views

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/sponsor"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

type SponsorTemplate struct {
	ID      int
	Name    string
	Website null.String
	Purpose string
}

func (v *Views) SponsorsFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	var s1 []sponsor.Sponsor
	var err error

	s1, err = v.sponsor.GetSponsors(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get sponsors for sponsors: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year     int
		Sponsors []SponsorTemplate
		User     user.User
	}{
		Year:     year,
		Sponsors: DBSponsorsToTemplateFormat(s1),
		User:     c1.User,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.SponsorsTemplate, templates.RegularType)
}

func DBSponsorsToTemplateFormat(s1 []sponsor.Sponsor) []SponsorTemplate {
	sponsorsTemplate := make([]SponsorTemplate, 0, len(s1))
	for _, sponsorDB := range s1 {
		var s2 SponsorTemplate
		s2.ID = sponsorDB.ID
		s2.Name = sponsorDB.Name
		s2.Website = sponsorDB.Website
		s2.Purpose = sponsorDB.Purpose
		sponsorsTemplate = append(sponsorsTemplate, s2)
	}
	return sponsorsTemplate
}

func (v *Views) SponsorAddFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}

func (v *Views) SponsorDeleteFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}
