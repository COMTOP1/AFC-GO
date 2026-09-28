package auth

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves the /auth endpoints.
type Handlers struct {
	sessions *Sessions
	svc      *Service
	files    *upload.Files
}

func NewHandlers(sessions *Sessions, svc *Service, files *upload.Files) *Handlers {
	return &Handlers{sessions: sessions, svc: svc, files: files}
}

// Register mounts the /auth routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/auth/me", h.me, guards.Login)
	g.POST("/auth/login", h.login)
	g.POST("/auth/logout", h.logout, guards.Login)
	g.POST("/auth/password", h.password, guards.Login)
	g.GET("/auth/reset/:token", h.checkReset)
	g.POST("/auth/reset/:token", h.reset)
}

// me returns the logged-in user.
//
//	@Summary	Current user
//	@Tags		auth
//	@Produce	json
//	@Success	200	{object}	auth.CurrentUser
//	@Failure	401	{object}	web.ErrorResponse
//	@Router		/auth/me [get]
func (h *Handlers) me(c echo.Context) error {
	u, _ := Current(c)
	return c.JSON(http.StatusOK, NewCurrentUser(u, h.files))
}

// login starts a session, or returns a reset link for reset-flagged accounts.
//
//	@Summary	Log in
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		request	body		auth.LoginInput	true	"Credentials"
//	@Success	200		{object}	auth.LoginResponse
//	@Failure	401		{object}	web.ErrorResponse
//	@Router		/auth/login [post]
func (h *Handlers) login(c echo.Context) error {
	var in LoginInput
	if err := web.BindJSON(c, &in); err != nil {
		return err
	}
	res, err := h.svc.Login(c.Request().Context(), in.Email, in.Password)
	if errors.Is(err, ErrInvalidCredentials) {
		return echo.NewHTTPError(http.StatusUnauthorized, ErrInvalidCredentials.Error())
	}
	if err != nil {
		return err
	}
	if res.ResetRequired {
		return c.JSON(http.StatusOK, LoginResponse{ResetRequired: true, ResetURL: res.ResetURL})
	}
	if err = h.sessions.Login(c.Response(), c.Request(), res.User, in.Remember); err != nil {
		return err
	}
	me := NewCurrentUser(res.User, h.files)
	return c.JSON(http.StatusOK, LoginResponse{User: &me})
}

// logout ends the session.
//
//	@Summary	Log out
//	@Tags		auth
//	@Success	204
//	@Router		/auth/logout [post]
func (h *Handlers) logout(c echo.Context) error {
	if err := h.sessions.Logout(c.Response(), c.Request()); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// password changes the logged-in user's password.
//
//	@Summary	Change password
//	@Tags		auth
//	@Accept		json
//	@Param		request	body	auth.PasswordInput	true	"Old and new passwords"
//	@Success	204
//	@Failure	422	{object}	web.ErrorResponse
//	@Router		/auth/password [post]
func (h *Handlers) password(c echo.Context) error {
	var in PasswordInput
	if err := web.BindJSON(c, &in); err != nil {
		return err
	}
	u, _ := Current(c)
	if err := h.svc.ChangePassword(c.Request().Context(), u, in.OldPassword, in.NewPassword, in.ConfirmationPassword); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// checkReset reports whether a reset link is still valid.
//
//	@Summary	Check a reset link
//	@Tags		auth
//	@Param		token	path	string	true	"Reset token"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/auth/reset/{token} [get]
func (h *Handlers) checkReset(c echo.Context) error {
	if err := h.svc.CheckResetToken(c.Request().Context(), c.Param("token")); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// reset sets a new password with a reset link.
//
//	@Summary	Reset password
//	@Tags		auth
//	@Accept		json
//	@Param		token	path	string			true	"Reset token"
//	@Param		request	body	auth.ResetInput	true	"New password"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Failure	422	{object}	web.ErrorResponse
//	@Router		/auth/reset/{token} [post]
func (h *Handlers) reset(c echo.Context) error {
	var in ResetInput
	if err := web.BindJSON(c, &in); err != nil {
		return err
	}
	if err := h.svc.ResetPassword(c.Request().Context(), c.Param("token"), in.NewPassword, in.ConfirmationPassword); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
