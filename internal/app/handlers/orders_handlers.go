package handlers

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"gophermart/internal/app/domain"
	"gophermart/internal/app/services"
	"io"
	"net/http"
)

type OrderService interface {
	GetOrCreateOrder(ctx context.Context, orderToCreate domain.OrderDTO) (*domain.OrderDTO, bool, error)
	GetOrdersByUser(ctx context.Context, user *domain.UserDTO) ([]*domain.OrderDTO, error)
}

type OrderNumberValidator interface {
	Validate(orderNumber string) bool
}

type OrderHandler struct {
	authService          AuthService
	orderService         OrderService
	orderNumberValidator OrderNumberValidator
}

func NewOrderHandler(authService AuthService, orderService OrderService, orderValidator OrderNumberValidator) *OrderHandler {
	return &OrderHandler{authService: authService, orderService: orderService, orderNumberValidator: orderValidator}
}

func (h *OrderHandler) HandleCreateOrder(c *gin.Context) {
	user, ok := h.authService.GetUserFromContext(c.Request.Context())
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// читаем тело запроса
	b, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal server error")
		return
	}

	if len(b) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"errors": "Request body can not be empty"})
		return
	}
	orderNumber := string(b)
	if !h.orderNumberValidator.Validate(orderNumber) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": "Order number is not valid"})
		return
	}

	// создаем или получаем созданный ранее заказ
	orderToCreate := domain.OrderDTO{
		Number: orderNumber,
		UserID: user.ID,
	}
	order, created, err := h.orderService.GetOrCreateOrder(c.Request.Context(), orderToCreate)
	if errors.Is(err, services.ErrOrderExistsForOtherUser) {
		c.JSON(http.StatusConflict, gin.H{"errors": err.Error()})
		return
	}
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	responseStatus := http.StatusAccepted
	if !created {
		responseStatus = http.StatusOK
	}
	c.JSON(responseStatus, order)
}

func (h *OrderHandler) HandleListOrders(c *gin.Context) {
	user, ok := h.authService.GetUserFromContext(c.Request.Context())
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	orders, err := h.orderService.GetOrdersByUser(c.Request.Context(), user)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, orders)
}
