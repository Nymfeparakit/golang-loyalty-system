package handlers

import (
	"github.com/gin-gonic/gin"
	"gophermart/internal/app/configs"
	"gophermart/internal/app/middlewares"
	"gophermart/internal/app/services"
)

func InitRouter(cfg *configs.Config, orderService *services.OrderService, userService *services.UserService) *gin.Engine {
	r := gin.Default()
	r.Use(middlewares.DecompressingRequestMiddleware())
	r.Use(middlewares.CompressingResponseMiddleware())
	tokenService := services.NewAuthJWTTokenService(cfg.AuthSecretKey)
	registrationService := services.NewRegistrationService(userService, tokenService)
	authService := services.NewAuthService(userService, tokenService)
	registrationHandler := NewRegistrationHandler(registrationService)
	r.POST("/api/user/register", registrationHandler.HandleRegistration)

	loginHandler := NewLoginHandler(authService)
	r.POST("/api/user/login", loginHandler.HandleLogin)

	needAuthURLsGroup := r.Group("")
	needAuthURLsGroup.Use(middlewares.TokenAuthMiddleware(userService, authService))

	orderNumberValidator := services.NewOrderNumberValidator()
	orderHandler := NewOrderHandler(authService, orderService, orderNumberValidator)
	needAuthURLsGroup.POST("/api/user/orders", orderHandler.HandleCreateOrder)
	needAuthURLsGroup.GET("/api/user/orders", orderHandler.HandleListOrders)

	return r
}
