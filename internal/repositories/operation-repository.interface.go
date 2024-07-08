package repositories

import "github.com/ShukinDmitriy/gophermart/internal/models"

type OperationRepositoryInterface interface {
	CreateAccrual(accountID uint, orderNumber string, sum float32) error
	CreateWithdrawn(accountID uint, orderNumber string, sum float32) error
	GetWithdrawnByAccountID(accountID uint) (float32, error)
	GetWithdrawalsByAccountID(accountID uint) ([]models.GetWithdrawalsResponse, error)
}
