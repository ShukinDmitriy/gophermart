package services_test

import (
	"encoding/json"
	"fmt"
	"github.com/ShukinDmitriy/gophermart/internal/entities"
	"github.com/ShukinDmitriy/gophermart/internal/models"
	"github.com/ShukinDmitriy/gophermart/internal/services"
	"github.com/ShukinDmitriy/gophermart/mocks/internal_/repositories"
	"github.com/jfrog/go-mockhttp"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	"gorm.io/gorm"
	"net/http"
	"time"
)

var _ = Describe("AccrualService", func() {
	var e *echo.Echo
	var accountRepository *repositories.AccountRepositoryInterface
	var operationRepository *repositories.OperationRepositoryInterface
	var orderRepository *repositories.OrderRepositoryInterface
	var httpClient *http.Client
	var service *services.AccrualService

	userID := uint(1)
	accountID := uint(2)
	processingOrderNumber := "24619735244"
	noContentOrderNumber := "62794305672"
	processedOrderNumber := "61508349208"
	accrualProcessingResponse := models.AccrualOrderResponse{
		Order:  processingOrderNumber,
		Status: entities.OrderStatusProcessing,
	}
	accrualProcessingResponseJSON, _ := json.Marshal(accrualProcessingResponse)
	accrualNoContentResponse := models.AccrualOrderResponse{
		Order:  noContentOrderNumber,
		Status: entities.OrderStatusProcessing,
	}
	accrualProcessedResponse := models.AccrualOrderResponse{
		Order:   processedOrderNumber,
		Status:  entities.OrderStatusProcessed,
		Accrual: 123.45,
	}
	accrualProcessedResponseJSON, _ := json.Marshal(accrualProcessedResponse)

	BeforeEach(func() {
		e = echo.New()
		accountRepository = new(repositories.AccountRepositoryInterface)
		operationRepository = new(repositories.OperationRepositoryInterface)
		orderRepository = new(repositories.OrderRepositoryInterface)
		client := mockhttp.NewClient(
			mockhttp.NewClientEndpoint().
				When(mockhttp.Request().GET(fmt.Sprintf("/api/orders/%s", processingOrderNumber))).
				Respond(mockhttp.Response().StatusCode(http.StatusOK).Body(accrualProcessingResponseJSON)),
			mockhttp.NewClientEndpoint().
				When(mockhttp.Request().GET(fmt.Sprintf("/api/orders/%s", processedOrderNumber))).
				Respond(mockhttp.Response().StatusCode(http.StatusOK).Body(accrualProcessedResponseJSON)),
			mockhttp.NewClientEndpoint().
				When(mockhttp.Request().GET(fmt.Sprintf("/api/orders/%s", noContentOrderNumber))).
				Respond(mockhttp.Response().StatusCode(http.StatusNoContent)),
		)

		httpClient = client.HttpClient()
		service = services.NewAccrualService(
			"",
			accountRepository,
			operationRepository,
			orderRepository,
			httpClient,
		)
	})

	Describe("Process orders", func() {
		It("must update the order if the response is successful", func() {
			finishedChan := make(chan bool)
			timeout := time.After(time.Second * 1)

			// Arrange
			orderRepository.EXPECT().UpdateOrderByAccrualOrder(&accrualProcessingResponse).RunAndReturn(func(response *models.AccrualOrderResponse) error {
				finishedChan <- true

				return nil
			})

			// Act
			go service.ProcessOrders(e)
			service.SendOrderToQueue(entities.Order{
				Number: processingOrderNumber,
				UserID: userID,
			})
			select {
			case <-finishedChan:
			case <-timeout:
			}

			// Assertions
			orderRepository.MethodCalled("UpdateOrderByAccrualOrder", &accrualProcessingResponse)
		})

		It("must update the order if the response is processed", func() {
			finishedChan := make(chan bool)
			timeout := time.After(time.Second * 1)

			// Arrange
			orderRepository.EXPECT().UpdateOrderByAccrualOrder(&accrualProcessedResponse).Return(nil)
			accountRepository.EXPECT().FindByUserID(userID, entities.AccountTypeBonus).Return(&entities.Account{
				Model: gorm.Model{ID: accountID},
			}, nil)
			operationRepository.EXPECT().CreateAccrual(accountID, accrualProcessedResponse.Order, accrualProcessedResponse.Accrual).RunAndReturn(func(accountID uint, orderNumber string, sum float32) error {
				finishedChan <- true

				return nil
			})

			// Act
			go service.ProcessOrders(e)
			service.SendOrderToQueue(entities.Order{
				Number: processedOrderNumber,
				UserID: userID,
			})
			select {
			case <-finishedChan:
			case <-timeout:
			}

			// Assertions
			orderRepository.MethodCalled("UpdateOrderByAccrualOrder", &accrualProcessedResponse)
		})

		It("must update the order if the response is no content", func() {
			finishedChan := make(chan bool)
			timeout := time.After(time.Second * 1)

			// Arrange
			orderRepository.EXPECT().UpdateOrderByAccrualOrder(&models.AccrualOrderResponse{
				Order:  noContentOrderNumber,
				Status: entities.OrderStatusProcessing,
			}).RunAndReturn(func(response *models.AccrualOrderResponse) error {
				finishedChan <- true

				return nil
			})

			// Act
			go service.ProcessOrders(e)
			service.SendOrderToQueue(entities.Order{
				Number: noContentOrderNumber,
				UserID: userID,
			})
			select {
			case <-finishedChan:
			case <-timeout:
			}

			// Assertions
			orderRepository.MethodCalled("UpdateOrderByAccrualOrder", &accrualNoContentResponse)
		})
	})
})
