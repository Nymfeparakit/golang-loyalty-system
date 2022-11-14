package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gophermart/internal/app/domain"
	mock_handlers "gophermart/internal/app/handlers/mocks"
	"gophermart/internal/app/repositories"
	"gophermart/internal/app/services"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegistrationHandler_HandleRegistration(t *testing.T) {
	// ожидаемый ответ от сервера
	type WantResponse struct {
		statusCode  int
		response    string
		tokenHeader string
	}

	tokenValue := "123"
	tests := []struct {
		name                 string
		want                 WantResponse
		registerUserRes      *domain.TokenData
		registerUserErr      error
		shouldCallRegService bool
		regInput             registrationInput
	}{
		{
			name:            "positive test",
			registerUserRes: &domain.TokenData{Token: tokenValue},
			want: WantResponse{
				statusCode:  http.StatusOK,
				response:    `User successfully registered`,
				tokenHeader: "Bearer " + tokenValue,
			},
			shouldCallRegService: true,
			regInput: registrationInput{
				Login:    "John",
				Password: "123",
			},
		},
		{
			name:            "user with login exists",
			registerUserErr: repositories.ErrUserAlreadyExists,
			want: WantResponse{
				statusCode: http.StatusConflict,
				response:   `{"errors":"User with given login already exists"}`,
			},
			shouldCallRegService: true,
			regInput: registrationInput{
				Login:    "John",
				Password: "123",
			},
		},
		{
			name: "invalid input - no password field",
			want: WantResponse{
				statusCode: http.StatusBadRequest,
				response:   `{"errors":"Key: 'registrationInput.Password' Error:Field validation for 'Password' failed on the 'required' tag"}`,
			},
			shouldCallRegService: false,
			regInput: registrationInput{
				Login: "John",
			},
		},
		{
			name: "invalid input - no username field",
			want: WantResponse{
				statusCode: http.StatusBadRequest,
				response:   `{"errors":"Key: 'registrationInput.Login' Error:Field validation for 'Login' failed on the 'required' tag"}`,
			},
			shouldCallRegService: false,
			regInput: registrationInput{
				Password: "123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBodyBytes, err := json.Marshal(&tt.regInput)
			require.NoError(t, err)
			reqBody := bytes.NewReader(reqBodyBytes)
			request := httptest.NewRequest(http.MethodPost, "/", reqBody)
			w := httptest.NewRecorder()

			// создаем хэндлер, в который помещаем мок хранилища и настроек
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			regServiceMock := mock_handlers.NewMockRegistrationService(ctrl)
			if tt.shouldCallRegService {
				user := domain.UserDTO{Login: tt.regInput.Login, Password: tt.regInput.Password}
				regServiceMock.EXPECT().RegisterUser(request.Context(), user).Return(tt.registerUserRes, tt.registerUserErr)
			}
			r := gin.Default()
			registrationHandler := NewRegistrationHandler(regServiceMock)
			r.POST("/", registrationHandler.HandleRegistration)
			r.ServeHTTP(w, request)
			result := w.Result()
			err = result.Body.Close()
			require.NoError(t, err)

			// проверяем http статус ответа и тело ответа
			assert.Equal(t, tt.want.statusCode, w.Code)
			assert.Equal(t, tt.want.tokenHeader, result.Header.Get("Authorization"))
			assert.Equal(t, tt.want.response, w.Body.String())
		})
	}
}

func TestLoginHandler_HandleLogin(t *testing.T) {
	// ожидаемый ответ от сервера
	type WantResponse struct {
		statusCode  int
		response    string
		tokenHeader string
	}

	tokenValue := "123"
	logInput := loginInput{
		Login:    "John",
		Password: "123",
	}
	tests := []struct {
		name                  string
		want                  WantResponse
		authUserRes           *domain.TokenData
		authUserErr           error
		shouldCallAuthService bool
		logInput              loginInput
	}{
		{
			name:        "positive test",
			authUserRes: &domain.TokenData{Token: tokenValue},
			want: WantResponse{
				statusCode: http.StatusOK,
			},
			shouldCallAuthService: true,
			logInput:              logInput,
		},
		{
			name:        "invalid credentials",
			authUserErr: services.ErrInvalidCredentials,
			want: WantResponse{
				statusCode: http.StatusUnauthorized,
				response:   `{"errors":"Invalid login or password"}`,
			},
			shouldCallAuthService: true,
			logInput:              logInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBodyBytes, err := json.Marshal(&tt.logInput)
			require.NoError(t, err)
			reqBody := bytes.NewReader(reqBodyBytes)
			request := httptest.NewRequest(http.MethodPost, "/", reqBody)
			w := httptest.NewRecorder()

			// создаем хэндлер, в который помещаем мок хранилища и настроек
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			serviceMock := mock_handlers.NewMockAuthService(ctrl)
			if tt.shouldCallAuthService {
				serviceMock.EXPECT().AuthenticateUser(
					request.Context(), tt.logInput.Login, tt.logInput.Password,
				).Return(tt.authUserRes, tt.authUserErr)
			}

			r := gin.Default()
			registrationHandler := NewLoginHandler(serviceMock)
			r.POST("/", registrationHandler.HandleLogin)
			r.ServeHTTP(w, request)
			result := w.Result()
			err = result.Body.Close()
			require.NoError(t, err)

			// проверяем http статус ответа и тело ответа
			assert.Equal(t, tt.want.statusCode, w.Code)
			assert.Equal(t, tt.want.response, w.Body.String())
		})
	}
}
