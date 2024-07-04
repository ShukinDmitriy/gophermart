package application

import (
	"github.com/ShukinDmitriy/gophermart/internal/auth"
	"github.com/ShukinDmitriy/gophermart/internal/config"
	"github.com/ShukinDmitriy/gophermart/internal/controllers"
	"github.com/ShukinDmitriy/gophermart/internal/repositories"
	"github.com/ShukinDmitriy/gophermart/internal/services"
	"gorm.io/gorm"
	"net/http"
	"time"
)

var App *Application

type Application struct {
	Conf *config.Config
	DB   *gorm.DB
	// Auth
	AuthUser    *auth.AuthUser
	AuthService *auth.AuthService
	// Controllers
	BalanceController   *controllers.BalanceController
	OperationController *controllers.OperationController
	OrderController     *controllers.OrderController
	UserController      *controllers.UserController
	// Repositories
	AccountRepository   *repositories.AccountRepository
	OperationRepository *repositories.OperationRepository
	OrderRepository     *repositories.OrderRepository
	UserRepository      *repositories.UserRepository
	// Services
	AccrualService *services.AccrualService
}

func NewApplication(
	conf *config.Config,
	DB *gorm.DB,
	// Auth
	authUser *auth.AuthUser,
	authService *auth.AuthService,
	// Controllers
	balanceController *controllers.BalanceController,
	operationController *controllers.OperationController,
	orderController *controllers.OrderController,
	userController *controllers.UserController,
	// Repositories
	accountRepository *repositories.AccountRepository,
	operationRepository *repositories.OperationRepository,
	orderRepository *repositories.OrderRepository,
	userRepository *repositories.UserRepository,
	// Services
	accrualService *services.AccrualService,
) *Application {
	return &Application{
		Conf:                conf,
		DB:                  DB,
		AuthUser:            authUser,
		AuthService:         authService,
		BalanceController:   balanceController,
		OperationController: operationController,
		OrderController:     orderController,
		UserController:      userController,
		AccountRepository:   accountRepository,
		OperationRepository: operationRepository,
		OrderRepository:     orderRepository,
		UserRepository:      userRepository,
		AccrualService:      accrualService,
	}
}

func AppFactory(DB *gorm.DB, conf *config.Config) {
	accountRepository := repositories.NewAccountRepository(DB)
	operationRepository := repositories.NewOperationRepository(DB)
	orderRepository := repositories.NewOrderRepository(DB)
	userRepository := repositories.NewUserRepository(DB)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	accrualService := services.NewAccrualService(
		conf.AccrualSystemAddress,
		accountRepository,
		operationRepository,
		orderRepository,
		httpClient,
	)

	authUser := auth.NewAuthUser(userRepository)
	authService := auth.NewAuthService(*authUser)

	balanceController := controllers.NewBalanceController(
		authService,
		accountRepository,
		operationRepository,
	)
	operationController := controllers.NewOperationController(
		authService,
		accountRepository,
		operationRepository,
		orderRepository,
	)
	orderController := controllers.NewOrderController(
		authService,
		orderRepository,
		accrualService,
	)
	userController := controllers.NewUserController(
		authService,
		userRepository,
	)

	App = NewApplication(
		conf,
		DB,
		authUser,
		authService,
		balanceController,
		operationController,
		orderController,
		userController,
		accountRepository,
		operationRepository,
		orderRepository,
		userRepository,
		accrualService,
	)
}
