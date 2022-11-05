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
	registrationHandler := NewRegistrationHandler(registrationService)
	r.POST("/register", registrationHandler.HandleRegistration)

	authService := services.NewAuthService(userService)
	loginHandler := NewLoginHandler(authService)
	r.POST("/login", loginHandler.HandleLogin)

	needAuthURLsGroup := r.Group("")
	needAuthURLsGroup.Use(middlewares.TokenAuthMiddleware(userService, authService))

	ordersCh := make(chan string)
	orderRepository := repositories.NewOrderRepository(db)
	orderService := services.NewOrderService(orderRepository, ordersCh)
	orderNumberValidator := services.NewOrderNumberValidator()
	orderHandler := NewOrderHandler(authService, orderService, orderNumberValidator)
	needAuthURLsGroup.POST("/user/orders", orderHandler.HandleCreateOrder)
	needAuthURLsGroup.GET("/user/orders", orderHandler.HandleListOrders)

	workers.InitOrdersWorkers(ordersCh, config, userService)

	return r
}
