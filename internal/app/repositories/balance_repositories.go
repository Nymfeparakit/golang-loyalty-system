package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func (r *BalanceRepository) WithdrawBalanceForOrder(ctx context.Context, withdrawal *domain.Withdrawal) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// пытаемся вычесть сумму заказа из баланса пользователя
	// для этого сначала проверяем значение баланса
	var balance float32
	query := `SELECT current FROM user_balance where user_id=$1`
	if err := r.db.QueryRowContext(ctx, query, withdrawal.UserID).Scan(&balance); err != nil {
		return err
	}
	wSum := withdrawal.Sum
	if balance < wSum {
		return ErrCanNotWithdrawBalance
	}

	query = `UPDATE user_balance SET current=current-$1, withdrawn=withdrawn+$1 WHERE user_id=$2`
	_, err = r.db.ExecContext(ctx, query, &wSum, &withdrawal.UserID)
	if err != nil {
		return err
	}

	// записываем в историю withdrawal
	query = `INSERT INTO withdrawal (processed_at, sum, order_number, user_id) VALUES ($1, $2, $3, $4)`
	_, err = r.db.ExecContext(ctx, query, time.Now(), &wSum, &withdrawal.Order, &withdrawal.UserID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *BalanceRepository) GetBalanceWithdrawals(ctx context.Context, userID int) ([]*domain.Withdrawal, error) {
	query := `SELECT *
		FROM withdrawal
		WHERE user_id = $1
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

func (r *BalanceRepository) GetUserBalance(ctx context.Context, userID int) (*domain.BalanceData, error) {
	query := `SELECT current, withdrawn
		FROM user_balance
		WHERE user_id = $1
	`

	balanceData := &domain.BalanceData{}
	err := r.db.QueryRowxContext(ctx, query, userID).StructScan(balanceData)
	if err != nil {
		return nil, err
	}
	log.Info().Msg(fmt.Sprintf("got balance data for user %d, data: %v", userID, *balanceData))

	return balanceData, nil
}
