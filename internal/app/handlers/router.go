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
	r.Use(middlewares.DecompressingRequestMiddleware())
	r.Use(middlewares.CompressingResponseMiddleware())
	userRepository := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepository)
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
