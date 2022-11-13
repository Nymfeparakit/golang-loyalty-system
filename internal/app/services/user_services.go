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
	IncreaseBalanceForOrder(ctx context.Context, orderNumber string, points int) error
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

func (s *UserService) IncreaseBalanceForOrder(ctx context.Context, orderNumber string, accrual int) error {
	return s.userRepository.IncreaseBalanceForOrder(ctx, orderNumber, accrual)
}
