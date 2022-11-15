package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"gophermart/internal/app/configs"
	"gophermart/internal/app/middlewares"
	"gophermart/internal/app/repositories"
	"gophermart/internal/app/services"
)

func InitRouter(db *sqlx.DB, cfg *configs.Config, orderService OrderService) *gin.Engine {
	r := gin.Default()
	userRepository := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepository)
	tokenService := services.NewAuthJWTTokenService(cfg.AuthSecretKey)
	registrationService := services.NewRegistrationService(userService, tokenService)
	authService := services.NewAuthService(userService, tokenService)

	apiGroup := r.Group("/api/user")
	registrationHandler := NewRegistrationHandler(registrationService)
	apiGroup.POST("/register", registrationHandler.HandleRegistration)

	loginHandler := NewLoginHandler(authService)
	apiGroup.POST("/login", loginHandler.HandleLogin)

	needAuthURLsGroup := apiGroup.Group("")
	needAuthURLsGroup.Use(middlewares.TokenAuthMiddleware(userService, authService))

	orderNumberValidator := services.NewOrderNumberValidator()
	orderHandler := NewOrderHandler(authService, orderService, orderNumberValidator)
	needAuthURLsGroup.POST("/orders", orderHandler.HandleCreateOrder)
	needAuthURLsGroup.GET("/orders", orderHandler.HandleListOrders)

	balanceRepository := repositories.NewBalanceRepository(db)
	balanceService := services.NewUserBalanceService(balanceRepository)
	balanceHandler := NewUserBalanceHandler(orderNumberValidator, authService, balanceService)
	needAuthURLsGroup.POST("/balance/withdraw", balanceHandler.HandleWithdrawBalance)
	needAuthURLsGroup.GET("/withdrawals", balanceHandler.HandleListBalanceWithdrawals)
	needAuthURLsGroup.GET("/balance", balanceHandler.HandleGetUserBalance)

	return r
}
