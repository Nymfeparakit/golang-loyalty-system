package handlers

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"gophermart/internal/app/domain"
	"gophermart/internal/app/repositories"
	"gophermart/internal/app/services"
	"net/http"
)

type RegistrationService interface {
	RegisterUser(ctx context.Context, user domain.UserDTO) error
}

type RegistrationHandler struct {
	registrationService RegistrationService
}

func NewRegistrationHandler(authService RegistrationService) *RegistrationHandler {
	return &RegistrationHandler{registrationService: authService}
}

func (h *RegistrationHandler) HandleRegistration(c *gin.Context) {
	var input registrationInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": err.Error()})
		return
	}

	userDTO := domain.UserDTO{
		Username: input.Username,
		Password: input.Password,
	}
	err := h.registrationService.RegisterUser(c.Request.Context(), userDTO)
	if errors.Is(err, repositories.ErrUserAlreadyExists) {
		c.JSON(http.StatusConflict, gin.H{"errors": "User with given login already exists"})
		return
	}

	c.Status(http.StatusOK)
}

type AuthService interface {
	AuthenticateUser(ctx context.Context, username string, password string) (*domain.TokenData, error)
	GetUserFromContext(ctx context.Context) (*domain.UserDTO, bool)
}

type LoginHandler struct {
	authService AuthService
}

func NewLoginHandler(authService AuthService) *LoginHandler {
	return &LoginHandler{authService: authService}
}

func (h *LoginHandler) HandleLogin(c *gin.Context) {
	var input loginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": err.Error()})
		return
	}

	tokenData, err := h.authService.AuthenticateUser(c.Request.Context(), input.Username, input.Password)
	if errors.Is(err, services.ErrInvalidCredentials) {
		c.JSON(http.StatusUnauthorized, gin.H{"errors": "Invalid login or password"})
		return
	}

	c.JSON(http.StatusOK, tokenData)
}
