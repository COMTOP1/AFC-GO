package views

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"time"

	// importing time zones in case the system doesn't have them
	_ "time/tzdata"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/document"
	"github.com/COMTOP1/AFC-GO/news"
	"github.com/COMTOP1/AFC-GO/player"
	"github.com/COMTOP1/AFC-GO/programme"
	"github.com/COMTOP1/AFC-GO/role"
	"github.com/COMTOP1/AFC-GO/sponsor"
	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/user"
	"github.com/COMTOP1/AFC-GO/whatson"
)

type (
	// Context is a struct that is applied to the templates.
	Context struct {
		// Message is used for sending a message back to the user trying to log in, might decide to move later as it may not be needed
		Message string
		// MsgType is the bulma.io class used to indicate what should be displayed
		MsgType string
		// msgViewed is used to clear the message after it has been viewed once
		msgViewed bool
		// User is the stored logged-in user
		User user.User
	}

	InternalContext struct {
		Message   string
		MsgType   string
		msgViewed bool
	}

	ContactUserTemplate struct {
		ID    int
		Name  string
		Email string
		Role  string
	}

	DocumentTemplate struct {
		ID   int
		Name string
	}

	NewsTemplate struct {
		ID          int
		Title       string
		Content     string
		Date        string
		IsFileValid bool
	}

	PlayerTemplate struct {
		ID              int
		Name            string
		DateOfBirth     string
		DateOfBirthForm string
		IsFileValid     bool
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
	}

	SeasonTemplate struct {
		ID      int
		Name    string
		IsValid bool
	}

	SponsorTemplate struct {
		ID      int
		Name    string
		Website null.String
		Purpose string
	}

	TeamTemplate struct {
		ID       int
		Name     string
		IsActive bool
		IsYouth  bool
		IsValid  bool
	}

	UserTemplate struct {
		ID     int
		Name   string
		Email  string
		Phone  string
		TeamID int
		Role   string
	}

	WhatsOnTemplate struct {
		ID              int
		Title           string
		Content         string
		Date            string
		DateOfEvent     string
		DateOfEventForm string
		IsFileValid     bool
	}
)

