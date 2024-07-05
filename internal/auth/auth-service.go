package auth

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ShukinDmitriy/gophermart/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	userTokenCookieName    = "user"
	accessTokenCookieName  = "access-token"
	refreshTokenCookieName = "refresh-token"
)

type Claims struct {
	ID uint `json:"id"`
	jwt.RegisteredClaims
}

type AuthService struct {
	authUser AuthUser
}

func NewAuthService(authUser AuthUser) *AuthService {
	return &AuthService{authUser: authUser}
}

func GetJWTSecret() string {
	jwtSecretKey, exists := os.LookupEnv("JWT_SECRET_KEY")

	if !exists {
		panic("no JWT_SECRET_KEY in .env")
	}
	return jwtSecretKey
}

func (authService *AuthService) GetAccessTokenCookieName() string {
	return accessTokenCookieName
}

func (authService *AuthService) GetRefreshTokenCookieName() string {
	return refreshTokenCookieName
}

func (authService *AuthService) BeforeFunc(c echo.Context) {
	accessTokenCookie, err := c.Cookie(authService.GetAccessTokenCookieName())
	if err == nil {
		accessUserID, _ := authService.authUser.getUserIDByCookie(c, accessTokenCookie)
		if accessUserID != nil {
			return
		}
	}

	refreshTokenCookie, err := c.Cookie(authService.GetRefreshTokenCookieName())
	if err != nil {
		return
	}

	if accessTokenCookie == nil && refreshTokenCookie != nil {
		userID, err := authService.authUser.getUserIDByCookie(c, refreshTokenCookie)
		if err != nil {
			return
		}

		user := authService.authUser.getUserByID(c, *userID)

		err = authService.GenerateTokensAndSetCookies(c, user)
		if err != nil {
			return
		}
	}
}

func (authService *AuthService) GenerateTokensAndSetCookies(c echo.Context, user *models.UserInfoResponse) error {
	accessToken, accessTokenString, exp, err := authService.generateAccessToken(user)
	if err != nil {
		return err
	}

	_, refreshTokenString, refreshExp, err := authService.generateRefreshToken(user)
	if err != nil {
		return err
	}

	authService.setTokenCookie(c, accessTokenCookieName, accessTokenString, exp)
	authService.setTokenCookie(c, refreshTokenCookieName, refreshTokenString, refreshExp)
	c.Set("user", accessToken)
	authService.setUserCookie(c, user, exp)

	return nil
}

func (authService *AuthService) GetUserID(c echo.Context) uint {
	if c.Get("user") == nil {
		return 0
	}
	u, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return 0
	}

	claims, ok := u.Claims.(*Claims)
	if !ok {
		return 0
	}

	return claims.ID
}

func (authService *AuthService) JWTErrorChecker(c echo.Context, err error) error {
	if err != nil {
		zap.L().Error(
			"JWTErrorChecker",
			zap.Error(err),
		)
	}

	return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
}

func (authService *AuthService) generateAccessToken(user *models.UserInfoResponse) (*jwt.Token, string, time.Time, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	return authService.generateToken(user, expirationTime, []byte(GetJWTSecret()))
}

func (authService *AuthService) generateRefreshToken(user *models.UserInfoResponse) (*jwt.Token, string, time.Time, error) {
	expirationTime := time.Now().Add(30 * 24 * time.Hour)

	return authService.generateToken(user, expirationTime, []byte(GetJWTSecret()))
}

func (authService *AuthService) generateToken(user *models.UserInfoResponse, expirationTime time.Time, secret []byte) (*jwt.Token, string, time.Time, error) {
	claims := &Claims{
		ID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return nil, "", time.Now(), err
	}

	return token, tokenString, expirationTime, nil
}

func (authService *AuthService) setTokenCookie(c echo.Context, name, token string, expiration time.Time) {
	cookie := new(http.Cookie)
	cookie.Name = name
	cookie.Value = token
	cookie.Expires = expiration
	cookie.Path = "/"
	cookie.HttpOnly = true

	c.SetCookie(cookie)
}

func (authService *AuthService) setUserCookie(c echo.Context, user *models.UserInfoResponse, expiration time.Time) {
	cookie := new(http.Cookie)
	cookie.Name = userTokenCookieName
	cookie.Value = strconv.Itoa(int(user.ID))
	cookie.Expires = expiration
	cookie.Path = "/"
	c.SetCookie(cookie)
}
