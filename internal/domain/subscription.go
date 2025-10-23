package domain

import (
	"github.com/google/uuid"
	"time"
)

type Subscription struct {
	ID          uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ServiceName string    `json:"service_name" example:"Yandex Plus" binding:"required"`
	Price       int       `json:"price" example:"400" binding:"required,min=0"`
	UserID      uuid.UUID `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba" binding:"required"`
	StartDate   string    `json:"start_date" example:"07-2025" binding:"required"`
	EndDate     *string   `json:"end_date,omitempty" example:"12-2025"`
	CreatedAt   time.Time `json:"created_at" example:"2025-10-23T15:04:05Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2025-10-23T15:04:05Z"`
}

type CreateSubscriptionRequest struct {
	ServiceName string    `json:"service_name" binding:"required" example:"Yandex Plus"`
	Price       int       `json:"price" binding:"required,min=0" example:"400"`
	UserID      uuid.UUID `json:"user_id" binding:"required" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string    `json:"start_date" binding:"required" example:"07-2025"`
	EndDate     *string   `json:"end_date,omitempty" example:"12-2025"`
}

type UpdateSubscriptionRequest struct {
	ServiceName *string `json:"service_name,omitempty" example:"Yandex Plus"`
	Price       *int    `json:"price,omitempty" example:"400"`
	StartDate   *string `json:"start_date,omitempty" example:"07-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"12-2025"`
}

type ListSubscriptionsQuery struct {
	UserID      *string `form:"user_id"`
	ServiceName *string `form:"service_name"`
	Limit       int     `form:"limit" binding:"min=1,max=100"`
	Offset      int     `form:"offset" binding:"min=0"`
}

type CalculateTotalRequest struct {
	UserID      *string `form:"user_id"`
	ServiceName *string `form:"service_name"`
	StartPeriod string  `form:"start_period" binding:"required" example:"01-2025"`
	EndPeriod   string  `form:"end_period" binding:"required" example:"12-2025"`
}

type CalculateTotalResponse struct {
	TotalCost int `json:"total_cost" example:"4800"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"invalid request"`
}

type SuccessResponse struct {
	Message string `json:"message" example:"success"`
}
