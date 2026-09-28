package views

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	// importing time zones in case the system doesn't have them
	_ "time/tzdata"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/document"
	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/programme"
	"github.com/COMTOP1/AFC-GO/server/internal/role"
	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

type (
	// Context is a struct applied to the templates.
	Context struct {
		// Message is used for sending a message back to the user trying to log in, might decide to move later as it may not be needed
		Message string
		// MsgType is the bulma.io class used to indicate what should be displayed
		MsgType string
		// MsgViewed is used to clear the message after it has been viewed once
		MsgViewed bool
		// User is the stored logged-in user
		User user.User
	}

	InternalContext struct {
		Message   string
		MsgType   string
		MsgViewed bool
	}

	ContactUserTemplate struct {
		ID       int
		Name     string
		Email    string
		Role     string
		FileName null.String
	}

	DocumentTemplate struct {
		ID       int
		Name     string
		FileName string
	}

	NewsTemplate struct {
		ID          int
		Title       string
		Content     string
		Date        string
		IsFileValid bool
		FileName    null.String
	}

	ManagerTemplate struct {
		Name  string
		Email string
	}

	PlayerTemplate struct {
		ID              int
		Name            string
		DateOfBirth     string
		DateOfBirthForm string
		IsFileValid     bool
		FileName        null.String
		Age             int
		Position        null.String
		IsCaptain       bool
		Team            TeamTemplate
	}

	ProgrammeTemplate struct {
		ID              int
		Name            string
		DateOfProgramme string
		Season          SeasonTemplate
		FileName        string
	}

	SeasonTemplate struct {
		ID      int
		Name    string
		IsValid bool
	}

	SponsorTemplate struct {
		ID       int
		Name     string
		Website  null.String
		Purpose  null.String
		FileName null.String
	}

	TeamTemplate struct {
		ID       int
		Name     string
		IsActive bool
		IsYouth  bool
		IsValid  bool
	}

	UserTemplate struct {
		ID           int
		Name         string
		Email        string
		Phone        string
		TeamID       int
		Role         string
		RoleTemplate string
		IsFileValid  bool
		FileName     null.String
	}

	WhatsOnTemplate struct {
		ID              int
		Title           string
		Content         string
		Date            string
		DateOfEvent     string
		DateOfEventForm string
		IsFileValid     bool
		FileName        null.String
	}
)

func (v *Views) getSessionData(eC echo.Context) *Context {
	session, err := v.cookie.Get(eC.Request(), v.conf.SessionCookieName)
	if err != nil {
		slog.Info(fmt.Sprintf("error getting session: %+v", err))
		err = session.Save(eC.Request(), eC.Response())
		if err != nil {
			panic(fmt.Errorf("failed to save user session for getSessionData: %w", err))
		}
		i := InternalContext{}
		c := &Context{
			Message: i.Message,
			MsgType: i.MsgType,
		}
		return c
	}

	var u user.User
	userValue := session.Values["user"]
	u, ok := userValue.(user.User)
	if !ok {
		u = user.User{Authenticated: false}
	} else {
		_, err = role.GetRole(string(u.Role))
		if err != nil {
			slog.Info(fmt.Sprintf("failed to get role for getSessionData: %+v", err))
		}
	}

	internalValue := session.Values["internalContext"]
	i, ok := internalValue.(InternalContext)
	if !ok {
		i = InternalContext{}
	}
	if i.MsgViewed {
		err = v.clearMessagesInSession(eC)
		if err != nil {
			slog.Info("failed to clear message for getSessionData")
		}
		i.Message = ""
		i.MsgType = ""
	} else if len(i.Message) > 0 {
		err = v.setMessagesInSession(eC, &Context{
			Message:   i.Message,
			MsgType:   i.MsgType,
			MsgViewed: true,
		})
		if err != nil {
			slog.Info("failed to set viewed message for getSessionData")
		}
	}
	c := &Context{
		Message: i.Message,
		MsgType: i.MsgType,
		User:    u,
	}
	return c
}

