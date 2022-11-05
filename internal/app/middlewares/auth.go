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
	GetUserByUsername(ctx context.Context, username string) (*domain.UserDTO, error)
}

type AuthService interface {
	AddUserToContext(ctx context.Context, user *domain.UserDTO) context.Context
}

func TokenAuthMiddleware(userService UserService, authService AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenHeader := c.Request.Header["Authorization"]
		if tokenHeader == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenString := strings.Split(tokenHeader[0], " ")[1]
		username, err := services.ParseJWTToken(tokenString)
		if errors.Is(err, services.ErrInvalidAccessToken) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		user, err := userService.GetUserByUsername(c.Request.Context(), username)
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
