package model

import "time"

type InvoiceRequest struct {
	DueInDays int    `json:"dueInDays"`
	Amount    int64  `json:"amount"`
	CustName  string `json:"cust_name,omitempty"`
	CustEmail string `json:"cust_email,omitempty"`
}

type Invoice struct {
	ID               int64     `json:"id,omitempty"`
	InvoiceNumber    string    `json:"invoice_number"`
	MerchantID       int64     `json:"merchant_id,omitempty"`
	Amount           float64   `json:"amount"`
	Status           string    `json:"status"`
	DueDate          time.Time `json:"due_date"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at,omitempty"`
	CustName         string    `json:"cust_name,omitempty"`
	CustEmail        string    `json:"cust_email,omitempty"`
	Description      string    `json:"description,omitempty"`
	PaymentLinkToken string    `json:"payment_link_token,omitempty"`
}

type InvoiceStatisticReq struct {
	MerchantID       *int   `json:"merchant_id,omitempty"`
	StartCreatedDate string `json:"start_created_date,omitempty"`
	EndCreatedDate   string `json:"end_created_date,omitempty"`
}

type InvoiceStatistic struct {
	CountAll       int     `json:"count_all"`
	CountPaid      int     `json:"count_paid"`
	CountPending   int     `json:"count_failed"`
	CountExpired   int     `json:"count_expired"`
	SumTransaction float64 `json:"sum_transaction"`
	SumRefund      float64 `json:"sum_refund"`
}

type InvoiceStatisticResponse struct {
	Data  InvoiceStatistic `json:"data"`
	Error *ErrorBody       `json:"error,omitempty"`
}

type InvoiceResponse struct {
	Data  Invoice    `json:"data"`
	Error *ErrorBody `json:"error,omitempty"`
}

type ListInvoiceRequest struct {
	Status    string `json:"status,omitempty"`
	CustName  string `json:"cust_name,omitempty"`
	CustEmail string `json:"cust_email,omitempty"`
	Page      int    `json:"page,omitempty"`
	PerPage   int    `json:"perpage,omitempty"`
}

type ListInvoiceResponse struct {
	Data  []Invoice  `json:"data"`
	Error *ErrorBody `json:"error,omitempty"`
}

type InvoiceByCodeResponse struct {
	Data  Invoice    `json:"data"`
	Error *ErrorBody `json:"error,omitempty"`
}

type UpdateInvoiceReq struct {
	ID        int64
	Status    string
	UpdatedAt time.Time
}