func (v *Views) getSessionDataNoMsg(eC echo.Context) *Context {
	session, err := v.cookie.Get(eC.Request(), v.conf.SessionCookieName)
	if err != nil {
		slog.Info(fmt.Sprintf("error getting session: %+v", err))
		err = session.Save(eC.Request(), eC.Response())
		if err != nil {
			panic(fmt.Errorf("failed to save user session for getSessionData: %w", err))
		}
		i := InternalContext{}
		c := &Context{
			Message: i.Message,
			MsgType: i.MsgType,
		}
		return c
	}

	var u user.User
	userValue := session.Values["user"]
	u, ok := userValue.(user.User)
	if !ok {
		u = user.User{Authenticated: false}
	} else {
		_, err = role.GetRole(string(u.Role))
		if err != nil {
			slog.Info(fmt.Sprintf("failed to get role for getSessionData: %+v", err))
		}
	}

	c := &Context{
		User: u,
	}
	return c
}

func (v *Views) setMessagesInSession(eC echo.Context, c *Context) error {
	session, err := v.cookie.Get(eC.Request(), v.conf.SessionCookieName)
	if err != nil {
		return fmt.Errorf("error getting session: %w", err)
	}
	session.Values["internalContext"] = InternalContext{
		Message:   c.Message,
		MsgType:   c.MsgType,
		MsgViewed: c.MsgViewed,
	}

	err = session.Save(eC.Request(), eC.Response())
	if err != nil {
		return fmt.Errorf("failed to save session for set message: %w", err)
	}
	return nil
}

func (v *Views) clearMessagesInSession(eC echo.Context) error {
	session, err := v.cookie.Get(eC.Request(), v.conf.SessionCookieName)
	if err != nil {
		return fmt.Errorf("error getting session: %w", err)
	}
	session.Values["internalContext"] = InternalContext{}

	err = session.Save(eC.Request(), eC.Response())
	if err != nil {
		return fmt.Errorf("failed to save session for clear message: %w", err)
	}
	return nil
}

func DBDocumentsToTemplateFormat(documentsDB []document.Document) []DocumentTemplate {
	documentsTemplate := make([]DocumentTemplate, 0, len(documentsDB))
	for _, documentDB := range documentsDB {
		var documentTemplate DocumentTemplate
		documentTemplate.ID = documentDB.ID
		documentTemplate.Name = documentDB.Name
		documentTemplate.FileName = documentDB.FileName
		documentsTemplate = append(documentsTemplate, documentTemplate)
	}
	return documentsTemplate
}

func DBNewsToTemplateFormat(newsDB []news.News) []NewsTemplate {
	newsTemplate := make([]NewsTemplate, 0, len(newsDB))
	for _, newsArticleDB := range newsDB {
		var newsArticleTemplate NewsTemplate
		newsArticleTemplate.ID = newsArticleDB.ID
		newsArticleTemplate.Title = newsArticleDB.Title
		year, month, day := newsArticleDB.Date.Date()
		newsArticleTemplate.Date = fmt.Sprintf("%s %02d %s %d - %s", newsArticleDB.Date.Weekday().String()[0:3], day, month.String()[0:3], year, newsArticleDB.Date.Format("15:04:05"))
		newsArticleTemplate.FileName = newsArticleDB.FileName
		newsTemplate = append(newsTemplate, newsArticleTemplate)
	}
	return newsTemplate
}

func DBNewsToArticleTemplateFormat(newsDB news.News) NewsTemplate {
	var newsTemplate NewsTemplate
	newsTemplate.ID = newsDB.ID
	newsTemplate.Title = newsDB.Title
	newsTemplate.Content = newsDB.Content.String
	year, month, day := newsDB.Date.Date()
	newsTemplate.Date = fmt.Sprintf("%s %02d %s %d - %s", newsDB.Date.Weekday().String()[0:3], day, month.String()[0:3], year, newsDB.Date.Format("15:04:05"))
	newsTemplate.IsFileValid = newsDB.FileName.Valid
	newsTemplate.FileName = newsDB.FileName
	return newsTemplate
}

