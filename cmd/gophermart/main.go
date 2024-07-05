package main

import (
	"context"
	"errors"
	"flag"
	"github.com/ShukinDmitriy/gophermart/internal/auth"
	"github.com/ShukinDmitriy/gophermart/internal/config"
	"github.com/ShukinDmitriy/gophermart/internal/controllers"
	"github.com/ShukinDmitriy/gophermart/internal/repositories"
	"github.com/ShukinDmitriy/gophermart/internal/services"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http"
	"os"
	"time"
)

func main() {
	fx.New(
		fx.Provide(
			NewHTTPServer,
			NewConfig,
			NewDB,
			func(DB *gorm.DB) *repositories.AccountRepository {
				return repositories.NewAccountRepository(DB)
			},
			func(DB *gorm.DB) *repositories.OperationRepository {
				return repositories.NewOperationRepository(DB)
			},
			func(DB *gorm.DB) *repositories.OrderRepository {
				return repositories.NewOrderRepository(DB)
			},
			func(DB *gorm.DB) *repositories.UserRepository {
				return repositories.NewUserRepository(DB)
			},
			func() *http.Client {
				return &http.Client{
					Timeout: 10 * time.Second,
				}
			},
			func(
				conf *config.Config,
				accountRepository *repositories.AccountRepository,
				operationRepository *repositories.OperationRepository,
				orderRepository *repositories.OrderRepository,
				httpClient *http.Client,
			) *services.AccrualService {
				return services.NewAccrualService(
					conf.AccrualSystemAddress,
					accountRepository,
					operationRepository,
					orderRepository,
					httpClient,
				)
			},
			func(userRepository *repositories.UserRepository) *auth.AuthUser {
				return auth.NewAuthUser(userRepository)
			},
			func(authUser *auth.AuthUser) *auth.AuthService {
				return auth.NewAuthService(*authUser)
			},
			func(
				authService *auth.AuthService,
				accountRepository *repositories.AccountRepository,
				operationRepository *repositories.OperationRepository,
			) *controllers.BalanceController {
				return controllers.NewBalanceController(
					authService,
					accountRepository,
					operationRepository,
				)
			},
			func(
				authService *auth.AuthService,
				accountRepository *repositories.AccountRepository,
				operationRepository *repositories.OperationRepository,
				orderRepository *repositories.OrderRepository,
			) *controllers.OperationController {
				return controllers.NewOperationController(
					authService,
					accountRepository,
					operationRepository,
					orderRepository,
				)
			},
			func(
				authService *auth.AuthService,
				orderRepository *repositories.OrderRepository,
				accrualService *services.AccrualService,
			) *controllers.OrderController {
				return controllers.NewOrderController(
					authService,
					orderRepository,
					accrualService,
				)
			},
			func(
				authService *auth.AuthService,
				userRepository *repositories.UserRepository,
			) *controllers.UserController {
				return controllers.NewUserController(
					authService,
					userRepository,
				)
			},
		),
		fx.Invoke(func(*echo.Echo) {}),
		fx.Invoke(runMigrate),
		fx.Invoke(func(accrualService *services.AccrualService, e *echo.Echo) {
			go accrualService.ProcessOrders(e)
			go accrualService.ProcessFailedOrders(e)
		}),
	).Run()
}

func NewDB(conf *config.Config) *gorm.DB {
	postgresqlURL := conf.DatabaseURI

	if postgresqlURL == "" {
		log.Fatal("no DATABASE_URI in .env")
		return nil
	}

	db, err := gorm.Open(postgres.Open(postgresqlURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return db
}

func NewConfig() *config.Config {
	conf := config.NewConfig()

	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	flag.StringVar(&conf.RunAddress, "a", "localhost:8080", "Run address")
	flag.StringVar(&conf.DatabaseURI, "d", "", "Database dsn")
	flag.StringVar(&conf.AccrualSystemAddress, "r", "http://localhost:8082", "Accrual system address")
	flag.StringVar(&conf.JwtSecretKey, "s", "", "JWT secret key")

	flag.Parse()

	runAddress, exists := os.LookupEnv("RUN_ADDRESS")
	if exists {
		conf.RunAddress = runAddress
	}

	databaseURI, exists := os.LookupEnv("DATABASE_URI")
	if exists {
		conf.DatabaseURI = databaseURI
	}

	accrualSystemAddress, exists := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if exists {
		conf.AccrualSystemAddress = accrualSystemAddress
	}

	jwtSecretKey, exists := os.LookupEnv("JWT_SECRET_KEY")
	if exists {
		conf.JwtSecretKey = jwtSecretKey
	}

	return conf
}

func NewHTTPServer(
	lc fx.Lifecycle,
	conf *config.Config,
	authService *auth.AuthService,
	balanceController *controllers.BalanceController,
	operationController *controllers.OperationController,
	orderController *controllers.OrderController,
	userController *controllers.UserController,
) *echo.Echo {
	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	// middleware
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	// decompress
	e.Use(middleware.Decompress())

	jwtMiddleware := echojwt.WithConfig(echojwt.Config{
		BeforeFunc: authService.BeforeFunc,
		NewClaimsFunc: func(_ echo.Context) jwt.Claims {
			return &auth.Claims{}
		},
		SigningKey:    []byte(auth.GetJWTSecret()),
		SigningMethod: jwt.SigningMethodHS256.Alg(),
		TokenLookup:   "cookie:access-token", // "<source>:<name>"
		ErrorHandler:  authService.JWTErrorChecker,
	})

	// routes
	// POST /api/user/login — аутентификация пользователя;
	// POST /api/user/orders — загрузка пользователем номера заказа для расчёта;
	// GET /api/user/orders — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
	// GET /api/user/balance — получение текущего баланса счёта баллов лояльности пользователя;
	// POST /api/user/balance/withdraw — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
	// GET /api/user/withdrawals — получение информации о выводе средств с накопительного счёта пользователем.

	e.POST("/api/user/register", userController.UserRegister())
	e.POST("/api/user/login", userController.UserLogin())
	e.POST("/api/user/orders", orderController.CreateOrder(), jwtMiddleware)
	e.GET("/api/user/orders", orderController.GetOrders(), jwtMiddleware)
	e.GET("/api/user/balance", balanceController.GetBalance(), jwtMiddleware)
	e.POST("/api/user/balance/withdraw", operationController.CreateWithdraw(), jwtMiddleware)
	e.GET("/api/user/withdrawals", operationController.GetWithdrawals(), jwtMiddleware)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := e.Start(conf.RunAddress); err != nil && !errors.Is(err, http.ErrServerClosed) {
					e.Logger.Fatal("shutting down the gophermart ", err.Error())
				}

				e.Logger.Info("Running gophermart")
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return e.Shutdown(ctx)
		},
	})

	return e
}
