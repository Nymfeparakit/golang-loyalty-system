package domain

import "time"

type OrderDTO struct {
	ID         int       `db:"id" json:"-"`
	Number     string    `db:"number" json:"number"`
	UploadedAt time.Time `db:"uploaded_at" json:"uploaded_at"`
	UserID     int       `db:"user_id" json:"-"`
}
