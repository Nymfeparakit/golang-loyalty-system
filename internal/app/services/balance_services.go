package services

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"gophermart/internal/app/domain"
)

type BalanceRepository interface {
	WithdrawBalanceForOrder(ctx context.Context, withdrawal *domain.Withdrawal) error
	GetBalanceWithdrawals(ctx context.Context, userID int) ([]*domain.Withdrawal, error)
	GetUserBalance(ctx context.Context, userID int) (*domain.BalanceData, error)
}

type UserBalanceService struct {
	balanceRepository BalanceRepository
}

func NewUserBalanceService(balanceRepository BalanceRepository) *UserBalanceService {
	return &UserBalanceService{balanceRepository: balanceRepository}
}

func (s *UserBalanceService) GetUserBalance(ctx context.Context, userID int) (*domain.BalanceData, error) {
	balanceData, err := s.balanceRepository.GetUserBalance(ctx, userID)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("can not get balance for user: %v", err.Error()))
	}
	return balanceData, err
}

func (s *UserBalanceService) WithdrawBalanceForOrder(ctx context.Context, withdrawal *domain.Withdrawal) error {
	err := s.balanceRepository.WithdrawBalanceForOrder(ctx, withdrawal)
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
