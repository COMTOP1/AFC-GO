package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

func (v *Views) DownloadFunc(c echo.Context) error {
	source := c.QueryParam("s")

	temp := c.QueryParam("id")

	id, err := strconv.Atoi(temp)
	if err != nil {
		return fmt.Errorf("id must be a positive integer: %w", err)
	}

	if id < 1 {
		return fmt.Errorf("id must be a positive intege, id: %s", temp)
	}

	if len(source) != 1 {
		return fmt.Errorf("download failed to get source: source format does not conform: %s", source)
	}

	switch source {
	case "a": // Affiliation
		var affiliation affiliation1.Affiliation
		affiliation, err = v.affiliation.GetAffiliation(c.Request().Context(), affiliation1.Affiliation{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get affiliation: %w, id: %d", err, id)
		}
		if len(affiliation.FileName.String) == 0 || !affiliation.FileName.Valid {
			return fmt.Errorf("download failed to get affiliation file name: no file name is present, id: %d", id)
		}
		path := filepath.Join(v.conf.FileDir, affiliation.FileName.String)
		_, err = os.Stat(path)
		if err != nil {
			if strings.Contains(err.Error(), "no such file") {
				return fmt.Errorf("failed to get file for affiliation download: no such file, id: %d", id)
			}
			return fmt.Errorf("failed to get file for affiliation download: %w, id: %d", err, id)
		}
		return c.Inline(path, affiliation.FileName.String)
	case "d": // Document
		var document document1.Document
		document, err = v.document.GetDocument(c.Request().Context(), document1.Document{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get document: %w, id: %d", err, id)
		}
		if len(document.FileName) == 0 {
			return fmt.Errorf("download failed to get document file name: no file name is present, id: %d", id)
		}
		path := filepath.Join(v.conf.FileDir, document.FileName)
		_, err = os.Stat(path)
		if err != nil {
			if strings.Contains(err.Error(), "no such file") {
				return fmt.Errorf("failed to get file for document download: no such file, id: %d", id)
			}
			return fmt.Errorf("failed to get file for document download: %w, id: %d", err, id)
		}
		return c.Attachment(path, fmt.Sprintf("AFC_%s.pdf", strings.ReplaceAll(document.Name, " ", "_")))
	case "g": // Gallery
		var image image1.Image
		image, err = v.image.GetImage(c.Request().Context(), image1.Image{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get image: %w, id: %d", err, id)
		}
		if len(image.FileName) == 0 {
			return fmt.Errorf("download failed to get image file name: no file name is present, id: %d", id)
		}
		path := filepath.Join(v.conf.FileDir, image.FileName)
		_, err = os.Stat(path)
		if err != nil {
			if strings.Contains(err.Error(), "no such file") {
				return fmt.Errorf("failed to get file for image download: no such file, id: %d", id)
			}
			return fmt.Errorf("failed to get file for image download: %w, id: %d", err, id)
		}
		return c.Inline(path, image.FileName)
	case "l": // Player
		var player player1.Player
		player, err = v.player.GetPlayer(c.Request().Context(), player1.Player{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get player: %w, id: %d", err, id)
		}
		var teamDB team1.Team
		teamDB, err = v.team.GetTeam(c.Request().Context(), team1.Team{ID: player.TeamID})
		if err != nil {
			return fmt.Errorf("download failed to get team for player: %w, id: %d", err, id)
		}
		if teamDB.IsYouth {
			return nil // Prevent image being downloaded if the team is a youth team
		}
		if player.DateOfBirth.Valid {
			today := time.Now().In(player.DateOfBirth.Time.Location())
			ty, tm, td := today.Date()
			today = time.Date(ty, tm, td, 0, 0, 0, 0, time.UTC)
			by, bm, bd := player.DateOfBirth.Time.Date()
			birthdate := time.Date(by, bm, bd, 0, 0, 0, 0, time.UTC)
			if today.Before(birthdate) {
				return fmt.Errorf("failed to parse player dateOfBirth: %d", player.ID)
			}
			age := ty - by
			anniversary := birthdate.AddDate(age, 0, 0)
			if anniversary.After(today) {
				age--
			}
			if age < 18 {
				return nil // Prevent image download if player is under 18
			}
		}
		if len(player.FileName.String) == 0 || !player.FileName.Valid {
			return fmt.Errorf("download failed to get player file name: no file name is present, id: %d", id)
		}
		path := filepath.Join(v.conf.FileDir, player.FileName.String)
		_, err = os.Stat(path)
		if err != nil {
			if strings.Contains(err.Error(), "no such file") {
				return fmt.Errorf("failed to get file for player download: no such file, id: %d", id)
			}
			return fmt.Errorf("failed to get file for player download: %w, id: %d", err, id)
		}
		return c.Inline(path, player.FileName.String)
	case "n": // News
		var news news1.News
		news, err = v.news.GetNewsArticle(c.Request().Context(), news1.News{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get news: %w, id: %d", err, id)
		}
		if len(news.FileName.String) == 0 || !news.FileName.Valid {
			return fmt.Errorf("download failed to get news file name: no file name is present, id: %d", id)
		}
		path := filepath.Join(v.conf.FileDir, news.FileName.String)
		_, err = os.Stat(path)
		if err != nil {
			if strings.Contains(err.Error(), "no such file") {
				return fmt.Errorf("failed to get file for news download: no such file, id: %d", id)
			}
			return fmt.Errorf("failed to get file for news download: %w, id: %d", err, id)
		}
		return c.Inline(path, news.FileName.String)
	case "p": // Programme
		var programme programme1.Programme
		programme, err = v.programme.GetProgramme(c.Request().Context(), programme1.Programme{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get programme: %w, id: %d", err, id)
		}
		if len(programme.FileName) == 0 {
			return fmt.Errorf("download failed to get programme file name: no file name is present, id: %d", id)
		}
		path := filepath.Join(v.conf.FileDir, programme.FileName)
		_, err = os.Stat(path)
		if err != nil {
			if strings.Contains(err.Error(), "no such file") {
				return fmt.Errorf("failed to get file for programme download: no such file, id: %d", id)
			}
			return fmt.Errorf("failed to get file for programme download: %w, id: %d", err, id)
		}
		return c.Inline(path, programme.FileName)
	case "s": // Sponsor
		var sponsor sponsor1.Sponsor
		sponsor, err = v.sponsor.GetSponsor(c.Request().Context(), sponsor1.Sponsor{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get sponsor: %w, id: %d", err, id)
		}
		if len(sponsor.FileName.String) == 0 || !sponsor.FileName.Valid {
			return fmt.Errorf("download failed to get sponsor file name: no file name is present, id: %d", id)
		}
		path := filepath.Join(v.conf.FileDir, sponsor.FileName.String)
		_, err = os.Stat(path)
		if err != nil {
			if strings.Contains(err.Error(), "no such file") {
				return fmt.Errorf("failed to get file for sponsor download: no such file, id: %d", id)
			}
			return fmt.Errorf("failed to get file for sponsor download: %w, id: %d", err, id)
		}
		return c.Inline(path, sponsor.FileName.String)
	case "t": // Team
		var team team1.Team
		team, err = v.team.GetTeam(c.Request().Context(), team1.Team{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get team: %w, id: %d", err, id)
		}
		if len(team.FileName.String) == 0 || !team.FileName.Valid {
			return fmt.Errorf("download failed to get team file name: no file name is present, id: %d", id)
		}
		path := filepath.Join(v.conf.FileDir, team.FileName.String)
		_, err = os.Stat(path)
		if err != nil {
			if strings.Contains(err.Error(), "no such file") {
				return fmt.Errorf("failed to get file for team download: no such file, id: %d", id)
			}
			return fmt.Errorf("failed to get file for team download: %w, id: %d", err, id)
		}
		return c.Inline(path, team.FileName.String)
	case "u": // User
		var user user1.User
		user, err = v.user.GetUser(c.Request().Context(), user1.User{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get user: %w, id: %d", err, id)
		}
		if len(user.FileName.String) == 0 || !user.FileName.Valid {
			return fmt.Errorf("download failed to get user file name: no file name is present, id: %d", id)
		}
		path := filepath.Join(v.conf.FileDir, user.FileName.String)
		_, err = os.Stat(path)
		if err != nil {
			if strings.Contains(err.Error(), "no such file") {
				return fmt.Errorf("failed to get file for user download: no such file, id: %d", id)
			}
			return fmt.Errorf("failed to get file for user download: %w, id: %d", err, id)
		}
		return c.Inline(path, user.FileName.String)
	case "w": // WhatsOn
		var whatsOn whatson1.WhatsOn
		whatsOn, err = v.whatsOn.GetWhatsOnArticle(c.Request().Context(), whatson1.WhatsOn{ID: id})
		if err != nil {
			return fmt.Errorf("download failed to get what's on: %w, id: %d", err, id)
		}
		if len(whatsOn.FileName.String) == 0 || !whatsOn.FileName.Valid {
			return fmt.Errorf("download failed to get what's on file name: no file name is present, id: %d", id)
		}
		path := filepath.Join(v.conf.FileDir, whatsOn.FileName.String)
		_, err = os.Stat(path)
		if err != nil {
			if strings.Contains(err.Error(), "no such file") {
				return fmt.Errorf("failed to get file for whatson download: no such file, id: %d", id)
			}
			return fmt.Errorf("failed to get file for whatson download: %w, id: %d", err, id)
		}
		return c.Inline(path, whatsOn.FileName.String)
	default:
		return fmt.Errorf("download failed to get source: source format does not conform, source: %s, id: %d", source, id)
	}
}