func DBProgrammesToTemplateFormat(programmesDB []programme.Programme, seasonsDB []programme.Season) []ProgrammeTemplate {
	programmesTemplate := make([]ProgrammeTemplate, 0, len(programmesDB))
	for _, programmeDB := range programmesDB {
		var programmeTemplate ProgrammeTemplate
		programmeTemplate.ID = programmeDB.ID
		programmeTemplate.Name = programmeDB.Name
		programmeTemplate.FileName = programmeDB.FileName
		year, month, day := programmeDB.DateOfProgramme.Date()
		programmeTemplate.DateOfProgramme = fmt.Sprintf("%s %02d %s %d", programmeDB.DateOfProgramme.Weekday().String()[0:3], day, month.String()[0:3], year)
		found := false
		if programmeDB.SeasonID != 0 {
			for _, seasonDB := range seasonsDB {
				if seasonDB.ID == programmeDB.SeasonID {
					var seasonTemplate SeasonTemplate
					seasonTemplate.ID = seasonDB.ID
					seasonTemplate.Name = seasonDB.Season
					seasonTemplate.IsValid = true
					programmeTemplate.Season = seasonTemplate
					found = true
					break
				}
			}
			if !found {
				slog.Info(fmt.Sprintf("failed to find season for programme: %d", programmeDB.ID))
				programmeTemplate.Season = SeasonTemplate{IsValid: false}
			}
		}
		programmesTemplate = append(programmesTemplate, programmeTemplate)
	}
	return programmesTemplate
}

func DBSponsorsToTemplateFormat(sponsorsDB []sponsor.Sponsor) []SponsorTemplate {
	sponsorsTemplate := make([]SponsorTemplate, 0, len(sponsorsDB))
	for _, sponsorDB := range sponsorsDB {
		var sponsorTemplate SponsorTemplate
		sponsorTemplate.ID = sponsorDB.ID
		sponsorTemplate.Name = sponsorDB.Name
		sponsorTemplate.Website = sponsorDB.Website
		sponsorTemplate.Purpose = sponsorDB.Purpose
		sponsorTemplate.FileName = sponsorDB.FileName
		sponsorsTemplate = append(sponsorsTemplate, sponsorTemplate)
	}
	return sponsorsTemplate
}

func DBManagersToTemplateFormat(managersDB []user.User) []ManagerTemplate {
	managersString := make([]ManagerTemplate, 0, len(managersDB))
	for _, manager := range managersDB {
		m := ManagerTemplate{
			Name:  manager.Name,
			Email: manager.Email,
		}
		managersString = append(managersString, m)
	}
	return managersString
}

func DBPlayersToTemplateFormat(playersDB []player.Player, teamsDB []team.Team) []PlayerTemplate {
	playersTemplate := make([]PlayerTemplate, 0, len(playersDB))
	for _, playerDB := range playersDB {
		var playerTemplate PlayerTemplate
		playerTemplate.ID = playerDB.ID
		playerTemplate.Name = playerDB.Name
		playerTemplate.DateOfBirth = "Not provided"
		if playerDB.DateOfBirth.Valid {
			year, month, day := playerDB.DateOfBirth.Time.Date()
			playerTemplate.DateOfBirth = fmt.Sprintf("%s %02d %s %d", playerDB.DateOfBirth.Time.Weekday().String()[0:3], day, month.String()[0:3], year)
			today := time.Now().In(playerDB.DateOfBirth.Time.Location())
			ty, tm, td := today.Date()
			today = time.Date(ty, tm, td, 0, 0, 0, 0, time.UTC)
			by, bm, bd := playerDB.DateOfBirth.Time.Date()
			birthdate := time.Date(by, bm, bd, 0, 0, 0, 0, time.UTC)
			if today.Before(birthdate) {
				slog.Info(fmt.Sprintf("failed to parse player dateOfBirth: %d", playerDB.ID))
				playerTemplate.Age = -1
			} else {
				age := ty - by
				anniversary := birthdate.AddDate(age, 0, 0)
				if anniversary.After(today) {
					age--
				}
				playerTemplate.Age = age
				playerTemplate.DateOfBirthForm = fmt.Sprintf("%02d/%02d/%04d", bd, bm, by)
			}
		} else {
			playerTemplate.Age = -1
		}
		if len(playerDB.FileName.String) > 0 && playerDB.FileName.Valid {
			playerTemplate.IsFileValid = true
		}
		playerTemplate.Position = playerDB.Position
		playerTemplate.IsCaptain = playerDB.IsCaptain
		found := false
		if playerDB.TeamID < 0 {
			slog.Info(fmt.Sprintf("failed to find team for player: %d, teamID set below 1: %d", playerDB.ID, playerDB.TeamID))
			playerTemplate.Team = TeamTemplate{IsValid: false}
		} else {
			for _, teamDB := range teamsDB {
				if teamDB.ID == playerDB.TeamID {
					var teamTemplate TeamTemplate
					teamTemplate.ID = teamDB.ID
					teamTemplate.Name = teamDB.Name
					teamTemplate.IsYouth = teamDB.IsYouth
					teamTemplate.IsValid = true
					playerTemplate.Team = teamTemplate
					found = true
					break
				}
			}
			if !found {
				slog.Info(fmt.Sprintf("failed to find team for player: %d", playerDB.ID))
				playerTemplate.Team = TeamTemplate{IsValid: false}
			}
		}
		// A player whose team couldn't be found is treated as youth: we can't
		// prove otherwise, so fail closed (Review Focus #1).
		if player.PhotoVisible(playerDB, !playerTemplate.Team.IsValid || playerTemplate.Team.IsYouth, time.Now()) {
			playerTemplate.FileName = playerDB.FileName
		}
		playersTemplate = append(playersTemplate, playerTemplate)
	}
	return playersTemplate
}

