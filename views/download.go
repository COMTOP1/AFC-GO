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
	source := c.QueryParam("s")

	temp := c.QueryParam("id")

	id, err := strconv.Atoi(temp)
	if err != nil {
		return fmt.Errorf("id must be a positive integer: %w", err)
	}

	if id < 1 {
		return fmt.Errorf("id must be a positive integer")
	}

	if len(source) != 1 {
		return fmt.Errorf("download failed to get source: source format does not conform: %s", source)
	}

	switch source {
	case "a": // Affiliation
		var affiliation affiliation1.Affiliation
		affiliation, err = v.affiliation.GetAffiliation(c.Request().Context(), affiliation1.Affiliation{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get affiliation: %w", err)
		}
		if len(affiliation.FileName.String) == 0 || !affiliation.FileName.Valid {
			return fmt.Errorf("download failed to get affiliation file name: no file name is present")
		}
		return c.Inline(filepath.Join(v.conf.FileDir, affiliation.FileName.String), affiliation.FileName.String)
	case "d": // Document
		var document document1.Document
		document, err = v.document.GetDocument(c.Request().Context(), document1.Document{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get document: %w", err)
		}
		if len(document.FileName) == 0 {
			return fmt.Errorf("download failed to get document file name: no file name is present")
		}
		if len(document.FileName) == 0 {
			return fmt.Errorf("download failed to get document file name: no file name is present")
		}
		return c.Inline(filepath.Join(v.conf.FileDir, document.FileName), document.FileName)
	case "g": // Gallery
		var image image1.Image
		image, err = v.image.GetImage(c.Request().Context(), image1.Image{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get image: %w", err)
		}
		if len(image.FileName.String) == 0 {
			return fmt.Errorf("download failed to get image file name: no file name is present")
		}
		return c.Inline(filepath.Join(v.conf.FileDir, image.FileName.String), image.FileName.String)
	case "l": // Player
		var player player1.Player
		player, err = v.player.GetPlayer(c.Request().Context(), player1.Player{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get player: %w", err)
		}
		if len(player.FileName.String) == 0 || !player.FileName.Valid {
			return fmt.Errorf("download failed to get player file name: no file name is present")
		}
		return c.Inline(filepath.Join(v.conf.FileDir, player.FileName.String), player.FileName.String)
	case "n": // News
		var news news1.News
		news, err = v.news.GetNewsArticle(c.Request().Context(), news1.News{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get news: %w", err)
		}
		if len(news.FileName.String) == 0 || !news.FileName.Valid {
			return fmt.Errorf("download failed to get news file name: no file name is present")
		}
		return c.Inline(filepath.Join(v.conf.FileDir, news.FileName.String), news.FileName.String)
	case "p": // Programme
		var programme programme1.Programme
		programme, err = v.programme.GetProgramme(c.Request().Context(), programme1.Programme{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get programme: %w", err)
		}
		if len(programme.FileName) == 0 {
			return fmt.Errorf("download failed to get programme file name: no file name is present")
		}
		return c.Inline(filepath.Join(v.conf.FileDir, programme.FileName), programme.FileName)
	case "s": // Sponsor
		var sponsor sponsor1.Sponsor
		sponsor, err = v.sponsor.GetSponsor(c.Request().Context(), sponsor1.Sponsor{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get sponsor: %w", err)
		}
		if len(sponsor.FileName.String) == 0 || !sponsor.FileName.Valid {
			return fmt.Errorf("download failed to get sponsor file name: no file name is present")
		}
		return c.Inline(filepath.Join(v.conf.FileDir, sponsor.FileName.String), sponsor.FileName.String)
	case "t": // Team
		var team team1.Team
		team, err = v.team.GetTeam(c.Request().Context(), team1.Team{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get team: %w", err)
		}
		if len(team.FileName.String) == 0 || !team.FileName.Valid {
			return fmt.Errorf("download failed to get team file name: no file name is present")
		}
		return c.Inline(filepath.Join(v.conf.FileDir, team.FileName.String), team.FileName.String)
	case "u": // User
		var user user1.User
		user, err = v.user.GetUser(c.Request().Context(), user1.User{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get user: %w", err)
		}
		if len(user.FileName.String) == 0 || !user.FileName.Valid {
			return fmt.Errorf("download failed to get user file name: no file name is present")
		}
		return c.Inline(filepath.Join(v.conf.FileDir, user.FileName.String), user.FileName.String)
	case "w": // WhatsOn
		var whatsOn whatson1.WhatsOn
		whatsOn, err = v.whatsOn.GetWhatsOnArticle(c.Request().Context(), whatson1.WhatsOn{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get what's on: %w", err)
		}
		if len(whatsOn.FileName.String) == 0 || !whatsOn.FileName.Valid {
			return fmt.Errorf("download failed to get what's on file name: no file name is present")
		}
		return c.Inline(filepath.Join(v.conf.FileDir, whatsOn.FileName.String), whatsOn.FileName.String)
	default:
		return fmt.Errorf("download failed to get source: source format does not conform")
	}
}
