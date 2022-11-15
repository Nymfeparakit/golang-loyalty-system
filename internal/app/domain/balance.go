package domain

import "time"

type BalanceData struct {
	Balance        float32 `json:"balance" db:"balance"`
	WithdrawalsSum float32 `json:"withdrawn" db:"withdrawn"`
}

type Withdrawal struct {
	Order       string    `json:"order" db:"order_number"`
	Sum         float32   `json:"sum" db:"sum"`
	ProcessedAt time.Time `json:"processed_at" db:"processed_at"`
}
