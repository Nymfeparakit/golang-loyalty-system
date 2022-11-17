package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"gophermart/internal/app/domain"
	"time"
)

type BalanceRepository struct {
	db *sqlx.DB
}

func NewBalanceRepository(db *sqlx.DB) *BalanceRepository {
	return &BalanceRepository{db: db}
}

func (r *BalanceRepository) CreateOrderAndWithdrawBalance(ctx context.Context, order *domain.OrderDTO, sum float32) error {
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// сначала создаем заказ
	query := `INSERT INTO user_order (number, uploaded_at, user_id, status) VALUES ($1, $2, $3, $4)`
	_, err = r.db.ExecContext(ctx, query, &order.Number, &order.UploadedAt, &order.UserID, &order.Status)
	var pgErr pgx.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return ErrOrderAlreadyExists
		}
	}
	if err != nil {
		return err
	}

	// затем пытаемся вычесть сумму заказа из баланса пользователя
	// для этого сначала проверяем значение баланса
	var balance float32
	query = `SELECT balance FROM auth_user where id=$1`
	if err := r.db.QueryRowContext(ctx, query, order.UserID).Scan(&balance); err != nil {
		return err
	}
	if balance < sum {
		return ErrCanNotWithdrawBalance
	}

	query = `UPDATE auth_user SET balance=balance-$1 WHERE id=$2`
	_, err = r.db.ExecContext(ctx, query, &sum, &order.UserID)
	if err != nil {
		return err
	}

	// записываем в историю withdrawal
	query = `INSERT INTO withdrawal (processed_at, sum, order_number) VALUES ($1, $2, $3)`
	_, err = r.db.ExecContext(ctx, query, time.Now(), &sum, &order.Number)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *BalanceRepository) GetBalanceWithdrawals(ctx context.Context, userID int) ([]*domain.Withdrawal, error) {
	query := `SELECT w.order_number, w.processed_at, w.sum
		FROM withdrawal w
		JOIN user_order o ON o.number = w.order_number
		JOIN auth_user u ON u.id = o.user_id
		WHERE u.id = $1
  		ORDER BY processed_at
	`

	rows, err := r.db.QueryxContext(ctx, query, &userID)
	var withdrawals []*domain.Withdrawal
	if errors.Is(err, sql.ErrNoRows) {
		return withdrawals, nil
	}
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var withdrawal domain.Withdrawal
		err := rows.StructScan(&withdrawal)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, &withdrawal)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func (r *BalanceRepository) GetBalanceAndWithdrawalsSum(ctx context.Context, userID int) (*domain.BalanceData, error) {
	query := `SELECT u.balance, coalesce(sum(w.sum), 0) as withdrawn
		FROM auth_user u
		LEFT JOIN user_order o on u.id = o.user_id
		LEFT JOIN withdrawal w on w.order_number = o.number
		WHERE u.id = $1
		GROUP BY u.balance
	`

	balanceData := &domain.BalanceData{}
	err := r.db.QueryRowxContext(ctx, query, userID).StructScan(balanceData)
	if err != nil {
		return nil, err
	}
	log.Info().Msg(fmt.Sprintf("got balance data for user %d, data: %v", userID, *balanceData))

	return balanceData, nil
}
