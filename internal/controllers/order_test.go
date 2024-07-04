package controllers_test

import (
	"encoding/json"
	"errors"
	"github.com/ShukinDmitriy/gophermart/internal/models"

	//"encoding/json"
	//"errors"
	"github.com/ShukinDmitriy/gophermart/internal/controllers"
	"github.com/ShukinDmitriy/gophermart/internal/entities"
	//"github.com/ShukinDmitriy/gophermart/internal/models"
	"github.com/ShukinDmitriy/gophermart/mocks/internal_/auth"
	"github.com/ShukinDmitriy/gophermart/mocks/internal_/repositories"
	"github.com/ShukinDmitriy/gophermart/mocks/internal_/services"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	//"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	//"strconv"
	"strings"
)

var _ = Describe("Order", func() {
	var e *echo.Echo
	var c echo.Context
	var rec *httptest.ResponseRecorder
	var authService *auth.AuthServiceInterface
	var orderRepository *repositories.OrderRepositoryInterface
	var accrualService *services.AccrualServiceInterface
	var controller *controllers.OrderController
	createOrderRequestString := "12345678903"
	order := &entities.Order{}
	getOrdersResponse := []*models.GetOrdersResponse{
		{},
		{},
	}
	//account := &entities.Account{
	//	Model: gorm.Model{
	//		ID: 7,
	//	},
	//	Sum: response.Current,
	//}
	userID := uint(1)

	BeforeEach(func() {
		e = echo.New()
		rec = httptest.NewRecorder()
		authService = new(auth.AuthServiceInterface)
		orderRepository = new(repositories.OrderRepositoryInterface)
		accrualService = new(services.AccrualServiceInterface)
		controller = controllers.NewOrderController(
			authService,
			orderRepository,
			accrualService,
		)
	})

	Describe("Create Order", func() {
		It("should return the correct answer", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(createOrderRequestString))
			req.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)
			c = e.NewContext(req, rec)
			orderRepository.EXPECT().FindByNumber(createOrderRequestString).Return(nil, nil)
			authService.EXPECT().GetUserID(c).Return(userID)
			orderRepository.EXPECT().Create(createOrderRequestString, userID).Return(order, nil)
			accrualService.EXPECT().SendOrderToQueue(*order).Return()

			// Act
			err := controller.CreateOrder()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusAccepted))
		})

		It("should return an error if the order could not be created", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(createOrderRequestString))
			req.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)
			c = e.NewContext(req, rec)
			orderRepository.EXPECT().FindByNumber(createOrderRequestString).Return(nil, nil)
			authService.EXPECT().GetUserID(c).Return(userID)
			orderRepository.EXPECT().Create(createOrderRequestString, userID).Return(order, errors.New("test error"))

			// Act
			err := controller.CreateOrder()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should return status conflict if someone else has a created order with the same number", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(createOrderRequestString))
			req.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)
			c = e.NewContext(req, rec)
			orderRepository.EXPECT().FindByNumber(createOrderRequestString).Return(&entities.Order{
				Number: createOrderRequestString,
				UserID: userID + 1,
			}, nil)
			authService.EXPECT().GetUserID(c).Return(userID)

			// Act
			err := controller.CreateOrder()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusConflict))
		})

		It("should return success if the user has a created order with the same number", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(createOrderRequestString))
			req.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)
			c = e.NewContext(req, rec)
			orderRepository.EXPECT().FindByNumber(createOrderRequestString).Return(&entities.Order{
				Number: createOrderRequestString,
				UserID: userID,
			}, nil)
			authService.EXPECT().GetUserID(c).Return(userID)

			// Act
			err := controller.CreateOrder()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))
		})

		It("should return an error if the order could not be received", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(createOrderRequestString))
			req.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)
			c = e.NewContext(req, rec)
			orderRepository.EXPECT().FindByNumber(createOrderRequestString).Return(nil, errors.New("test error"))

			// Act
			err := controller.CreateOrder()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should return an error if the order number does not match the Luhn algorithm", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("12345678904"))
			req.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)
			c = e.NewContext(req, rec)

			// Act
			err := controller.CreateOrder()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
		})

		It("should return an error if the order number consists of letters", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test"))
			req.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)
			c = e.NewContext(req, rec)

			// Act
			err := controller.CreateOrder()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})
	})

	Describe("Get Orders", func() {
		It("should return the correct answer", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			authService.EXPECT().GetUserID(c).Return(userID)
			orderRepository.EXPECT().GetOrdersByUserID(userID).Return(getOrdersResponse, nil)

			// Act
			err := controller.GetOrders()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))

			var resJ []*models.GetOrdersResponse
			err = json.Unmarshal(rec.Body.Bytes(), &resJ)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(resJ)).To(Equal(len(getOrdersResponse)))
		})

		It("should return status no content if no orders found", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			authService.EXPECT().GetUserID(c).Return(userID)
			orderRepository.EXPECT().GetOrdersByUserID(userID).Return([]*models.GetOrdersResponse{}, nil)

			// Act
			err := controller.GetOrders()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusNoContent))
		})

		It("should return an error if orders could not be received", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			authService.EXPECT().GetUserID(c).Return(userID)
			orderRepository.EXPECT().GetOrdersByUserID(userID).Return(getOrdersResponse, errors.New("test error"))

			// Act
			err := controller.GetOrders()(c)

			// Assertions
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})
	})
})
