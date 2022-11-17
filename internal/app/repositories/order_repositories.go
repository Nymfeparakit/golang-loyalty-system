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

type OrderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) GetOrCreateOrder(ctx context.Context, orderToCreate domain.OrderDTO) (*domain.OrderDTO, bool, error) {
	query := `INSERT INTO user_order (number, uploaded_at, user_id) VALUES ($1, $2, $3) RETURNING *`
	var order domain.OrderDTO
	err := r.db.QueryRowxContext(
		ctx,
		query,
		&orderToCreate.Number,
		&orderToCreate.UploadedAt,
		&orderToCreate.UserID,
	).StructScan(&order)
	var pgErr pgx.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			query := `SELECT * FROM user_order WHERE number=$1`
			if err := r.db.QueryRowxContext(ctx, query, &orderToCreate.Number).StructScan(&order); err != nil {
				return nil, false, err
			}
			return &order, false, nil
		}
	}

	return &order, true, nil
}

func (r *OrderRepository) GetOrdersByUser(ctx context.Context, user *domain.UserDTO) ([]*domain.OrderDTO, error) {
	query := `SELECT * FROM user_order WHERE user_id=$1  ORDER BY uploaded_at`
	var orders []*domain.OrderDTO
	rows, err := r.db.QueryxContext(ctx, query, user.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return orders, nil
	}
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var order domain.OrderDTO
		err := rows.StructScan(&order)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	err = rows.Err()
	if err != nil {
		return orders, err
	}

	return orders, nil
}

func (r *OrderRepository) UpdateOrderStatusAndAccrual(
	ctx context.Context,
	orderNumber string,
	orderStatus string,
	accrual float32,
) error {
	query := `UPDATE user_order SET status=$1, accrual=$2 WHERE number=$3`
	_, err := r.db.ExecContext(ctx, query, &orderStatus, &accrual, &orderNumber)
	if err != nil {
		return err
	}

	return nil
}

func (r *OrderRepository) GetOrdersWithStatusesNotIn(ctx context.Context, statuses []string) ([]*domain.OrderDTO, error) {
	query := `SELECT * FROM user_order WHERE status NOT IN (?)`
	query, args, err := sqlx.In(query, statuses)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	rows, err := r.db.QueryxContext(ctx, query, args...)

	var orders []*domain.OrderDTO
	for rows.Next() {
		var order domain.OrderDTO
		err := rows.StructScan(&order)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	err = rows.Err()
	if err != nil {
		return orders, err
	}

	return orders, nil
}
