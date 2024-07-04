package auth

import (
	"errors"
	"github.com/ShukinDmitriy/gophermart/internal/models"
	"github.com/ShukinDmitriy/gophermart/internal/repositories"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"net/http"
)

type AuthUser struct {
	userRepository repositories.UserRepositoryInterface
}

func NewAuthUser(userRepository repositories.UserRepositoryInterface) *AuthUser {
	return &AuthUser{
		userRepository: userRepository,
	}
}

func (aUser *AuthUser) getUserIDByCookie(c echo.Context, cookie *http.Cookie) (*uint, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(GetJWTSecret()), nil
	})
	if errors.Is(err, jwt.ErrSignatureInvalid) {
		c.Logger().Error(err)
		return nil, err
	}

	return &claims.ID, nil
}

func (aUser *AuthUser) getUserByID(c echo.Context, id uint) *models.UserInfoResponse {
	user, err := aUser.userRepository.Find(id)
	if err != nil {
		c.Logger().Error(err)
		return nil
	}

	return user
}
