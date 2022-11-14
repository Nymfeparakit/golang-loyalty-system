package repositories

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/jmoiron/sqlx"
	"gophermart/internal/app/domain"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user domain.UserDTO) error {
	query := `INSERT INTO auth_user (login, password) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, user.Login, user.Password)

	var pgErr pgx.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return ErrUserAlreadyExists
		}
	}

	return err
}

func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*domain.UserDTO, error) {
	query := `SELECT id, login, password FROM auth_user WHERE login=$1`
	var existingUser domain.UserDTO
	err := r.db.QueryRowxContext(ctx, query, login).StructScan(&existingUser)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserDoesNotExist
	}
	if err != nil {
		return nil, err
	}

	return &existingUser, nil
}

func (r *UserRepository) IncreaseBalanceForOrder(ctx context.Context, orderNumber string, accrual float32) error {
	// находим пользователя, для которого сущесвует заказ
	// и прибавляем ему баланс
	query := `
	UPDATE auth_user u SET balance=balance+$1
	FROM user_order o
	WHERE o.user_id = u.id 
	AND o.number = $2
	`
	_, err := r.db.ExecContext(ctx, query, accrual, orderNumber)
	if err != nil {
		return err
	}

	return nil
}