func DBPlayersTeamToTemplateFormat(playersDB []player.Player, isYouthTeam bool) []PlayerTemplate {
	playersTemplate := make([]PlayerTemplate, 0, len(playersDB))
	for _, playerDB := range playersDB {
		var playerTemplate PlayerTemplate
		playerTemplate.ID = playerDB.ID
		playerTemplate.Name = playerDB.Name
		playerTemplate.Position = playerDB.Position
		playerTemplate.IsCaptain = playerDB.IsCaptain
		if len(playerDB.FileName.String) > 0 && playerDB.FileName.Valid {
			playerTemplate.IsFileValid = true
		}
		if player.PhotoVisible(playerDB, isYouthTeam, time.Now()) {
			playerTemplate.FileName = playerDB.FileName
		}
		playersTemplate = append(playersTemplate, playerTemplate)
	}
	return playersTemplate
}

func DBTeamsToTemplateFormat(teamsDB []team.Team) []TeamTemplate {
	teamsTemplate := make([]TeamTemplate, 0, len(teamsDB))
	for _, teamDB := range teamsDB {
		var teamTemplate TeamTemplate
		teamTemplate.ID = teamDB.ID
		teamTemplate.Name = teamDB.Name
		teamTemplate.IsActive = teamDB.IsActive
		teamTemplate.IsYouth = teamDB.IsYouth
		teamsTemplate = append(teamsTemplate, teamTemplate)
	}
	return teamsTemplate
}

func DBUserToTemplateFormat(userDB user.User) UserTemplate {
	var userTemplate UserTemplate
	userTemplate.ID = userDB.ID
	userTemplate.Name = userDB.Name
	userTemplate.Email = userDB.Email
	userTemplate.Phone = "No number provided"
	if userDB.Phone.Valid {
		userTemplate.Phone = userDB.Phone.String
	}
	userTemplate.TeamID = userDB.TeamID
	userTemplate.Role = userDB.Role.String()
	if len(userDB.FileName.String) > 0 && userDB.FileName.Valid {
		userTemplate.IsFileValid = true
	}
	userTemplate.FileName = userDB.FileName
	return userTemplate
}

func DBUsersToTemplateFormat(usersDB []user.User) []UserTemplate {
	usersTemplate := make([]UserTemplate, 0, len(usersDB))
	for _, userDB := range usersDB {
		var userTemplate UserTemplate
		userTemplate.ID = userDB.ID
		userTemplate.Name = userDB.Name
		userTemplate.Email = userDB.Email
		userTemplate.Phone = "No number provided"
		if userDB.Phone.Valid {
			userTemplate.Phone = userDB.Phone.String
		}
		userTemplate.TeamID = userDB.TeamID
		userTemplate.Role = userDB.Role.String()
		userTemplate.RoleTemplate = strings.ToLower(userDB.Role.DBString())
		if len(userDB.FileName.String) > 0 && userDB.FileName.Valid {
			userTemplate.IsFileValid = true
		}
		userTemplate.FileName = userDB.FileName
		usersTemplate = append(usersTemplate, userTemplate)
	}
	return usersTemplate
}

