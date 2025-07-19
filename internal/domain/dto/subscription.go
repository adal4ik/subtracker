package dto

import "time"

type CreateSubscriptionRequest struct {
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	UserID      string `json:"user_id"`    // UUID в string
	StartDate   string `json:"start_date"` // "MM-YYYY"
	EndDate     string `json:"end_date,omitempty"`
}

type UpdateSubscriptionRequest struct {
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date,omitempty"`
}

type SubscriptionResponse struct {
	ID          string `json:"id"`
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	UserID      string `json:"user_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date,omitempty"`
}

type SubscriptionFilter struct {
	UserID      string `json:"user_id"`
	ServiceName string `json:"service_name"`
	MinPrice    int    `json:"min_price"`
	MaxPrice    int    `json:"max_price"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	HasEndDate  *bool  `json:"has_end_date"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
}

type CostFilter struct {
	UserID      string    `json:"user_id"`
	ServiceName string    `json:"service_name"`
	PeriodStart time.Time `json:"period_start"` // "MM-YYYY"
	PeriodEnd   time.Time `json:"period_end"`   // "MM-YYYY"
}

type CostResponse struct {
	TotalCost int `json:"total_cost"`
}
