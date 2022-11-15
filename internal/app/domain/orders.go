package domain

import "time"

const OrderProcessedStatus = "PROCESSED"
const OrderProcessingStatus = "PROCESSING"

type OrderDTO struct {
	Number     string    `db:"number" json:"number"`
	UploadedAt time.Time `db:"uploaded_at" json:"uploaded_at"`
	UserID     int       `db:"user_id" json:"-"`
	Status     string    `db:"status" json:"status"`
}

type AccrualCalculationRes struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}
