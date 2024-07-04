package controllers

import (
	"github.com/ShukinDmitriy/gophermart/internal/auth"
	"github.com/ShukinDmitriy/gophermart/internal/repositories"
	"github.com/ShukinDmitriy/gophermart/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/theplant/luhn"
	"io"
	"net/http"
	"strconv"
)

type OrderController struct {
	authService     auth.AuthServiceInterface
	orderRepository repositories.OrderRepositoryInterface
	accrualService  services.AccrualServiceInterface
}

func NewOrderController(
	authService auth.AuthServiceInterface,
	orderRepository repositories.OrderRepositoryInterface,
	accrualService services.AccrualServiceInterface,
) *OrderController {
	return &OrderController{
		authService:     authService,
		orderRepository: orderRepository,
		accrualService:  accrualService,
	}
}

// CreateOrder Создать заказ
func (controller *OrderController) CreateOrder() echo.HandlerFunc {
	return func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		defer c.Request().Body.Close()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		orderNumberString := string(body)
		orderNumber, err := strconv.Atoi(orderNumberString)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		if !luhn.Valid(orderNumber) {
			return c.JSON(http.StatusUnprocessableEntity, nil)
		}

		existOrder, err := controller.orderRepository.FindByNumber(orderNumberString)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		currentUserID := controller.authService.GetUserID(c)
		if existOrder != nil {
			if existOrder.UserID == currentUserID {
				return c.JSON(http.StatusOK, nil)
			}

			return c.JSON(http.StatusConflict, nil)
		}

		order, err := controller.orderRepository.Create(orderNumberString, currentUserID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		controller.accrualService.SendOrderToQueue(*order)

		return c.JSON(http.StatusAccepted, nil)
	}
}

// GetOrders Получение списка загруженных номеров заказов
func (controller *OrderController) GetOrders() echo.HandlerFunc {
	return func(c echo.Context) error {
		currentUserID := controller.authService.GetUserID(c)

		orders, err := controller.orderRepository.GetOrdersByUserID(currentUserID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		if len(orders) == 0 {
			return c.JSON(http.StatusNoContent, nil)
		}

		return c.JSON(http.StatusOK, orders)
	}
}
