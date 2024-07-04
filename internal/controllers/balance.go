package controllers

import (
	"github.com/ShukinDmitriy/gophermart/internal/auth"
	"github.com/ShukinDmitriy/gophermart/internal/entities"
	"github.com/ShukinDmitriy/gophermart/internal/models"
	"github.com/ShukinDmitriy/gophermart/internal/repositories"
	"github.com/labstack/echo/v4"
	"net/http"
)

type BalanceController struct {
	authService         auth.AuthServiceInterface
	accountRepository   repositories.AccountRepositoryInterface
	operationRepository repositories.OperationRepositoryInterface
}

func NewBalanceController(
	authService auth.AuthServiceInterface,
	accountRepository repositories.AccountRepositoryInterface,
	operationRepository repositories.OperationRepositoryInterface,
) *BalanceController {
	return &BalanceController{
		authService:         authService,
		accountRepository:   accountRepository,
		operationRepository: operationRepository,
	}
}

// GetBalance Получение баланса пользователя
func (controller *BalanceController) GetBalance() echo.HandlerFunc {
	return func(c echo.Context) error {
		resp := &models.GetBalanceResponse{}
		currentUserID := controller.authService.GetUserID(c)

		bonusAccount, err := controller.accountRepository.FindByUserID(currentUserID, entities.AccountTypeBonus)
		if err != nil || bonusAccount == nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		resp.Current = bonusAccount.Sum

		withdrawn, err := controller.operationRepository.GetWithdrawnByAccountID(bonusAccount.ID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		if withdrawn != 0 {
			resp.Withdrawn = withdrawn
		}

		return c.JSON(http.StatusOK, resp)
	}
}
