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
}
