package controllers_test

import (
	"encoding/json"
	"errors"
	"github.com/ShukinDmitriy/gophermart/internal/controllers"
	"github.com/ShukinDmitriy/gophermart/internal/entities"
	"github.com/ShukinDmitriy/gophermart/internal/models"
	"github.com/ShukinDmitriy/gophermart/mocks/internal_/auth"
	"github.com/ShukinDmitriy/gophermart/mocks/internal_/repositories"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"strings"
)

var _ = Describe("User", func() {
	var e *echo.Echo
	var c echo.Context
	var rec *httptest.ResponseRecorder
	var authService *auth.AuthServiceInterface
	var userRepository *repositories.UserRepositoryInterface
	var controller *controllers.UserController
	userRequest := models.UserRegisterRequest{
		Login:    "fxf9kP0pO4w",
		Password: "QT4jmi5LhBCDgusiqObdgvfS",
	}
	userRequestJSON, _ := json.Marshal(userRequest)
	user := &entities.User{
		Model: gorm.Model{
			ID: uint(1),
		},
		Login:    userRequest.Login,
		Password: "$2a$08$hCNT/wAMPJgZEAX4EzoTPO1OLahx8sU/3gQvMwwlyLrUyBysTdIKi",
	}
	userInvalidRequest := models.UserRegisterRequest{
		Login:    "fx1",
		Password: "QT4",
	}
	userInvalidRequestJSON, _ := json.Marshal(userInvalidRequest)
	userRegisterResponse := &models.UserInfoResponse{
		ID:    uint(1),
		Login: "fxf9kP0pO4w",
	}

	BeforeEach(func() {
		e = echo.New()
		rec = httptest.NewRecorder()
		authService = new(auth.AuthServiceInterface)
		userRepository = new(repositories.UserRepositoryInterface)
		controller = controllers.NewUserController(
			authService,
			userRepository,
		)
	})

	Describe("User register", func() {
		It("should return the correct answer", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(userRequestJSON)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			userRepository.EXPECT().FindBy(models.UserSearchFilter{Login: userRequest.Login}).Return(nil, nil)
			userRepository.EXPECT().Create(userRequest).Return(userRegisterResponse, nil)
			authService.EXPECT().GenerateTokensAndSetCookies(c, userRegisterResponse).Return(nil)

			// Act
			err := controller.UserRegister()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))

			resJ := &models.UserInfoResponse{}
			err = json.Unmarshal(rec.Body.Bytes(), resJ)
			Expect(err).NotTo(HaveOccurred())
			Expect(resJ.Login).To(Equal(userRegisterResponse.Login))
			Expect(resJ.ID).To(Equal(userRegisterResponse.ID))
		})

		It("should return an error if the cookie could not be created", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(userRequestJSON)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			userRepository.EXPECT().FindBy(models.UserSearchFilter{Login: userRequest.Login}).Return(nil, nil)
			userRepository.EXPECT().Create(userRequest).Return(userRegisterResponse, nil)
			authService.EXPECT().GenerateTokensAndSetCookies(c, userRegisterResponse).Return(errors.New("test error"))

			// Act
			err := controller.UserRegister()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should return an error if the user could not be created", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(userRequestJSON)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			userRepository.EXPECT().FindBy(models.UserSearchFilter{Login: userRequest.Login}).Return(nil, nil)
			userRepository.EXPECT().Create(userRequest).Return(userRegisterResponse, errors.New("test error"))

			// Act
			err := controller.UserRegister()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should return an error if the user could not be created", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(userRequestJSON)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			userRepository.EXPECT().FindBy(models.UserSearchFilter{Login: userRequest.Login}).Return(nil, nil)
			userRepository.EXPECT().Create(userRequest).Return(userRegisterResponse, errors.New("test error"))

			// Act
			err := controller.UserRegister()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should return an error if the login is already taken", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(userRequestJSON)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			userRepository.EXPECT().FindBy(models.UserSearchFilter{Login: userRequest.Login}).Return(&entities.User{Login: userRequest.Login}, nil)

			// Act
			err := controller.UserRegister()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusConflict))
		})

		It("should return an error if the login could not be verified", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(userRequestJSON)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			userRepository.EXPECT().FindBy(models.UserSearchFilter{Login: userRequest.Login}).Return(nil, errors.New("test error"))

			// Act
			err := controller.UserRegister()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should return a validation error", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(userInvalidRequestJSON)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)

			// Act
			err := controller.UserRegister()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusBadRequest))

			resJ := &struct {
				Login    map[string]bool `json:"login"`
				Password map[string]bool `json:"password"`
			}{}
			err = json.Unmarshal(rec.Body.Bytes(), resJ)
			Expect(err).NotTo(HaveOccurred())
			Expect(resJ.Login["min"]).To(BeTrue())
			Expect(resJ.Password["min"]).To(BeTrue())
		})
	})

	Describe("User login", func() {
		It("should return the correct answer", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(userRequestJSON)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			userRepository.EXPECT().FindBy(models.UserSearchFilter{Login: userRequest.Login}).Return(user, nil)
			authService.EXPECT().GenerateTokensAndSetCookies(c, &models.UserInfoResponse{
				ID:         user.ID,
				LastName:   user.LastName,
				FirstName:  user.FirstName,
				MiddleName: user.MiddleName,
				Login:      user.Login,
				Email:      user.Email,
			}).Return(nil)

			// Act
			err := controller.UserLogin()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))

			resJ := &models.UserInfoResponse{}
			err = json.Unmarshal(rec.Body.Bytes(), resJ)
			Expect(err).NotTo(HaveOccurred())
			Expect(resJ.Login).To(Equal(userRegisterResponse.Login))
		})

		It("should return an error if the cookie could not be created", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(userRequestJSON)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			userRepository.EXPECT().FindBy(models.UserSearchFilter{Login: userRequest.Login}).Return(user, nil)
			authService.EXPECT().GenerateTokensAndSetCookies(c, &models.UserInfoResponse{
				ID:         user.ID,
				LastName:   user.LastName,
				FirstName:  user.FirstName,
				MiddleName: user.MiddleName,
				Login:      user.Login,
				Email:      user.Email,
			}).Return(errors.New("test error"))

			// Act
			err := controller.UserLogin()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should return an error if the passwords do not match", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(userRequestJSON)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			userRepository.EXPECT().FindBy(models.UserSearchFilter{Login: userRequest.Login}).Return(&entities.User{
				Login:    userRequest.Login,
				Password: "",
			}, nil)

			// Act
			err := controller.UserLogin()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusUnauthorized))
		})

		It("should return an error if the user is not found", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(userRequestJSON)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			userRepository.EXPECT().FindBy(models.UserSearchFilter{Login: userRequest.Login}).Return(nil, nil)

			// Act
			err := controller.UserLogin()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusUnauthorized))
		})

		It("should return an error if it was not possible to get the user", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(userRequestJSON)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			userRepository.EXPECT().FindBy(models.UserSearchFilter{Login: userRequest.Login}).Return(user, errors.New("test error"))

			// Act
			err := controller.UserLogin()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should return a validation error", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(userInvalidRequestJSON)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)

			// Act
			err := controller.UserLogin()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusBadRequest))

			resJ := &struct {
				Login    map[string]bool `json:"login"`
				Password map[string]bool `json:"password"`
			}{}
			err = json.Unmarshal(rec.Body.Bytes(), resJ)
			Expect(err).NotTo(HaveOccurred())
			Expect(resJ.Login["min"]).To(BeTrue())
			Expect(resJ.Password["min"]).To(BeTrue())
		})
	})
})
