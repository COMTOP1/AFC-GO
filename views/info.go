package views

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/microcosm-cc/bluemonday"

	"github.com/COMTOP1/AFC-GO/setting"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

func (v *Views) InfoFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	year, _, _ := time.Now().Date()

	infoContent, err := v.setting.GetSetting(c.Request().Context(), "infoContent")
	if err != nil {
		log.Printf("failed to get info content for info, error: %+v, continuing", err)
	}

	data := struct {
		Year        int
		InfoContent string
		User        user.User
	}{
		Year:        year,
		InfoContent: infoContent.SettingText,
		User:        c1.User,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.InfoTemplate, templates.RegularType)
}

func (v *Views) InfoEditFunc(c echo.Context) error {
	switch c.Request().Method {
	case http.MethodGet:
		return v._infoEditGet(c)
	case http.MethodPost:
		return v._infoEditPost(c)
	default:
		return c.String(http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (v *Views) _infoEditGet(c echo.Context) error {
	c1 := v.getSessionData(c)

	infoContent, err := v.setting.GetSetting(c.Request().Context(), "infoContent")
	if err != nil {
		log.Printf("failed to get infoContent for infoEditGet, error: %+v, continuing", err)
	}

	if len(infoContent.SettingText) == 0 {
		infoContent.SettingText = `<div><hcustom>Club History</hcustom><br></div><div>Our club was founded as AWRE Football Club in 1952 by Charles Green, Ted Hall, Gordon Carter, and Don Sharp, and were nicknamed "The Atom Men" after the newly founded Atomic Weapons Research Establishment. As the AWRE complex was still under construction when the club was established, the players used empty contractors' huts as changing rooms – carrying the tin bath into whichever building was designated for that day's football. The players were also responsible for the pitch, and would mow and mark out the pitch before the match started. The club received support from Sir William Penney during his work at AWRE on the Operation Hurricane project.</div><div><br></div><div>In the late 1960s/early 1970s, we changed our name, to AFC Aldermaston. In 1979, we were promoted from the Reading &amp; District League into Division One of the Hellenic League, where we spent the next seven seasons. From 1986 until 1991 we played in local football leagues, including the North Hampshire League, before joining Division Three of the Hampshire League in 1991. We finished fifth in the first season, earning promotion to Division Two. However, we were soon relegated back to Division Three the following season, where we remained until 1999. When the Hampshire League was re-organised prior to the 1999–2000 season, we were placed in the Premier Division.</div><div><br></div><div>When the Hampshire and Wessex Leagues merged in 2004, we became members of Division Two of the Wessex League. At the end of the 2004–05 season we suffered relegation. The division was renamed Division Two in 2006, and was disbanded at the end of the 2006–07 season, and as a result we moved up back to Division One. 2009–10 season, was our worst season and during a very difficult time we lost 40 consecutive games, which broke the previous record of 39 straight losses held jointly by Stockport United and Poole Town. The 40th defeat came on 8 April 2010, a 2–0 loss to Downton. This was undoubtedly the worst time for the club in living memory, and it took a great effort to keep going at this time. The losing streak thankfully came to an end on 10 April 2010, when we club drew 1–1 against Warminster Town, and then finally winning the next match against Petersfield Town 2–1 win for Aldermaston, to provide a much-needed boost for all of us at that difficult time.</div><div><br></div><div>Unsurprisingly, we finished bottom of Division One with just the one win in 40 matches, and were relegated to the Hampshire Premier League, where played for four seasons before moving to the Thames Valley Premier League for season 2014-15, when we reached the final of the Berks &amp; Bucks Intermediate Cup, which we narrowly lost 2-1 to Hale Leys Utd, the final of the Reading Senior Cup played at the Madjeski which we lost 1-0 to champions Marlow Utd, and the final of the Basingstoke Senior Cup where we lost to our local rivals Tadley Calleva 2-1 AET. 2015-16 season we finished seventh in the TVPL Premier Division, and again reaching the final of the Reading Senior Cup only to be beaten in a fourth final 1-0 to Cookham Dean.</div><div><br></div><div>We made a welcome return to Step 6 football as we were promoted to Division One East of the Hellenic League. 2016-17 saw us finish 7th in the Hellenic League this matched our highest ever finish at Step 6 football. 2017-18 season saw us finish 11th in the Hellenic League. 2018-19 season saw us finish 3rd in the Hellenic League this beat 2016-17 seasons highest ever finish. This was well deserved by everyone connected to the club. Also getting to the Reading Challenge Cup Final at the Madjeski Stadium which saw us lose 5-0 to Hellenic Premier League Winners Wantage Town and collecting the runners-up medals. We are hoping to improve on last season’s 3rd place finish for the 2019-20 season. Our 2019-20 season was expunged by the FA due to the COVID-19 outbreak at the point in time we were 14th place in the league. For our 2020-21 season we are aiming to improve on this position towards the top half of the table.</div><div><br></div><div><hcustom>Join Us</hcustom><br></div><div><a href="secretary@afcaldermaston.co.uk" style="text-decoration: underline;">secretary@afcaldermaston.co.uk</a></div><div><a href="mailto:safeguardingofficer@afcaldermaston.co.uk" style="text-decoration: underline;">safeguardingofficer@afcaldermaston.co.uk</a></div>`
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year        int
		InfoContent string
		User        user.User
		Context     *Context
	}{
		Year:        year,
		InfoContent: infoContent.SettingText,
		User:        c1.User,
		Context:     c1,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.InfoEditTemplate, templates.RegularType)
}

func (v *Views) _infoEditPost(c echo.Context) error {
	c1 := v.getSessionData(c)
	_ = c1

	content := c.FormValue("htmlContent")

	p := bluemonday.NewPolicy()
	// Common structural tags
	p.AllowElements("div", "br", "p", "blockquote", "pre", "hr")

	// Text formatting
	p.AllowElements("b", "i", "u", "strike")

	// Custom heading from your toolbar
	p.AllowElements("hcustom")

	// Lists
	p.AllowElements("ul", "ol", "li")

	// Links
	p.AllowElements("a")
	p.AllowAttrs("href", "style").OnElements("a")
	p.AllowURLSchemes("mailto", "http", "https")
	p.RequireNoFollowOnLinks(false)

	// Justification - via inline style
	p.AllowAttrs("style").OnElements("div", "p", "hcustom")

	safe := p.Sanitize(content)

	_, err := v.setting.GetSetting(c.Request().Context(), "infoContent")
	if err != nil {
		_, err = v.setting.AddSetting(c.Request().Context(), setting.Setting{
			ID:          "infoContent",
			SettingText: safe,
		})
		if err != nil {
			return fmt.Errorf("failed to add setting for info content, error: %w", err)
		}
	} else {
		_, err = v.setting.EditSetting(c.Request().Context(), setting.Setting{
			ID:          "infoContent",
			SettingText: safe,
		})
		if err != nil {
			return fmt.Errorf("failed to edit setting for info content, error: %w", err)
		}
	}
	return c.Redirect(http.StatusFound, "/info/edit")
}
