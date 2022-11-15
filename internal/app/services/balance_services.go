package services

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"gophermart/internal/app/domain"
	"time"
)

type BalanceRepository interface {
	CreateOrderAndWithdrawBalance(ctx context.Context, order *domain.OrderDTO, sum float32) error
	GetBalanceWithdrawals(ctx context.Context, userID int) ([]*domain.Withdrawal, error)
	GetBalanceAndWithdrawalsSum(ctx context.Context, userID int) (*domain.BalanceData, error)
}

type UserBalanceService struct {
	balanceRepository BalanceRepository
}

func NewUserBalanceService(balanceRepository BalanceRepository) *UserBalanceService {
	return &UserBalanceService{balanceRepository: balanceRepository}
}

func (s *UserBalanceService) GetBalanceAndWithdrawalsSum(ctx context.Context, userID int) (*domain.BalanceData, error) {
	balanceData, err := s.balanceRepository.GetBalanceAndWithdrawalsSum(ctx, userID)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("can not get balance for user: %v", err.Error()))
	}
	return balanceData, err
}

func (s *UserBalanceService) WithdrawBalanceForOrder(ctx context.Context, order *domain.OrderDTO, sum float32) error {
	order.UploadedAt = time.Now()
	// устанавливаем статус сразу в processed для указания того, что заказ не нужно обрабатывать
	order.Status = domain.OrderProcessedStatus
	err := s.balanceRepository.CreateOrderAndWithdrawBalance(ctx, order, sum)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("can not withdraw balance for order: %v", err.Error()))
	}

	return err
}

func (s *UserBalanceService) GetBalanceWithdrawals(ctx context.Context, userID int) ([]*domain.Withdrawal, error) {
	withdrawals, err := s.balanceRepository.GetBalanceWithdrawals(ctx, userID)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("can not get withdrawals for user: %v", err.Error()))
	}
	return withdrawals, err
}
