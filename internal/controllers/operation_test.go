package controllers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

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

var _ = Describe("Operation", func() {
	var e *echo.Echo
	var c echo.Context
	var rec *httptest.ResponseRecorder
	var authService *auth.AuthServiceInterface
	var accountRepository *repositories.AccountRepositoryInterface
	var operationRepository *repositories.OperationRepositoryInterface
	var orderRepository *repositories.OrderRepositoryInterface
	var controller *controllers.OperationController
	createWithdrawRequest := &models.CreateWithdrawRequest{
		Order: "12345678903",
		Sum:   123.45,
	}
	createWithdrawRequestJSON, _ := json.Marshal(createWithdrawRequest)
	account := &entities.Account{
		Model: gorm.Model{
			ID: 7,
		},
		Sum: 789.58,
	}
	userID := uint(1)
	withdrawals := []models.GetWithdrawalsResponse{
		{},
		{},
		{},
	}

	BeforeEach(func() {
		e = echo.New()
		rec = httptest.NewRecorder()
		authService = new(auth.AuthServiceInterface)
		accountRepository = new(repositories.AccountRepositoryInterface)
		operationRepository = new(repositories.OperationRepositoryInterface)
		orderRepository = new(repositories.OrderRepositoryInterface)
		controller = controllers.NewOperationController(
			authService,
			accountRepository,
			operationRepository,
			orderRepository,
		)
	})

	Describe("CreateWithdraw", func() {
		It("should return the correct answer", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(createWithdrawRequestJSON)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			authService.EXPECT().GetUserID(c).Return(userID)
			accountRepository.EXPECT().FindByUserID(userID, entities.AccountTypeBonus).Return(account, nil)
			orderRepository.EXPECT().FindByNumber(createWithdrawRequest.Order).Return(nil, nil)
			operationRepository.EXPECT().CreateWithdrawn(account.ID, createWithdrawRequest.Order, createWithdrawRequest.Sum).Return(nil)

			// Act
			err := controller.CreateWithdraw()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))
		})

		It("should return an error if the user is not found", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(createWithdrawRequestJSON)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			authService.EXPECT().GetUserID(c).Return(0)

			// Act
			err := controller.CreateWithdraw()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusUnauthorized))
		})

		It("should return an error if the account is not found", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(createWithdrawRequestJSON)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			authService.EXPECT().GetUserID(c).Return(userID)
			accountRepository.EXPECT().FindByUserID(userID, entities.AccountTypeBonus).Return(nil, nil)

			// Act
			err := controller.CreateWithdraw()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should return an error if there are not enough funds", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(createWithdrawRequestJSON)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			authService.EXPECT().GetUserID(c).Return(userID)
			accountRepository.EXPECT().FindByUserID(userID, entities.AccountTypeBonus).Return(&entities.Account{
				Model: gorm.Model{
					ID: 7,
				},
				Sum: 100,
			}, nil)

			// Act
			err := controller.CreateWithdraw()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusPaymentRequired))
		})

		It("should return an error if the requested order does not belong to the user", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(createWithdrawRequestJSON)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			authService.EXPECT().GetUserID(c).Return(userID)
			accountRepository.EXPECT().FindByUserID(userID, entities.AccountTypeBonus).Return(account, nil)
			orderRepository.EXPECT().FindByNumber(createWithdrawRequest.Order).Return(&entities.Order{
				UserID: 123,
			}, nil)

			// Act
			err := controller.CreateWithdraw()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
		})

		It("should return an error if the write-off could not be created", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(createWithdrawRequestJSON)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			authService.EXPECT().GetUserID(c).Return(userID)
			accountRepository.EXPECT().FindByUserID(userID, entities.AccountTypeBonus).Return(account, nil)
			orderRepository.EXPECT().FindByNumber(createWithdrawRequest.Order).Return(nil, nil)
			operationRepository.EXPECT().CreateWithdrawn(account.ID, createWithdrawRequest.Order, createWithdrawRequest.Sum).Return(errors.New("test error"))

			// Act
			err := controller.CreateWithdraw()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})
	})

	Describe("GetWithdrawals", func() {
		It("should return the correct answer", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			authService.EXPECT().GetUserID(c).Return(userID)
			accountRepository.EXPECT().FindByUserID(userID, entities.AccountTypeBonus).Return(account, nil)
			operationRepository.EXPECT().GetWithdrawalsByAccountID(account.ID).Return(withdrawals, nil)

			// Act
			err := controller.GetWithdrawals()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))

			var resJ []models.GetWithdrawalsResponse
			err = json.Unmarshal(rec.Body.Bytes(), &resJ)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(resJ)).To(Equal(len(withdrawals)))
		})

		It("should return the status if there are no write-offs", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			authService.EXPECT().GetUserID(c).Return(userID)
			accountRepository.EXPECT().FindByUserID(userID, entities.AccountTypeBonus).Return(account, nil)
			operationRepository.EXPECT().GetWithdrawalsByAccountID(account.ID).Return([]models.GetWithdrawalsResponse{}, nil)

			// Act
			err := controller.GetWithdrawals()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusNoContent))
		})

		It("should return an error if the user is not found", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			authService.EXPECT().GetUserID(c).Return(0)

			// Act
			err := controller.GetWithdrawals()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusUnauthorized))
		})

		It("should return an error if the account is not found", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			authService.EXPECT().GetUserID(c).Return(userID)
			accountRepository.EXPECT().FindByUserID(userID, entities.AccountTypeBonus).Return(nil, nil)

			// Act
			err := controller.GetWithdrawals()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should return an error if an error occurred when receiving debits", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			authService.EXPECT().GetUserID(c).Return(userID)
			accountRepository.EXPECT().FindByUserID(userID, entities.AccountTypeBonus).Return(account, nil)
			operationRepository.EXPECT().GetWithdrawalsByAccountID(account.ID).Return(withdrawals, errors.New("test error"))

			// Act
			err := controller.GetWithdrawals()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})
	})
})
