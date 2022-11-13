package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gophermart/internal/app/domain"
	"gophermart/internal/app/repositories"
	"gophermart/internal/app/services"
	"net/http"
)

type RegistrationService interface {
	RegisterUser(ctx context.Context, user domain.UserDTO) (*domain.TokenData, error)
}

type AuthService interface {
	AuthenticateUser(ctx context.Context, username string, password string) (*domain.TokenData, error)
	GetUserFromContext(ctx context.Context) (*domain.UserDTO, bool)
}

type RegistrationHandler struct {
	registrationService RegistrationService
}

func NewRegistrationHandler(regService RegistrationService) *RegistrationHandler {
	return &RegistrationHandler{registrationService: regService}
}

func (h *RegistrationHandler) HandleRegistration(c *gin.Context) {
	var input registrationInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": err.Error()})
		return
	}

	userDTO := domain.UserDTO{
		Login:    input.Login,
		Password: input.Password,
	}
	tokenData, err := h.registrationService.RegisterUser(c.Request.Context(), userDTO)
	// todo: импортировать ошибку не из repositories?
	if errors.Is(err, repositories.ErrUserAlreadyExists) {
		c.JSON(http.StatusConflict, gin.H{"errors": "User with given login already exists"})
		return
	}
	if err != nil {
		log.Error().Msg(fmt.Sprintf("can not register user: %v", err.Error()))
		c.Status(http.StatusInternalServerError)
	}

	c.Header("Authorization", tokenData.Token)
	c.String(http.StatusOK, "User successfully registered")
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

	tokenData, err := h.authService.AuthenticateUser(c.Request.Context(), input.Login, input.Password)
	if errors.Is(err, services.ErrInvalidCredentials) {
		c.JSON(http.StatusUnauthorized, gin.H{"errors": "Invalid login or password"})
		return
	}
	if err != nil {
		log.Error().Msg(fmt.Sprintf("can not login user: %v", err.Error()))
		c.Status(http.StatusInternalServerError)
	}

	c.Header("Authorization", tokenData.Token)
	c.Status(http.StatusOK)
}
