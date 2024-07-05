package controllers

import (
	"net/http"

	"github.com/ShukinDmitriy/gophermart/internal/auth"
	"github.com/ShukinDmitriy/gophermart/internal/entities"
	"github.com/ShukinDmitriy/gophermart/internal/models"
	"github.com/ShukinDmitriy/gophermart/internal/repositories"
	"github.com/labstack/echo/v4"
)

type OperationController struct {
	authService         auth.AuthServiceInterface
	accountRepository   repositories.AccountRepositoryInterface
	operationRepository repositories.OperationRepositoryInterface
	orderRepository     repositories.OrderRepositoryInterface
}

func NewOperationController(
	authService auth.AuthServiceInterface,
	accountRepository repositories.AccountRepositoryInterface,
	operationRepository repositories.OperationRepositoryInterface,
	orderRepository repositories.OrderRepositoryInterface,
) *OperationController {
	return &OperationController{
		authService:         authService,
		accountRepository:   accountRepository,
		operationRepository: operationRepository,
		orderRepository:     orderRepository,
	}
}

// CreateWithdraw Cписание баллов с накопительного счёта в счёт оплаты нового заказа
func (controller *OperationController) CreateWithdraw() echo.HandlerFunc {
	return func(c echo.Context) error {
		var createWithdrawRequest models.CreateWithdrawRequest
		err := c.Bind(&createWithdrawRequest)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, nil)
		}

		currentUserID := controller.authService.GetUserID(c)
		if currentUserID == 0 {
			c.Logger().Error("Unauthorized user create withdraw")
			return c.JSON(http.StatusUnauthorized, nil)
		}

		bonusAccount, err := controller.accountRepository.FindByUserID(currentUserID, entities.AccountTypeBonus)
		if err != nil || bonusAccount == nil {
			c.Logger().Error("Can't find bonus account", err)
			return c.JSON(http.StatusInternalServerError, nil)
		}
		if bonusAccount.Sum < createWithdrawRequest.Sum {
			c.Logger().Error("Balance error")
			return c.JSON(http.StatusPaymentRequired, nil)
		}

		order, err := controller.orderRepository.FindByNumber(createWithdrawRequest.Order)
		if err != nil || (order != nil && order.UserID != currentUserID) {
			c.Logger().Error("Can't find order", createWithdrawRequest.Order, currentUserID, err)
			return c.JSON(http.StatusUnprocessableEntity, nil)
		}

		err = controller.operationRepository.CreateWithdrawn(bonusAccount.ID, createWithdrawRequest.Order, createWithdrawRequest.Sum)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		return c.JSON(http.StatusOK, nil)
	}
}

// GetWithdrawals Получение информации о выводе средств
func (controller *OperationController) GetWithdrawals() echo.HandlerFunc {
	return func(c echo.Context) error {
		currentUserID := controller.authService.GetUserID(c)
		if currentUserID == 0 {
			c.Logger().Error("Unauthorized user create withdraw")
			return c.JSON(http.StatusUnauthorized, nil)
		}

		bonusAccount, err := controller.accountRepository.FindByUserID(currentUserID, entities.AccountTypeBonus)
		if err != nil || bonusAccount == nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		operations, err := controller.operationRepository.GetWithdrawalsByAccountID(bonusAccount.ID)
		if err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, nil)
		}

		if len(operations) == 0 {
			return c.JSON(http.StatusNoContent, nil)
		}

		return c.JSON(http.StatusOK, operations)
	}
}
