package repositories

import (
	"github.com/ShukinDmitriy/gophermart/internal/entities"
	"github.com/ShukinDmitriy/gophermart/internal/models"
)

type UserRepositoryInterface interface {
	Create(userRegister models.UserRegisterRequest) (*models.UserInfoResponse, error)
	Find(id uint) (*models.UserInfoResponse, error)
	FindBy(filter models.UserSearchFilter) (*entities.User, error)
	GeneratePasswordHash(password string) ([]byte, error)
}
