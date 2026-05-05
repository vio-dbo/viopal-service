package model

import "time"

// Balance
type BalanceResponse struct {
	Balance float64    `json:"balance"`
	Error   *ErrorBody `json:"error,omitempty"`
}

// Top Up
type TopUpRequest struct {
	Amount float64 `json:"amount"`
}

type TopUpResponse struct {
	Data  TopUpRes   `json:"data"`
	Error *ErrorBody `json:"error,omitempty"`
}

type TopUpRes struct {
	ID          int64      `json:"id,omitempty"`
	RefNumber   string     `json:"ref_number,omitempty"`
	Amount      float64    `json:"amount,omitempty"`
	Status      string     `json:"status,omitempty"`
	RequestAt   time.Time  `json:"request_at,omitempty"`
	ProcessedBy *int       `json:"processed_by,omitempty"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	Merchant    string     `json:"merchant,omitempty"`
	MerchantID  int        `json:"merchant_id,omitempty"`
}

type TopUp struct {
	RefNumber   string
	MerchantID  int
	Amount      float64
	Status      string
	RequestAt   time.Time
	ProcessedBy *int
	ProcessedAt *time.Time
}

type TopUpLog struct {
	ID            int64
	TopUpID       int64
	Event         string
	ActorType     string
	ActorID       int
	OldStatus     string
	NewStatus     string
	BalanceBefore float64
	BalanceAfter  float64
	CreatedAt     time.Time
}

// List Top Up
type ListTopUpResponse struct {
	Data  []TopUpRes `json:"data"`
	Error *ErrorBody `json:"error,omitempty"`
}

type ListTopUpRequest struct {
	MerchantID int    `json:"merchant_id,omitempty"`
	RefNumber  string `json:"ref_number,omitempty"`
	Status     string `json:"status,omitempty"`
	Page       int    `json:"page,omitempty"`
	PerPage    int    `json:"perpage,omitempty"`
}

// Update Top Up
type UpdateTopUpRequest struct {
	RefNumber   string
	Status      string `json:"status,omitempty"`
	ProcessedBy *int
	ProcessedAt *time.Time
}

type UpdateTopUpRequestBody struct {
	Status string `json:"status" example:"SUCCESS" enums:"SUCCESS,FAILED"`
}
