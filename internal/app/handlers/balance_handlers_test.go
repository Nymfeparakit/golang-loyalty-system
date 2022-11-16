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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUserBalanceHandler_HandleGetUserBalance(t *testing.T) {
	type WantResponse struct {
		statusCode   int
		responseData interface{}
		contentType  string
	}

	balanceData := &domain.BalanceData{Current: 100, WithdrawalsSum: 50}
	user := &domain.UserDTO{ID: 1}
	tests := []struct {
		name                        string
		getBalanceAndWithdrawalsRes *domain.BalanceData
		want                        WantResponse
	}{
		{
			name:                        "positive test",
			getBalanceAndWithdrawalsRes: balanceData,
			want: WantResponse{
				statusCode:   http.StatusOK,
				responseData: balanceData,
				contentType:  "application/json; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			authServiceMock := mock_handlers.NewMockAuthService(ctrl)
			authServiceMock.EXPECT().GetUserFromContext(request.Context()).Return(user, true)
			balanceServiceMock := mock_handlers.NewMockUserBalanceService(ctrl)
			balanceServiceMock.EXPECT().GetBalanceAndWithdrawalsSum(
				request.Context(), user.ID,
			).Return(tt.getBalanceAndWithdrawalsRes, nil)
			orderValidatorMock := mock_handlers.NewMockOrderNumberValidator(ctrl)

			r := gin.Default()
			userBalanceHandler := NewUserBalanceHandler(orderValidatorMock, authServiceMock, balanceServiceMock)
			r.GET("/", userBalanceHandler.HandleGetUserBalance)
			r.ServeHTTP(w, request)
			result := w.Result()
			err := result.Body.Close()
			require.NoError(t, err)

			// проверяем http статус ответа и тело ответа
			assert.Equal(t, tt.want.statusCode, w.Code)
			expectedResponse, err := json.Marshal(&tt.want.responseData)
			require.NoError(t, err)
			assert.Equal(t, string(expectedResponse), w.Body.String())
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}

func TestUserBalanceHandler_HandleWithdrawBalance(t *testing.T) {
	type WantErrorResponseBody struct {
		Errors string `json:"errors"`
	}

	user := &domain.UserDTO{ID: 1}
	orderNumber := "123"
	inputData := &BalanceWithdrawalInput{OrderNumber: orderNumber, Sum: 100}
	tests := []struct {
		name                       string
		withdrawBalanceForOrderRes error
		wantStatusCode             int
		wantErrRespBody            *WantErrorResponseBody
		reqInput                   *BalanceWithdrawalInput
	}{
		{
			name:                       "positive test",
			withdrawBalanceForOrderRes: nil,
			wantStatusCode:             http.StatusOK,
			reqInput:                   inputData,
		},
		{
			name:                       "order already exists",
			withdrawBalanceForOrderRes: repositories.ErrOrderAlreadyExists,
			wantStatusCode:             http.StatusConflict,
			wantErrRespBody:            &WantErrorResponseBody{Errors: "Order already exists"},
			reqInput:                   inputData,
		},
		{
			name:                       "not enough points in balance",
			withdrawBalanceForOrderRes: repositories.ErrCanNotWithdrawBalance,
			wantErrRespBody:            &WantErrorResponseBody{Errors: "Not enough points in user's balance"},
			wantStatusCode:             http.StatusPaymentRequired,
			reqInput:                   inputData,
		},
		{
			name:           "sum is negative",
			wantStatusCode: http.StatusBadRequest,
			reqInput:       &BalanceWithdrawalInput{OrderNumber: orderNumber, Sum: -100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tt.reqInput)
			reqBodyReader := bytes.NewReader(reqBody)
			request := httptest.NewRequest(http.MethodPost, "/", reqBodyReader)
			w := httptest.NewRecorder()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			authServiceMock := mock_handlers.NewMockAuthService(ctrl)
			authServiceMock.EXPECT().GetUserFromContext(request.Context()).Return(user, true)
			balanceServiceMock := mock_handlers.NewMockUserBalanceService(ctrl)
			balanceServiceMock.EXPECT().WithdrawBalanceForOrder(
				request.Context(), &domain.OrderDTO{Number: orderNumber, UserID: user.ID}, tt.reqInput.Sum,
			).Return(tt.withdrawBalanceForOrderRes).AnyTimes()
			orderValidatorMock := mock_handlers.NewMockOrderNumberValidator(ctrl)
			orderValidatorMock.EXPECT().Validate(tt.reqInput.OrderNumber).Return(true).AnyTimes()

			r := gin.Default()
			userBalanceHandler := NewUserBalanceHandler(orderValidatorMock, authServiceMock, balanceServiceMock)
			r.POST("/", userBalanceHandler.HandleWithdrawBalance)
			r.ServeHTTP(w, request)
			require.NoError(t, err)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			if tt.wantErrRespBody != nil {
				expectedResponse, err := json.Marshal(tt.wantErrRespBody)
				require.NoError(t, err)
				assert.Equal(t, string(expectedResponse), w.Body.String())
			}
		})
	}
}

func TestUserBalanceHandler_HandleListBalanceWithdrawals(t *testing.T) {
	type WantResponse struct {
		contentType  string
		statusCode   int
		responseData interface{}
	}

	withdrawals := []*domain.Withdrawal{
		{Order: "123", Sum: 100, ProcessedAt: time.Now()},
		{Order: "345", Sum: 200, ProcessedAt: time.Now()},
	}
	user := &domain.UserDTO{ID: 1}
	tests := []struct {
		name                  string
		want                  WantResponse
		balanceWithdrawalsRes []*domain.Withdrawal
	}{
		{
			name: "positive test",
			want: WantResponse{
				contentType:  "application/json; charset=utf-8",
				statusCode:   http.StatusOK,
				responseData: withdrawals,
			},
			balanceWithdrawalsRes: withdrawals,
		},
		{
			name: "no data",
			want: WantResponse{
				statusCode: http.StatusNoContent,
			},
			balanceWithdrawalsRes: []*domain.Withdrawal{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			authServiceMock := mock_handlers.NewMockAuthService(ctrl)
			authServiceMock.EXPECT().GetUserFromContext(request.Context()).Return(user, true)
			balanceServiceMock := mock_handlers.NewMockUserBalanceService(ctrl)
			balanceServiceMock.EXPECT().GetBalanceWithdrawals(
				request.Context(), user.ID,
			).Return(tt.balanceWithdrawalsRes, nil)
			orderValidatorMock := mock_handlers.NewMockOrderNumberValidator(ctrl)

			r := gin.Default()
			userBalanceHandler := NewUserBalanceHandler(orderValidatorMock, authServiceMock, balanceServiceMock)
			r.GET("/", userBalanceHandler.HandleListBalanceWithdrawals)
			r.ServeHTTP(w, request)

			assert.Equal(t, tt.want.statusCode, w.Code)
			if tt.want.responseData != nil {
				expectedResponse, err := json.Marshal(tt.want.responseData)
				require.NoError(t, err)
				assert.Equal(t, string(expectedResponse), w.Body.String())
			}
			assert.Equal(t, tt.want.contentType, w.Result().Header.Get("Content-Type"))
		})
	}
}
