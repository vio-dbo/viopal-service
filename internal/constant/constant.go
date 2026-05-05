package constant

const (
	ErrValidation = "VALIDATION_ERROR"
	ErrNotFound   = "NOT_FOUND"
	ErrInternal   = "INTERNAL_SERVER_ERROR"
)

const (
	TopUpStatusPending = "PENDING"
	TopUpStatusSuccess = "SUCCESS"
	TopUpStatusFailed  = "FAILED"
)

const (
	InvoiceStatusPending = "PENDING"
	InvoiceStatusPaid    = "PAID"
	InvoiceStatusExpired = "EXPIRED"
)

const (
	PaymentIntentStatusPending = "PENDING"
	PaymentIntentStatusSuccess = "SUCCESS"
	PaymentIntentStatusFailed  = "FAILED"
)

const (
	RefundStatusRequested = "REQUESTED"
	RefundStatusApproved  = "APPROVED" // decision
	RefundStatusRejected  = "REJECTED" // decision
	RefundStatusSuccess   = "SUCCESS"  // process
	RefundStatusFailed    = "FAILED"   // process
)
