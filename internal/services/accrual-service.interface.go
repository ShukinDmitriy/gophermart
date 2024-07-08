package services

import (
	"github.com/ShukinDmitriy/gophermart/internal/entities"
	"github.com/labstack/echo/v4"
)

type AccrualServiceInterface interface {
	SendOrderToQueue(order entities.Order)
	ProcessOrders(e *echo.Echo)
	ProcessFailedOrders(e *echo.Echo)
}
