package repositories

import (
	"github.com/ShukinDmitriy/gophermart/internal/entities"
	"github.com/ShukinDmitriy/gophermart/internal/models"
)

type OrderRepositoryInterface interface {
	Create(number string, userID uint) (*entities.Order, error)
	UpdateOrderByAccrualOrder(accrualOrder *models.AccrualOrderResponse) error
	GetOrdersForProcess() ([]*entities.Order, error)
	FindByNumber(number string) (*entities.Order, error)
	GetOrdersByUserID(userID uint) ([]*models.GetOrdersResponse, error)
}
