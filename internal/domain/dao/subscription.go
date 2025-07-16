package dao

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionRow struct {
	ID          uuid.UUID  `db:"id"`
	UserID      uuid.UUID  `db:"user_id"`
	ServiceName string     `db:"service_name"`
	Price       int        `db:"price"`
	StartDate   time.Time  `db:"start_date"`
	EndDate     *time.Time `db:"end_date"`
}
