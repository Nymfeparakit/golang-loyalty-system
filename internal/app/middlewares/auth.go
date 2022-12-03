package middlewares

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gophermart/internal/app/domain"
	"gophermart/internal/app/services"
	"net/http"
	"strings"
)

type UserService interface {
	GetUserByLogin(ctx context.Context, username string) (*domain.UserDTO, error)
}

type AuthService interface {
	AddUserToContext(ctx context.Context, user *domain.UserDTO) context.Context
	ParseUserToken(tokenString string) (string, error)
}

func TokenAuthMiddleware(userService UserService, authService AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenHeader := c.Request.Header["Authorization"]
		if tokenHeader == nil {
			log.Error().Msg("there is no auth header in request")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenHeaderStr := strings.Split(tokenHeader[0], " ")
		if len(tokenHeaderStr) != 2 {
			log.Error().Msg(fmt.Sprintf("failed to parse token header: %v", tokenHeader[0]))
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenString := tokenHeaderStr[1]

		username, err := authService.ParseUserToken(tokenString)
		if errors.Is(err, services.ErrInvalidAccessToken) {
			log.Error().Msg(fmt.Sprintf("failed to parse token value: %v", err.Error()))
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if err != nil {
			log.Error().Msg(fmt.Sprintf("failed to parse token value: %v", err.Error()))
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		user, err := userService.GetUserByLogin(c.Request.Context(), username)
		if errors.Is(err, services.ErrUserDoesNotExist) {
			log.Error().Msg(fmt.Sprintf("can not find user with username %s", username))
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if err != nil {
			log.Error().Msg(fmt.Sprintf("can not find user with username %s: %v", username, err.Error()))
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		ctx := authService.AddUserToContext(c.Request.Context(), user)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
