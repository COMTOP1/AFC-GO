package web

import "github.com/labstack/echo/v4"

// Guards are the authorisation middlewares handlers attach to routes. They
// are built by the auth package; web only defines the shape so domain
// packages don't import auth.
type Guards struct {
	Login               echo.MiddlewareFunc // any logged-in user
	Editor              echo.MiddlewareFunc // logged in, not Manager or Photographer
	NotManager          echo.MiddlewareFunc // logged in, not Manager (gallery)
	ClubSecretaryHigher echo.MiddlewareFunc // Safeguarding Officer, Club Secretary, Chairperson, Webmaster
	Identify            echo.MiddlewareFunc // loads the user if logged in; never rejects
}

// LoggedInKey is set on the request context once a guard has loaded a
// logged-in user.
const LoggedInKey = "web.loggedIn"

// LoggedIn reports whether a guard found a logged-in user for this request.
func LoggedIn(c echo.Context) bool {
	ok, _ := c.Get(LoggedInKey).(bool)
	return ok
}

// ActorKey is set on the request context once a guard has loaded a
// logged-in user, alongside LoggedInKey. It carries just enough of the user
// (id and email) for audit logging in domain packages that can't import
// auth (auth imports user, so user can't import auth back without a cycle).
const ActorKey = "web.actor"

// Actor is who performed a write, for audit logging.
type Actor struct {
	ID    int
	Email string
}

// CurrentActor returns the acting user for this request, set by a guard in
// the auth package once a session is loaded.
func CurrentActor(c echo.Context) (Actor, bool) {
	a, ok := c.Get(ActorKey).(Actor)
	return a, ok
}
