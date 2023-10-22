package views

import (
	"fmt"
	"log"
	"regexp"

	// importing time zones in case the system doesn't have them
	_ "time/tzdata"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/user"
)

type (
	// Context is a struct that is applied to the templates.
	Context struct {
		// TitleText is used for sending pages to the user with custom titles
		TitleText string
		// Message is used for sending a message back to the user trying to log in, might decide to move later as it may not be needed
		Message string
		// MsgType is the bulma.io class used to indicate what should be displayed
		MsgType string
		// Callback is the address to redirect the user to
		Callback string
		// User is the stored logged-in user
		User user.User
	}

	InternalContext struct {
		TitleText string
		Message   string
		MesType   string
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
			TitleText: i.TitleText,
			Message:   i.Message,
			MsgType:   i.MesType,
			Callback:  "/internal",
		}
		return c
	}

	var u user.User
	userValue := session.Values["user"]
	u, ok := userValue.(user.User)
	if !ok {
		u = user.User{Authenticated: false}
	}

	internalValue := session.Values["internalContext"]
	i, ok := internalValue.(InternalContext)
	if !ok {
		i = InternalContext{}
	}
	c := &Context{
		TitleText: i.TitleText,
		Message:   i.Message,
		MsgType:   i.MesType,
		Callback:  "/internal",
		User:      u,
	}
	return c
}

func (v *Views) setMessagesInSession(eC echo.Context, c *Context) error {
	session, err := v.cookie.Get(eC.Request(), v.conf.SessionCookieName)
	if err != nil {
		return fmt.Errorf("error getting session: %w", err)
	}
	session.Values["internalContext"] = InternalContext{
		TitleText: c.TitleText,
		Message:   c.Message,
		MesType:   c.MsgType,
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

// removeDuplicates removes all duplicate permissions
func removeDuplicate(strSlice []string) []string {
	allKeys := make(map[string]bool)
	var list []string
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			//if _, value := allKeys[item.PermissionID]; !value {
			//	allKeys[item.PermissionID] = true
			list = append(list, item)
		}
	}
	return list
}

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
