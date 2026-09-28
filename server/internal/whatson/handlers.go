package whatson

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /whatson.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the what's-on routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/whatson", h.list)
	g.GET("/whatson/:id", h.get)
	g.POST("/whatson", h.create, guards.Editor)
	g.PATCH("/whatson/:id", h.update, guards.Editor)
	g.DELETE("/whatson/:id", h.remove, guards.Editor)
}

// list returns events, optionally filtered by period.
//
//	@Summary	List what's on
//	@Tags		whatson
//	@Produce	json
//	@Param		period	query		string	false	"all (default), future or past"
//	@Success	200		{array}		whatson.Event
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/whatson [get]
func (h *Handlers) list(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context(), Period(c.QueryParam("period")))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// get returns one event.
//
//	@Summary	Get a what's-on event
//	@Tags		whatson
//	@Produce	json
//	@Param		id	path		int	true	"Event ID"
//	@Success	200	{object}	whatson.Event
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/whatson/{id} [get]
func (h *Handlers) get(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	e, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, e)
}

// create adds an event.
//
//	@Summary	Create a what's-on event
//	@Tags		whatson
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		title		formData	string	true	"Title"
//	@Param		content		formData	string	false	"HTML content (sanitised)"
//	@Param		dateOfEvent	formData	string	true	"YYYY-MM-DD"
//	@Param		image		formData	file	false	"Image"
//	@Success	201			{object}	whatson.Event
//	@Failure	422			{object}	web.ErrorResponse
//	@Router		/whatson [post]
func (h *Handlers) create(c echo.Context) error {
	if err := web.RequireForm(c); err != nil {
		return err
	}
	date, err := web.FormDate(c, "dateOfEvent")
	if err != nil {
		return err
	}
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	in := CreateInput{Title: c.FormValue("title"), Content: c.FormValue("content")}
	if date != nil {
		in.DateOfEvent = *date
	}
	e, err := h.svc.Create(c.Request().Context(), in, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, e)
}

// update changes an event; omitted fields are left as they are.
//
//	@Summary	Update a what's-on event
//	@Tags		whatson
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		id			path		int		true	"Event ID"
//	@Param		title		formData	string	false	"Title"
//	@Param		content		formData	string	false	"HTML content; empty clears it"
//	@Param		dateOfEvent	formData	string	false	"YYYY-MM-DD"
//	@Param		image		formData	file	false	"Replacement image"
//	@Param		removeImage	formData	bool	false	"Remove the current image"
//	@Success	200			{object}	whatson.Event
//	@Failure	404			{object}	web.ErrorResponse
//	@Failure	422			{object}	web.ErrorResponse
//	@Router		/whatson/{id} [patch]
func (h *Handlers) update(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	if err = web.RequireForm(c); err != nil {
		return err
	}
	var date *time.Time
	if date, err = web.FormDate(c, "dateOfEvent"); err != nil {
		return err
	}
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	remove, err := web.FormBool(c, "removeImage")
	if err != nil {
		return err
	}
	e, err := h.svc.Update(c.Request().Context(), id, UpdateInput{
		Title:       web.FormString(c, "title"),
		Content:     web.FormString(c, "content"),
		DateOfEvent: date,
		RemoveImage: remove != nil && *remove,
	}, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, e)
}

// remove deletes an event.
//
//	@Summary	Delete a what's-on event
//	@Tags		whatson
//	@Param		id	path	int	true	"Event ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/whatson/{id} [delete]
func (h *Handlers) remove(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	if _, err = h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
