package model

import "time"

type Refund struct {
	ID                  int64      `json:"id,omitempty"`
	RefundNumber        string     `json:"refund_number"`
	Reason              string     `json:"reason"`
	PaymentIntentID     int64      `json:"payment_intent_id,omitempty"`
	PaymentIntentNumber string     `json:"payment_intent_number,omitempty"`
	InvoiceID           int64      `json:"invoice_id,omitempty"`
	InvoiceNumber       string     `json:"invoice_number,omitempty"`
	MerchantID          int        `json:"merchant_id,omitempty"`
	Merchant            string     `json:"merchant,omitempty"`
	Amount              float64    `json:"amount"`
	Status              string     `json:"status"`
	RequestBy           int        `json:"request_by,omitempty"`
	DecidedBy           *int       `json:"decided_by,omitempty"`   // nullable
	DecidedAt           *time.Time `json:"decided_at,omitempty"`   // nullable
	ProcessedBy         *int       `json:"processed_by,omitempty"` // nullable
	ProcessedAt         *time.Time `json:"processed_at,omitempty"` // nullable
	CreatedAt           *time.Time `json:"created_at,omitempty"`
}

type RefundLog struct {
	ID        int64     `json:"id"`
	RefundID  int64     `json:"refund_id"`
	Event     string    `json:"event"`
	ActorType string    `json:"actor_type"`
	ActorID   int       `json:"actor_id"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	CreatedAt time.Time `json:"created_at"`
}

type RefundResponse struct {
	Data  Refund     `json:"data"`
	Error *ErrorBody `json:"error,omitempty"`
}

type RefundDecideRequest struct {
	Status       string    `json:"status"  example:"APPROVE" enums:"APPROVE,REJECT"`
	RefundNumber string    `json:"refund_number"`
	DecidedBy    int       `json:"decided_by"`
	DecidedAt    time.Time `json:"decided_at"`
}

type RefundDecideRequestBody struct {
	Status string `json:"status"  example:"APPROVE" enums:"APPROVE,REJECT"`
}

type RefundProcessRequest struct {
	Status       string    `json:"status"  example:"SUCCESS" enums:"SUCCESS,FAILED"`
	RefundNumber string    `json:"refund_number"`
	ProcessedBy  int       `json:"processed_by"`
	ProcessedAt  time.Time `json:"processed_at"`
}

type RefundRequest struct {
	PaymentIntentCode string `json:"payment_intent_code"`
	Reason            string `json:"reason"`
}

type ListRefundRequest struct {
	MerchantID int    `json:"merchant_id,omitempty"`
	Status     string `json:"status,omitempty"`
	Page       int    `json:"page,omitempty"`
	PerPage    int    `json:"perpage,omitempty"`
}

type ListRefundResponse struct {
	Data  []Refund   `json:"data"`
	Error *ErrorBody `json:"error,omitempty"`
}
