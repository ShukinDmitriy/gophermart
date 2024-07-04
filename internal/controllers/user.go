package controllers

import (
	"github.com/ShukinDmitriy/gophermart/internal/auth"
	"github.com/ShukinDmitriy/gophermart/internal/models"
	"github.com/ShukinDmitriy/gophermart/internal/repositories"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type UserController struct {
	authService    auth.AuthServiceInterface
	userRepository repositories.UserRepositoryInterface
}

func NewUserController(
	authService auth.AuthServiceInterface,
	userRepository repositories.UserRepositoryInterface,
) *UserController {
	return &UserController{
		authService:    authService,
		userRepository: userRepository,
	}
}

func (controller *UserController) UserRegister() echo.HandlerFunc {
	return func(c echo.Context) error {
		var userRegisterRequest models.UserRegisterRequest
		err := c.Bind(&userRegisterRequest)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusBadRequest, nil)
		}

		validate := validator.New(validator.WithRequiredStructEnabled())
		err = validate.Struct(userRegisterRequest)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.ExtractErrors(err))
		}

		existUser, err := controller.userRepository.FindBy(models.UserSearchFilter{Login: userRegisterRequest.Login})
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, "internal gophermart error")
		}
		if existUser != nil {
			c.Logger().Error("login already exist")
			return c.JSON(http.StatusConflict, "login already exist")
		}

		user, err := controller.userRepository.Create(userRegisterRequest)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, "internal gophermart error")
		}

		err = controller.authService.GenerateTokensAndSetCookies(c, user)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, "internal gophermart error")
		}

		return c.JSON(http.StatusOK, user)
	}
}

func (controller *UserController) UserLogin() echo.HandlerFunc {
	return func(c echo.Context) error {
		var userLoginRequest models.UserLoginRequest
		err := c.Bind(&userLoginRequest)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusBadRequest, nil)
		}

		validate := validator.New(validator.WithRequiredStructEnabled())
		err = validate.Struct(userLoginRequest)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.ExtractErrors(err))
		}

		existUser, err := controller.userRepository.FindBy(models.UserSearchFilter{Login: userLoginRequest.Login})
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, "internal gophermart error")
		}
		if existUser == nil {
			c.Logger().Error("user not exist")
			return c.JSON(http.StatusUnauthorized, "user not exist")
		}

		if bcrypt.CompareHashAndPassword([]byte(existUser.Password), []byte(userLoginRequest.Password)) != nil {
			c.Logger().Error("invalid password")
			return c.JSON(http.StatusUnauthorized, "invalid password")
		}

		err = controller.authService.GenerateTokensAndSetCookies(c, &models.UserInfoResponse{
			ID:         existUser.ID,
			LastName:   existUser.LastName,
			FirstName:  existUser.FirstName,
			MiddleName: existUser.MiddleName,
			Login:      existUser.Login,
			Email:      existUser.Email,
		})
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, "internal gophermart error")
		}

		return c.JSON(http.StatusOK, models.MapUserToUserLoginResponse(existUser))
	}
}