func DBUsersContactToTemplateFormat(usersDB []user.User) ([]ContactUserTemplate, error) {
	usersContactTemplate := make([]ContactUserTemplate, 0, len(usersDB))
	for _, userDB := range usersDB {
		var userContactTemplate ContactUserTemplate
		userContactTemplate.ID = userDB.ID
		userContactTemplate.Name = userDB.Name
		userContactTemplate.Email = userDB.Email
		userContactTemplate.Role = userDB.Role.String()
		userContactTemplate.FileName = userDB.FileName
		usersContactTemplate = append(usersContactTemplate, userContactTemplate)
	}
	return usersContactTemplate, nil
}

func DBWhatsOnToTemplateFormat(whatsOnsDB []whatson.WhatsOn) []WhatsOnTemplate {
	whatsOnsTemplate := make([]WhatsOnTemplate, 0, len(whatsOnsDB))
	for _, whatsOnDB := range whatsOnsDB {
		var whatsOnTemplate WhatsOnTemplate
		whatsOnTemplate.ID = whatsOnDB.ID
		whatsOnTemplate.Title = whatsOnDB.Title
		whatsOnTemplate.Date = whatsOnDB.Date.Format("2006-01-02 15:04:05")
		year, month, day := whatsOnDB.DateOfEvent.Date()
		whatsOnTemplate.DateOfEvent = fmt.Sprintf("%s %02d %s %d", whatsOnDB.DateOfEvent.Weekday().String()[0:3], day, month.String()[0:3], year)
		whatsOnTemplate.Date = fmt.Sprintf("%s %02d %s %d - %s", whatsOnDB.Date.Weekday().String()[0:3], day, month.String()[0:3], year, whatsOnDB.Date.Format("15:04:05"))
		whatsOnTemplate.FileName = whatsOnDB.FileName
		whatsOnsTemplate = append(whatsOnsTemplate, whatsOnTemplate)
	}
	return whatsOnsTemplate
}

func DBWhatsOnToArticleTemplateFormat(whatsOnDB whatson.WhatsOn) WhatsOnTemplate {
	var whatsOnTemplate WhatsOnTemplate
	whatsOnTemplate.ID = whatsOnDB.ID
	whatsOnTemplate.Title = whatsOnDB.Title
	if whatsOnDB.Content.Valid {
		whatsOnTemplate.Content = whatsOnDB.Content.String
	}
	year, month, day := whatsOnDB.DateOfEvent.Date()
	whatsOnTemplate.Date = fmt.Sprintf("%s %02d %s %d - %s", whatsOnDB.Date.Weekday().String()[0:3], day, month.String()[0:3], year, whatsOnDB.Date.Format("15:04:05"))
	whatsOnTemplate.DateOfEvent = fmt.Sprintf("%s %02d %s %d", whatsOnDB.DateOfEvent.Weekday().String()[0:3], day, month.String()[0:3], year)
	whatsOnTemplate.DateOfEventForm = fmt.Sprintf("%02d/%02d/%04d", day, month, year)
	whatsOnTemplate.IsFileValid = whatsOnDB.FileName.Valid
	whatsOnTemplate.FileName = whatsOnDB.FileName
	return whatsOnTemplate
}

// legacyUpload returns the optional file in field, or nil when none was sent.
//
//nolint:unparam // every legacy form names its file field "upload"; kept explicit at call sites
func legacyUpload(c echo.Context, field string) (*upload.File, error) {
	fh, err := c.FormFile(field)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return nil, nil //nolint:nilerr // uploads are optional
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	return upload.FromHeader(fh), nil
}

// flash queues a success message for the next page view.
func (v *Views) flash(c echo.Context, c1 *Context, message string) {
	c1.Message = message
	c1.MsgType = "is-success"
	if err := v.setMessagesInSession(c, c1); err != nil {
		slog.Info(fmt.Sprintf("failed to set flash message: %+v", err))
	}
}

// formYes maps a legacy "Y" checkbox to a bool; anything else is false.
func formYes(c echo.Context, field string) bool {
	return c.FormValue(field) == "Y"
}
