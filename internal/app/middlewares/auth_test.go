package middlewares

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gophermart/internal/app/domain"
	mock_middlewares "gophermart/internal/app/middlewares/mocks"
	"gophermart/internal/app/services"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTokenAuthMiddleware(t *testing.T) {
	tokenValue := "123"
	username := "abc"
	user := domain.UserDTO{
		Login: username,
	}

	tests := []struct {
		name                 string
		tokenValue           string
		passTokenHeader      bool
		ParseUserTokenRes    string
		ParseUserTokenErr    error
		GetUserByUsernameRes *domain.UserDTO
		GetUserByUsernameErr error
		wantStatusCode       int
	}{
		{
			name:            "no authorization header",
			passTokenHeader: false,
			wantStatusCode:  http.StatusUnauthorized,
		},
		{
			name:              "authorization token is not valid",
			passTokenHeader:   true,
			tokenValue:        tokenValue,
			ParseUserTokenRes: "",
			ParseUserTokenErr: services.ErrInvalidAccessToken,
			wantStatusCode:    http.StatusUnauthorized,
		},
		{
			name:                 "user with username from token does not exist",
			passTokenHeader:      true,
			tokenValue:           tokenValue,
			ParseUserTokenRes:    username,
			GetUserByUsernameErr: services.ErrUserDoesNotExist,
			wantStatusCode:       http.StatusUnauthorized,
		},
		{
			name:                 "positive test",
			passTokenHeader:      true,
			tokenValue:           tokenValue,
			ParseUserTokenRes:    username,
			GetUserByUsernameRes: &user,
			wantStatusCode:       http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			// создаем моки
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			authServiceMock := mock_middlewares.NewMockAuthService(ctrl)
			authServiceMock.EXPECT().ParseUserToken(
				tt.tokenValue,
			).Return(tt.ParseUserTokenRes, tt.ParseUserTokenErr).AnyTimes()
			authServiceMock.EXPECT().AddUserToContext(request.Context(), &user).Return(request.Context()).AnyTimes()
			userServiceMock := mock_middlewares.NewMockUserService(ctrl)
			userServiceMock.EXPECT().GetUserByUsername(
				request.Context(), tt.ParseUserTokenRes,
			).Return(tt.GetUserByUsernameRes, tt.GetUserByUsernameErr).AnyTimes()

			router := gin.Default()
			router.Use(TokenAuthMiddleware(userServiceMock, authServiceMock))
			router.GET("/", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			require.NoError(t, err)
			if tt.passTokenHeader {
				request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tt.tokenValue))
			}

			router.ServeHTTP(w, request)
			result := w.Result()
			err = result.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, result.StatusCode)
		})
	}
}
