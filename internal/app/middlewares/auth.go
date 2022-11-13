package middlewares

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
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
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenHeaderStr := strings.Split(tokenHeader[0], " ")
		if len(tokenHeaderStr) != 2 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenString := tokenHeaderStr[1]

		username, err := authService.ParseUserToken(tokenString)
		if errors.Is(err, services.ErrInvalidAccessToken) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		user, err := userService.GetUserByLogin(c.Request.Context(), username)
		if errors.Is(err, services.ErrUserDoesNotExist) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		ctx := authService.AddUserToContext(c.Request.Context(), user)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
