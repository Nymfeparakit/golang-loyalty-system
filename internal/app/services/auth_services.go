package services

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gophermart/internal/app/domain"
)

type RegistrationService struct {
	userService  *UserService
	tokenService *AuthJWTTokenService
}

func NewRegistrationService(userService *UserService) *RegistrationService {
	return &RegistrationService{userService: userService}
}

func (s *RegistrationService) RegisterUser(ctx context.Context, user domain.UserDTO) (*domain.TokenData, error) {
	err := s.userService.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	// генерируем токены для пользователя
	tokenString, err := s.tokenService.generateAuthToken(user.Login)
	if err != nil {
		return nil, err
	}

	tokenData := domain.TokenData{Token: tokenString}
	return &tokenData, nil
}

type UserCtxKey string

type AuthService struct {
	userService  *UserService
	tokenService *AuthJWTTokenService
}

func NewAuthService(userService *UserService) *AuthService {
	return &AuthService{userService: userService}
}

func (s *AuthService) AuthenticateUser(ctx context.Context, login string, password string) (*domain.TokenData, error) {
	// находим пользователя по логину
	existingUser, err := s.userService.GetUserByLogin(ctx, login)
	if errors.Is(err, ErrUserDoesNotExist) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	// проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	// генерируем токены для пользователя
	tokenString, err := s.tokenService.generateAuthToken(existingUser.Login)
	if err != nil {
		return nil, err
	}

	tokenData := domain.TokenData{Token: tokenString}
	return &tokenData, nil
}

func (s *AuthService) AddUserToContext(ctx context.Context, user *domain.UserDTO) context.Context {
	return context.WithValue(ctx, UserCtxKey("user"), user)
}

func (s *AuthService) GetUserFromContext(ctx context.Context) (*domain.UserDTO, bool) {
	userValue := ctx.Value(UserCtxKey("user"))
	if userValue == nil {
		return nil, false
	}
	userID, ok := userValue.(*domain.UserDTO)
	if !ok {
		return nil, false
	}

	return userID, true
}

func (s *AuthService) ParseUserToken(tokenString string) (string, error) {
	return s.tokenService.parseJWTToken(tokenString)
}

type JWTClaims struct {
	jwt.RegisteredClaims
	Username string
}

type AuthJWTTokenService struct {
}

func (s *AuthJWTTokenService) generateAuthToken(login string) (string, error) {
	claims := JWTClaims{
		Username:         login,
		RegisteredClaims: jwt.RegisteredClaims{},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// todo: ключ для подписи берем из env
	return token.SignedString([]byte("123"))
}

func (s *AuthJWTTokenService) parseJWTToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", ErrInvalidAccessToken
		}
		return []byte("123"), nil
	})

	if err != nil {
		return "", ErrInvalidAccessToken
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims.Username, nil
	}

	return "", ErrInvalidAccessToken
}
