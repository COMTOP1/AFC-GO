package views

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/labstack/echo/v4"

	affiliation1 "github.com/COMTOP1/AFC-GO/affiliation"
	document1 "github.com/COMTOP1/AFC-GO/document"
	image1 "github.com/COMTOP1/AFC-GO/image"
	news1 "github.com/COMTOP1/AFC-GO/news"
	player1 "github.com/COMTOP1/AFC-GO/player"
	programme1 "github.com/COMTOP1/AFC-GO/programme"
	sponsor1 "github.com/COMTOP1/AFC-GO/sponsor"
	team1 "github.com/COMTOP1/AFC-GO/team"
	user1 "github.com/COMTOP1/AFC-GO/user"
	whatson1 "github.com/COMTOP1/AFC-GO/whatson"
)

func (v *Views) Download(c echo.Context) error {
	source := c.Param("source")

	temp := c.Param("id")

	temp1 := []rune(temp)
	for _, r2 := range temp1 {
		if !unicode.IsNumber(r2) {
			return fmt.Errorf("id expects a positive number, the provided is not a positive number")
		}
	}
	id64, err := strconv.ParseUint(temp, 10, 64)
	if err != nil {
		return fmt.Errorf("download failed to get id: %w", err)
	}
	if len(source) != 1 {
		return fmt.Errorf("download failed to get source: source format does not conform: %s", source)
	}

	id := int(id64)

	switch source {
	case "a": // Affiliation
		var affiliation affiliation1.Affiliation
		affiliation, err = v.affiliation.GetAffiliation(c.Request().Context(), affiliation1.Affiliation{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get affiliation: %w", err)
		}
		if len(affiliation.FileName) == 0 {
			return fmt.Errorf("download failed to get affiliation file name: no file name is present")
		}
		return c.Inline("files/"+affiliation.FileName, affiliation.FileName)
	case "d": // Document
		var document document1.Document
		document, err = v.document.GetDocument(c.Request().Context(), document1.Document{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get document: %w", err)
		}
		if len(document.FileName) == 0 {
			return fmt.Errorf("download failed to get document file name: no file name is present")
		}
		return c.Inline("files/"+document.FileName, document.FileName)
	case "g": // Gallery
		var image image1.Image
		image, err = v.image.GetImage(c.Request().Context(), image1.Image{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get image: %w", err)
		}
		if len(image.FileName) == 0 {
			return fmt.Errorf("download failed to get image file name: no file name is present")
		}
		return c.Inline("files/"+image.FileName, image.FileName)
	case "l": // Player
		var player player1.Player
		player, err = v.player.GetPlayer(c.Request().Context(), news1.News{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get player: %w", err)
		}
		if len(player.FileName) == 0 {
			return fmt.Errorf("download failed to get player file name: no file name is present")
		}
		return c.Inline("files/"+player.FileName, player.FileName)
	case "n": // News
		var news news1.News
		news, err = v.news.GetNews(c.Request().Context(), news1.News{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get news: %w", err)
		}
		if len(news.FileName) == 0 {
			return fmt.Errorf("download failed to get news file name: no file name is present")
		}
		return c.Inline("files/"+news.FileName, news.FileName)
	case "p": // Programme
		var programme programme1.Programme
		programme, err = v.programme.GetProgramme(c.Request().Context(), programme1.Programme{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get programme: %w", err)
		}
		if len(programme.FileName) == 0 {
			return fmt.Errorf("download failed to get programme file name: no file name is present")
		}
		return c.Inline("files/"+programme.FileName, programme.FileName)
	case "s": // Sponsor
		var sponsor sponsor1.Sponsor
		sponsor, err = v.sponsor.GetSponsor(c.Request().Context(), sponsor1.Sponsor{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get sponsor: %w", err)
		}
		if len(sponsor.FileName) == 0 {
			return fmt.Errorf("download failed to get sponsor file name: no file name is present")
		}
		return c.Inline("files/"+sponsor.FileName, sponsor.FileName)
	case "t": // Team
		var team team1.Team
		team, err = v.team.GetTeam(c.Request().Context(), team1.Team{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get team: %w", err)
		}
		if len(team.FileName) == 0 {
			return fmt.Errorf("download failed to get team file name: no file name is present")
		}
		return c.Inline("files/"+team.FileName, team.FileName)
	case "u": // User
		var user user1.User
		user, err = v.user.GetUser(c.Request().Context(), user1.User{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get user: %w", err)
		}
		if len(user.FileName) == 0 {
			return fmt.Errorf("download failed to get user file name: no file name is present")
		}
		return c.Inline("files/"+user.FileName, user.FileName)
	case "w": // WhatsOn
		var whatsOn whatson1.WhatsOn
		whatsOn, err = v.user.GetUser(c.Request().Context(), whatson1.WhatsOn{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get what's on: %w", err)
		}
		if len(whatsOn.FileName) == 0 {
			return fmt.Errorf("download failed to get what's on file name: no file name is present")
		}
		return c.Inline("files/"+whatsOn.FileName, whatsOn.FileName)
	default:
		return fmt.Errorf("download failed to get source: source format does not conform")
	}
}