func (v *Views) getSessionData(eC echo.Context) *Context {
	session, err := v.cookie.Get(eC.Request(), v.conf.SessionCookieName)
	if err != nil {
		log.Printf("error getting session: %+v", err)
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
		if len(u.TempRole) > 0 {
			u.Role, err = role.GetRole(u.TempRole)
			if err != nil {
				log.Printf("failed to get role for getSessionData: %+v", err)
			}
		} else {
			u.Role, err = role.GetRole(string(u.Role))
			if err != nil {
				log.Printf("failed to get role for getSessionData: %+v", err)
			}
		}
	}

	internalValue := session.Values["internalContext"]
	i, ok := internalValue.(InternalContext)
	if !ok {
		i = InternalContext{}
	}
	if i.msgViewed {
		err = v.clearMessagesInSession(eC)
		if err != nil {
			log.Printf("failed to clear message for getSessionData")
		}
		i.Message = ""
		i.MsgType = ""
	} else if len(i.Message) > 0 {
		err = v.setMessagesInSession(eC, &Context{
			Message:   i.Message,
			MsgType:   i.MsgType,
			msgViewed: true,
		})
		if err != nil {
			log.Printf("failed to set viewed message for getSessionData")
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
		log.Printf("error getting session: %+v", err)
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
		if len(u.TempRole) > 0 {
			u.Role, err = role.GetRole(u.TempRole)
			if err != nil {
				log.Printf("failed to get role for getSessionData: %+v", err)
			}
		} else {
			u.Role, err = role.GetRole(string(u.Role))
			if err != nil {
				log.Printf("failed to get role for getSessionData: %+v", err)
			}
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
		msgViewed: c.msgViewed,
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

//// removeDuplicates removes all duplicate permissions
// func removeDuplicate(strSlice []string) []string {
//	allKeys := make(map[string]bool)
//	var list []string
//	for _, item := range strSlice {
//		if _, value := allKeys[item]; !value {
//			allKeys[item] = true
//			//if _, value := allKeys[item.PermissionID]; !value {
//			//	allKeys[item.PermissionID] = true
//			list = append(list, item)
//		}
//	}
//	return list
//}

// minRequirementsMet tests if the password meets the minimum requirements
func minRequirementsMet(password string) (errString string) {
	var match bool
	match, err := regexp.MatchString("^.*[a-z].*$", password)
	if err != nil || !match {
		errString = "password must contain at least 1 lower case letter"
	}
	match, err = regexp.MatchString("^.*[A-Z].*$", password)
	if err != nil || !match {
		if len(errString) > 0 {
			errString += " and password must contain at least 1 upper case letter"
		} else {
			errString = "password must contain at least 1 upper case letter"
		}
	}
	match, err = regexp.MatchString("^.*\\d.*$", password)
	if err != nil || !match {
		if len(errString) > 0 {
			errString += " and password must contain at least 1 number"
		} else {
			errString = "password must contain at least 1 number"
		}
	}
	match, err = regexp.MatchString("^.*[@$!%*?&|^£;:/.,<>()_=+~§±#{}-].*$", password)
	if err != nil || !match {
		if len(errString) > 0 {
			errString += " and password must contain at least 1 special character"
		} else {
			errString = "password must contain at least 1 special character"
		}
	}
	if len(password) <= 8 {
		if len(errString) > 0 {
			errString += " and password must be at least 8 characters long"
		} else {
			errString = "password must be at least 8 characters long"
		}
	}
	return errString
}

func (v *Views) fileUpload(file *multipart.FileHeader) (string, error) {
	var fileName, fileType string
	switch file.Header.Get("content-type") {
	case "application/pdf":
		fileType = ".pdf"
		break
	case "image/apng":
		fileType = ".apng"
		break
	case "image/avif":
		fileType = ".avif"
		break
	case "image/gif":
		fileType = ".gif"
		break
	case "image/jpeg":
		fileType = ".jpg"
		break
	case "image/png":
		fileType = ".png"
		break
	case "image/svg+xml":
		fileType = ".svg"
		break
	case "image/webp":
		fileType = ".webp"
		break
	default:
		return "", fmt.Errorf("invalid image type: %s", file.Header.Get("content-type"))
	}

	fileName = uuid.NewString() + fileType

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file for fileUpload: %w", err)
	}
	defer src.Close()

	// Destination
	dst, err := os.Create(filepath.Join(v.conf.FileDir, fileName))
	if err != nil {
		return "", fmt.Errorf("failed to create file for fileUpload: %w", err)
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy contents to file for fileUpload: %w", err)
	}

	return fileName, nil
}

func DBDocumentsToTemplateFormat(documentsDB []document.Document) []DocumentTemplate {
	documentsTemplate := make([]DocumentTemplate, 0, len(documentsDB))
	for _, documentDB := range documentsDB {
		var documentTemplate DocumentTemplate
		documentTemplate.ID = documentDB.ID
		documentTemplate.Name = documentDB.Name
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
	return newsTemplate
}

func DBProgrammesToTemplateFormat(programmesDB []programme.Programme, seasonsDB []programme.Season) []ProgrammeTemplate {
	programmesTemplate := make([]ProgrammeTemplate, 0, len(programmesDB))
	for _, programmeDB := range programmesDB {
		var programmeTemplate ProgrammeTemplate
		programmeTemplate.ID = programmeDB.ID
		programmeTemplate.Name = programmeDB.Name
		year, month, day := programmeDB.DateOfProgramme.Date()
		programmeTemplate.DateOfProgramme = fmt.Sprintf("%s %02d %s %d", programmeDB.DateOfProgramme.Weekday().String()[0:3], day, month.String()[0:3], year)
		found := false
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
			log.Printf("failed to find season for programme: %d", programmeDB.ID)
			programmeTemplate.Season = SeasonTemplate{IsValid: false}
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
		sponsorsTemplate = append(sponsorsTemplate, sponsorTemplate)
	}
	return sponsorsTemplate
}

func DBManagersToTemplateFormat(managersDB []user.User) []string {
	managersString := make([]string, 0, len(managersDB))
	for _, manager := range managersDB {
		managersString = append(managersString, manager.Name)
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
				log.Printf("failed to parse player dateOfBirth: %d", playerDB.ID)
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
			log.Printf("failed to find team for player: %d, teamID set below 1: %d", playerDB.ID, playerDB.TeamID)
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
				log.Printf("failed to find team for player: %d", playerDB.ID)
				playerTemplate.Team = TeamTemplate{IsValid: false}
			}
		}
		playersTemplate = append(playersTemplate, playerTemplate)
	}
	return playersTemplate
}

func DBPlayersTeamToTemplateFormat(playersDB []player.Player) []PlayerTemplate {
	playersTemplate := make([]PlayerTemplate, 0, len(playersDB))
	for _, playerDB := range playersDB {
		var playerTemplate PlayerTemplate
		playerTemplate.ID = playerDB.ID
		playerTemplate.Name = playerDB.Name
		playerTemplate.Position = playerDB.Position
		playerTemplate.IsCaptain = playerDB.IsCaptain
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

func DBUsersToTemplateFormat(usersDB []user.User) []UserTemplate {
	usersTemplate := make([]UserTemplate, 0, len(usersDB))
	for _, userDB := range usersDB {
		var userTemplate UserTemplate
		userTemplate.ID = userDB.ID
		userTemplate.Name = userDB.Name
		userTemplate.Email = userDB.Email
		userTemplate.Phone = userDB.Phone
		if userDB.TeamID.Valid {
			userTemplate.TeamID = int(userDB.TeamID.Int64)
		}
		roleDB, err := role.GetRole(userDB.TempRole)
		userTemplate.Role = roleDB.String()
		if err != nil {
			userTemplate.Role = fmt.Sprintf("failed to get role for users: %+v", err)
		}
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
		temp, err := role.GetRole(userDB.TempRole)
		if err != nil {
			return nil, fmt.Errorf("failed to parse role for contactTemplate: %w", err)
		}
		userContactTemplate.Role = temp.String()
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
	whatsOnTemplate.Date = whatsOnDB.Date.Format("2006-01-02 15:04:05")
	year, month, day := whatsOnDB.DateOfEvent.Date()
	whatsOnTemplate.DateOfEvent = fmt.Sprintf("%s %02d %s %d", whatsOnDB.DateOfEvent.Weekday().String()[0:3], day, month.String()[0:3], year)
	whatsOnTemplate.DateOfEventForm = fmt.Sprintf("%02d/%02d/%04d", day, month, year)
	whatsOnTemplate.IsFileValid = whatsOnDB.FileName.Valid
	return whatsOnTemplate
}

// const DaysPerCycle = 146097
// const Days0000To1970 = int64(DaysPerCycle*5) - (int64(30)*int64(365) + int64(7))
//
// func ofEpochDay(epochDay int64) time.Time {
//	var zeroDay, adjust, adjustCycles, yearEst, doyEst int64
//	zeroDay = epochDay + Days0000To1970
//	// find the march-based year
//	zeroDay -= 60 // adjust to 0000-03-01 so leap day is at end of four year cycle
//	adjust = 0
//	if zeroDay < 0 {
//		// adjust negative years to positive for calculation
//		adjustCycles = (zeroDay+1)/DaysPerCycle - 1
//		adjust = adjustCycles * 400
//		zeroDay += -adjustCycles * DaysPerCycle
//	}
//	yearEst = (400*zeroDay + 591) / DaysPerCycle
//	doyEst = zeroDay - (365*yearEst + yearEst/4 - yearEst/100 + yearEst/400)
//	if doyEst < 0 {
//		// fix estimate
//		yearEst--
//		doyEst = zeroDay - (365*yearEst + yearEst/4 - yearEst/100 + yearEst/400)
//	}
//	yearEst += adjust // reset any negative year
//	marchDoy0 := int(doyEst)
//
//	// convert march-based values back to january-based
//
//	marchMonth0 := (marchDoy0*5 + 2) / 153
//
//	month := (marchMonth0+2)%12 + 1
//	dom := marchDoy0 - (marchMonth0*306+5)/10 + 1
//	yearEst += int64(marchMonth0) / 10
//
//	location, err := time.LoadLocation("Europe/London")
//	if err != nil {
//		log.Printf("failed to get timezone: %+v", err)
//	}
//
//	// check year now we are certain it is correct
//	return time.Date(int(yearEst), time.Month(month), dom, 0, 0, 0, 0, location)
// }
