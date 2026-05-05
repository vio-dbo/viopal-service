package model

import (
	"encoding/json"
	"time"
)

type PaymentIntent struct {
	ID                  int64      `json:"id,omitempty"`
	PaymentIntentNumber string     `json:"payment_intent_number,omitempty"`
	InvoiceID           int64      `json:"invoice_id,omitempty"`
	PaymentMethodID     int        `json:"payment_method_id,omitempty"`
	PaymentMethod       string     `json:"payment_method,omitempty"`
	Amount              float64    `json:"amount,omitempty"`
	PayedBy             int64      `json:"payed_by,omitempty"`
	ApprovedBy          *int64     `json:"approved_by,omitempty"`    // nullable
	ApprovedAt          *time.Time `json:"approved_at,omitempty"`    // nullable
	FailureReason       *string    `json:"failure_reason,omitempty"` // nullable
	ExpiredAt           time.Time  `json:"expired_at,omitempty"`
	Status              string     `json:"status"`
	CreatedAt           time.Time  `json:"created_at"`
	InvoiceNumber       string     `json:"invoice_number"`
	MerchantID          int        `json:"merchant_id,omitempty"`
}

type PaymentLog struct {
	ID              int64           `json:"id"`
	PaymentIntentID int64           `json:"payment_intent_id"`
	Event           string          `json:"event"`
	ActorType       string          `json:"actor_type"`
	ActorID         int64           `json:"actor_id"`
	OldStatus       string          `json:"old_status"`
	NewStatus       string          `json:"new_status"`
	MetaData        json.RawMessage `json:"meta_data"` // JSON field
	CreatedAt       time.Time       `json:"created_at"`
}

type PaymentResponse struct {
	Data  PaymentIntent `json:"data"`
	Error *ErrorBody    `json:"error,omitempty"`
}

type PaymentIntentRequest struct {
	PaymentMethodID int `json:"payment_method_id,omitempty"`
}

type ListPaymentIntentRequest struct {
	PaymentIntentNumber string `json:"payment_intent_number,omitempty"`
	PaymentMethodID     int    `json:"payment_method_id,omitempty"`
	Status              string `json:"status,omitempty"`
	Page                int    `json:"page,omitempty"`
	PerPage             int    `json:"perpage,omitempty"`
}

type ListPaymentIntentResponse struct {
	Data  []PaymentIntent `json:"data"`
	Error *ErrorBody      `json:"error,omitempty"`
}

// Update
type UpdatePaymentIntentRequest struct {
	PaymentIntentNumber string
	FailureReason       string `json:"failur_reason"`
	Status              string `json:"status"`
	ApprovedBy          *int
	ApprovedAt          *time.Time
}

type UpdatePaymentIntentRequestBody struct {
	Status        string `json:"status" example:"SUCCESS" enums:"SUCCESS,FAILED"`
	FailureReason string `json:"failur_reason"`
}
