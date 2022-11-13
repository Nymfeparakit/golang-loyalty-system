package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"gophermart/internal/app/configs"
	"gophermart/internal/app/middlewares"
	"gophermart/internal/app/repositories"
	"gophermart/internal/app/services"
	"gophermart/internal/app/workers"
)

func InitRouter(db *sqlx.DB, config *configs.Config) *gin.Engine {
	r := gin.Default()
	userRepository := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepository)
	registrationService := services.NewRegistrationService(userService)
	authService := services.NewAuthService(userService)
	registrationHandler := NewRegistrationHandler(registrationService)
	r.POST("/api/user/register", registrationHandler.HandleRegistration)

	loginHandler := NewLoginHandler(authService)
	r.POST("/api/user/login", loginHandler.HandleLogin)

	needAuthURLsGroup := r.Group("")
	needAuthURLsGroup.Use(middlewares.TokenAuthMiddleware(userService, authService))

	ordersCh := make(chan string)
	orderRepository := repositories.NewOrderRepository(db)
	orderService := services.NewOrderService(orderRepository, ordersCh)
	orderNumberValidator := services.NewOrderNumberValidator()
	orderHandler := NewOrderHandler(authService, orderService, orderNumberValidator)
	needAuthURLsGroup.POST("/api/user/orders", orderHandler.HandleCreateOrder)
	needAuthURLsGroup.GET("/api/user/orders", orderHandler.HandleListOrders)

	requestsWorker := workers.NewRateLimitedReqWorker()
	go requestsWorker.ProcessRequests()
	accrualCalculator := services.NewAccrualCalculationService(config.AccrualSystemAddr, requestsWorker)
	workers.InitOrdersWorkers(ordersCh, userService, orderService, accrualCalculator)

	return r
}
