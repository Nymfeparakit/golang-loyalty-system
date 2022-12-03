package services

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gophermart/internal/app/domain"
	"gophermart/internal/app/repositories"
)

const PasswordHashCost = 10

type UserRepository interface {
	CreateUser(ctx context.Context, user domain.UserDTO) error
	GetUserByLogin(ctx context.Context, username string) (*domain.UserDTO, error)
	IncreaseBalanceAndUpdateOrderStatus(ctx context.Context, orderNumber string, accrual float32, orderStatus string) error
}

type UserService struct {
	userRepository UserRepository
}

func NewUserService(userRepository UserRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

func (s *UserService) CreateUser(ctx context.Context, user domain.UserDTO) error {
	// хэшируем пароль пользователя
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(user.Password), PasswordHashCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPwd)
	return s.userRepository.CreateUser(ctx, user)
}

func (s *UserService) GetUserByLogin(ctx context.Context, username string) (*domain.UserDTO, error) {
	user, err := s.userRepository.GetUserByLogin(ctx, username)
	if errors.Is(err, repositories.ErrUserDoesNotExist) {
		return nil, ErrUserDoesNotExist
	}

	return user, err
}

func (s *UserService) IncreaseBalanceAndUpdateOrderStatus(ctx context.Context, orderNumber string, accrual float32, orderStatus string) error {
	return s.userRepository.IncreaseBalanceAndUpdateOrderStatus(ctx, orderNumber, accrual, orderStatus)
}
