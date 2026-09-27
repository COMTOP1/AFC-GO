package document

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /documents.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the document routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/documents", h.list)
	g.POST("/documents", h.create, guards.Editor)
	g.DELETE("/documents/:id", h.remove, guards.Editor)
}

// list returns every document, by name.
//
//	@Summary	List documents
//	@Tags		documents
//	@Produce	json
//	@Success	200	{array}	document.Public
//	@Router		/documents [get]
func (h *Handlers) list(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// create uploads a document.
//
//	@Summary	Upload a document
//	@Tags		documents
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		name	formData	string	true	"Display name"
//	@Param		file	formData	file	true	"PDF, DOCX, PPTX, TXT or image"
//	@Success	201		{object}	document.Public
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/documents [post]
func (h *Handlers) create(c echo.Context) error {
	file, err := web.FormFile(c, "file")
	if err != nil {
		return err
	}
	out, err := h.svc.Create(c.Request().Context(), CreateInput{Name: c.FormValue("name")}, file)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}

// remove deletes a document.
//
//	@Summary	Delete a document
//	@Tags		documents
//	@Param		id	path	int	true	"Document ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/documents/{id} [delete]
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
