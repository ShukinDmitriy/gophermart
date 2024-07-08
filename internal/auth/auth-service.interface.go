package auth

import (
	"github.com/ShukinDmitriy/gophermart/internal/models"
	"github.com/labstack/echo/v4"
)

type AuthServiceInterface interface {
	GetUserID(c echo.Context) uint
	GenerateTokensAndSetCookies(c echo.Context, user *models.UserInfoResponse) error
}
