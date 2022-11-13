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
	"gophermart/internal/app/services"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOrderHandler_HandleCreateOrder(t *testing.T) {
	type WantResponse struct {
		statusCode   int
		responseData interface{}
		contentType  string
	}
	type ErrorResponse struct {
		Errors string `json:"errors"`
	}

	orderUploadedTime := time.Now()
	orderNumber := "123"
	order := domain.OrderDTO{
		Number:     orderNumber,
		UploadedAt: orderUploadedTime,
		UserID:     1,
		Status:     "NEW",
	}

	tests := []struct {
		name                   string
		want                   WantResponse
		createOrderRes         *domain.OrderDTO
		createOrderCreated     bool
		createOrderErr         error
		validateNumberRes      bool
		shouldCallOrderService bool
		reqBody                string
		validateOrderNumberRes bool
	}{
		{
			name:               "positive test #1",
			createOrderRes:     &order,
			createOrderCreated: true,
			want: WantResponse{
				statusCode:   http.StatusAccepted,
				responseData: order,
			},
			shouldCallOrderService: true,
			reqBody:                orderNumber,
			validateOrderNumberRes: true,
		},
		{
			name:               "positive test #2 - order already exists",
			createOrderRes:     &order,
			createOrderCreated: false,
			want: WantResponse{
				statusCode:   http.StatusOK,
				responseData: order,
			},
			shouldCallOrderService: true,
			reqBody:                orderNumber,
			validateOrderNumberRes: true,
		},
		{
			name: "negative test #1 - empty request body",
			want: WantResponse{
				statusCode:   http.StatusBadRequest,
				responseData: ErrorResponse{Errors: "Request body can not be empty"},
			},
			shouldCallOrderService: false,
		},
		{
			name: "negative test #2 - invalid order number format",
			want: WantResponse{
				statusCode:   http.StatusUnprocessableEntity,
				responseData: ErrorResponse{Errors: "Order number is not valid"},
			},
			validateOrderNumberRes: false,
			reqBody:                orderNumber,
			shouldCallOrderService: false,
		},
		{
			name: "negative test #3 - order was uploaded by other user",
			want: WantResponse{
				statusCode:   http.StatusConflict,
				responseData: ErrorResponse{Errors: "order already exists for other user"},
			},
			validateOrderNumberRes: true,
			reqBody:                orderNumber,
			shouldCallOrderService: true,
			createOrderErr:         services.ErrOrderExistsForOtherUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBodyBytes := []byte(tt.reqBody)
			reqBody := bytes.NewReader(reqBodyBytes)
			request := httptest.NewRequest(http.MethodPost, "/", reqBody)
			w := httptest.NewRecorder()

			// создаем хэндлер, в который помещаем мок хранилища и настроек
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			orderServiceMock := mock_handlers.NewMockOrderService(ctrl)
			if tt.shouldCallOrderService {
				orderServiceMock.EXPECT().GetOrCreateOrder(
					request.Context(), gomock.Any(),
				).Return(tt.createOrderRes, tt.createOrderCreated, tt.createOrderErr)
			}
			authServiceMock := mock_handlers.NewMockAuthService(ctrl)
			authServiceMock.EXPECT().GetUserFromContext(request.Context()).Return(&domain.UserDTO{ID: 1}, true)
			orderValidatorMock := mock_handlers.NewMockOrderNumberValidator(ctrl)
			orderValidatorMock.EXPECT().Validate(gomock.Any()).Return(tt.validateOrderNumberRes).AnyTimes()

			r := gin.Default()
			createOrderHandler := NewOrderHandler(authServiceMock, orderServiceMock, orderValidatorMock)
			r.POST("/", createOrderHandler.HandleCreateOrder)
			r.ServeHTTP(w, request)
			result := w.Result()
			err := result.Body.Close()
			require.NoError(t, err)

			// проверяем http статус ответа и тело ответа
			assert.Equal(t, tt.want.statusCode, w.Code)
			expectedResponse, err := json.Marshal(&tt.want.responseData)
			require.NoError(t, err)
			assert.Equal(t, string(expectedResponse), w.Body.String())
		})
	}
}

func TestOrderHandler_HandleListOrders(t *testing.T) {
	type WantResponse struct {
		statusCode   int
		responseData interface{}
		contentType  string
	}

	orders := []*domain.OrderDTO{
		{Number: "123", UploadedAt: time.Now(), UserID: 1, Status: "NEW"},
		{Number: "456", UploadedAt: time.Now(), UserID: 1, Status: "NEW"},
	}
	tests := []struct {
		name               string
		want               WantResponse
		getOrdersByUserRes []*domain.OrderDTO
		getOrdersByUserErr error
	}{
		{
			name:               "positive test #1",
			getOrdersByUserRes: orders,
			want: WantResponse{
				statusCode:   http.StatusOK,
				responseData: orders,
			},
		},
		{
			name:               "positive test #2 - no orders for user",
			getOrdersByUserRes: []*domain.OrderDTO{},
			want: WantResponse{
				statusCode: http.StatusNoContent,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			// создаем хэндлер, в который помещаем мок хранилища и настроек
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userMock := &domain.UserDTO{ID: 1}
			orderServiceMock := mock_handlers.NewMockOrderService(ctrl)
			orderServiceMock.EXPECT().GetOrdersByUser(
				request.Context(), userMock,
			).Return(tt.getOrdersByUserRes, tt.getOrdersByUserErr)
			authServiceMock := mock_handlers.NewMockAuthService(ctrl)
			authServiceMock.EXPECT().GetUserFromContext(request.Context()).Return(userMock, true)
			orderValidatorMock := mock_handlers.NewMockOrderNumberValidator(ctrl)

			r := gin.Default()
			createOrderHandler := NewOrderHandler(authServiceMock, orderServiceMock, orderValidatorMock)
			r.GET("/", createOrderHandler.HandleListOrders)
			r.ServeHTTP(w, request)
			result := w.Result()
			err := result.Body.Close()
			require.NoError(t, err)

			// проверяем http статус ответа и тело ответа
			assert.Equal(t, tt.want.statusCode, w.Code)
			expectedResponse := []byte("")
			if tt.want.responseData != nil {
				expectedResponse, err = json.Marshal(&tt.want.responseData)
				require.NoError(t, err)
			}
			assert.Equal(t, string(expectedResponse), w.Body.String())
		})
	}
}
