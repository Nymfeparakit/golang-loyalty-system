package repositories

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"gophermart/internal/app/domain"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user domain.UserDTO) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// создаем пользователя и получаем его id
	var createdUserID int
	query := `INSERT INTO auth_user (login, password) VALUES ($1, $2) RETURNING id`
	err = r.db.QueryRowxContext(ctx, query, user.Login, user.Password).Scan(&createdUserID)

	var pgErr pgx.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return ErrUserAlreadyExists
		}
	}
	if err != nil {
		return err
	}

	// создаем для пользователя также строку в таблице баланса
	query = `INSERT INTO user_balance (user_id) values ($1)`
	_, err = r.db.ExecContext(ctx, query, createdUserID)
	if err != nil {
		return err
	}

	return tx.Commit()
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
	query := `UPDATE user_balance SET current=current+$1
		WHERE user_id = (SELECT user_id FROM user_order WHERE number = $2)
	`
	result, err := r.db.ExecContext(ctx, query, accrual, orderNumber)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		log.Error().Msg("increase user balance: expected one row to be affected")
	}

	return nil
}
