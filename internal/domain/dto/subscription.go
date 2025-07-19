package dto

import "time"

type CreateSubscriptionRequest struct {
	ServiceName string `json:"service_name" validate:"required,max=100" example:"Yandex Plus"`
	Price       int    `json:"price"        validate:"required,gte=0"   example:"299"`
	UserID      string `json:"user_id"      validate:"required,uuid4"   example:"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"`
	StartDate   string `json:"start_date"   validate:"required,datetime=01-2006" example:"07-2025"`
	EndDate     string `json:"end_date,omitempty" validate:"omitempty,datetime=01-2006" example:"08-2026"`
}

type UpdateSubscriptionRequest struct {
	ServiceName string `json:"service_name" validate:"required,max=100" example:"Yandex Plus Family"`
	Price       int    `json:"price"        validate:"required,gte=0"   example:"499"`
	StartDate   string `json:"start_date"   validate:"required,datetime=01-2006" example:"07-2025"`
	EndDate     string `json:"end_date,omitempty" validate:"omitempty,datetime=01-2006" example:"08-2027"`
}

type SubscriptionResponse struct {
	ID          string `json:"id" example:"d290f1ee-6c54-4b01-90e6-d701748f0851"`
	ServiceName string `json:"service_name" example:"Yandex Plus"`
	Price       int    `json:"price" example:"299"`
	UserID      string `json:"user_id" example:"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"`
	StartDate   string `json:"start_date" example:"07-2025"`
	EndDate     string `json:"end_date,omitempty" example:"08-2026"`
}

type SubscriptionFilter struct {
	UserID      string `form:"user_id"      validate:"omitempty,uuid4"`
	ServiceName string `form:"service_name" validate:"omitempty,max=100"`
	MinPrice    int    `form:"min_price"    validate:"omitempty,gte=0"`
	MaxPrice    int    `form:"max_price"    validate:"omitempty,gte=0,gtefield=MinPrice"`
	StartDate   string `form:"start_date"   validate:"omitempty,datetime=01-2006"`
	EndDate     string `form:"end_date"     validate:"omitempty,datetime=01-2006"`
	HasEndDate  *bool  `form:"has_end_date" validate:"omitempty"`
	Limit       int    `form:"limit"        validate:"gte=0,lte=100"`
	Offset      int    `form:"offset"       validate:"gte=0"`
}

type CostRequest struct {
	UserID      string `form:"user_id"      validate:"required,uuid4"`
	ServiceName string `form:"service_name" validate:"omitempty,max=100"`
	PeriodStart string `form:"period_start" validate:"required,datetime=01-2006"`
	PeriodEnd   string `form:"period_end"   validate:"required,datetime=01-2006"`
}

type CostFilter struct {
	UserID      string
	ServiceName string
	PeriodStart time.Time
	PeriodEnd   time.Time
}

type CostResponse struct {
	TotalCost int `json:"total_cost" example:"2434"`
}
