package controllers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/ShukinDmitriy/gophermart/internal/controllers"
	"github.com/ShukinDmitriy/gophermart/internal/entities"
	"github.com/ShukinDmitriy/gophermart/internal/models"
	"github.com/ShukinDmitriy/gophermart/mocks/internal_/auth"
	"github.com/ShukinDmitriy/gophermart/mocks/internal_/repositories"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = Describe("Balance", func() {
	var e *echo.Echo
	var c echo.Context
	var rec *httptest.ResponseRecorder
	var authService *auth.AuthServiceInterface
	var accountRepository *repositories.AccountRepositoryInterface
	var operationRepository *repositories.OperationRepositoryInterface
	var controller *controllers.BalanceController
	response := &models.GetBalanceResponse{
		Current:   789.58,
		Withdrawn: 456.25,
	}
	account := &entities.Account{
		Model: gorm.Model{
			ID: 7,
		},
		Sum: response.Current,
	}
	userID := uint(1)

	BeforeEach(func() {
		e = echo.New()
		rec = httptest.NewRecorder()
		authService = new(auth.AuthServiceInterface)
		accountRepository = new(repositories.AccountRepositoryInterface)
		operationRepository = new(repositories.OperationRepositoryInterface)
		controller = controllers.NewBalanceController(
			authService,
			accountRepository,
			operationRepository,
		)
	})

	Describe("Get balance", func() {
		It("should return the correct answer", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			authService.EXPECT().GetUserID(c).Return(userID)
			accountRepository.EXPECT().FindByUserID(userID, entities.AccountTypeBonus).Return(account, nil)
			operationRepository.EXPECT().GetWithdrawnByAccountID(account.ID).Return(response.Withdrawn, nil)

			// Act
			err := controller.GetBalance()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))

			resJ := &models.GetBalanceResponse{}
			err = json.Unmarshal(rec.Body.Bytes(), resJ)
			Expect(err).NotTo(HaveOccurred())
			Expect(resJ.Current).To(Equal(response.Current))
			Expect(resJ.Withdrawn).To(Equal(response.Withdrawn))
		})

		It("should return an error if it was not possible to receive the withdrawn", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			authService.EXPECT().GetUserID(c).Return(userID)
			accountRepository.EXPECT().FindByUserID(userID, entities.AccountTypeBonus).Return(account, nil)
			operationRepository.EXPECT().GetWithdrawnByAccountID(account.ID).Return(0, errors.New("test error"))

			// Act
			err := controller.GetBalance()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should return an error if the account is not found", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			authService.EXPECT().GetUserID(c).Return(userID)
			accountRepository.EXPECT().FindByUserID(userID, entities.AccountTypeBonus).Return(nil, nil)

			// Act
			err := controller.GetBalance()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})
	})
})
