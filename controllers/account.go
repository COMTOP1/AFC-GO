package controllers

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type AccountRepo struct {
	controller Controller
}

func NewAccountRepo(controller Controller) *AccountRepo {
	return &AccountRepo{
		controller: controller,
	}
}

func (r *AccountRepo) Account(c echo.Context) error {
	return c.JSON(http.StatusOK, "account")
}
