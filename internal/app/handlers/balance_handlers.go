package handlers

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"gophermart/internal/app/domain"
	"gophermart/internal/app/repositories"
	"net/http"
)

type UserBalanceService interface {
	WithdrawBalanceForOrder(ctx context.Context, withdrawal *domain.Withdrawal) error
	GetBalanceWithdrawals(ctx context.Context, userID int) ([]*domain.Withdrawal, error)
	GetUserBalance(ctx context.Context, userID int) (*domain.BalanceData, error)
}

type UserBalanceHandler struct {
	orderNumberValidator OrderNumberValidator
	authService          AuthService
	balanceService       UserBalanceService
}

func NewUserBalanceHandler(
	orderNumberValidator OrderNumberValidator, authService AuthService, balanceService UserBalanceService,
) *UserBalanceHandler {
	return &UserBalanceHandler{
		orderNumberValidator: orderNumberValidator,
		authService:          authService,
		balanceService:       balanceService,
	}
}

func (h *UserBalanceHandler) HandleGetUserBalance(c *gin.Context) {
	user, ok := h.authService.GetUserFromContext(c.Request.Context())
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	balanceData, err := h.balanceService.GetUserBalance(c.Request.Context(), user.ID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, balanceData)
}

// HandleWithdrawBalance обрабатывает POST запрос для оплаты заказа пользователя в счет баллов со счета
func (h *UserBalanceHandler) HandleWithdrawBalance(c *gin.Context) {
	user, ok := h.authService.GetUserFromContext(c.Request.Context())
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var input BalanceWithdrawalInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": err.Error()})
		return
	}

	if !h.orderNumberValidator.Validate(input.OrderNumber) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": "Order number is not valid"})
		return
	}

	withdrawalToCreate := &domain.Withdrawal{
		Order:  input.OrderNumber,
		Sum:    input.Sum,
		UserID: user.ID,
	}
	err := h.balanceService.WithdrawBalanceForOrder(c.Request.Context(), withdrawalToCreate)
	if errors.Is(err, repositories.ErrOrderAlreadyExists) {
		c.JSON(http.StatusConflict, gin.H{"errors": "Order already exists"})
		return
	}
	if errors.Is(err, repositories.ErrCanNotWithdrawBalance) {
		c.JSON(http.StatusPaymentRequired, gin.H{"errors": "Not enough points in user's balance"})
		return
	}
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}

func (h *UserBalanceHandler) HandleListBalanceWithdrawals(c *gin.Context) {
	user, ok := h.authService.GetUserFromContext(c.Request.Context())
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.balanceService.GetBalanceWithdrawals(c.Request.Context(), user.ID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, &withdrawals)
}
