package repository

import (
	"context"
	"database/sql"
	"net/http"
	"viopal-service/internal/model"
)

type InvoiceRepository struct {
	db *sql.DB
}

func NewInvoiceRepository(db *sql.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) BeginTxInvoice(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *InvoiceRepository) CreateInvoice(ctx context.Context, tx *sql.Tx, inv model.Invoice) model.InvoiceResponse {
	var data model.InvoiceResponse

	query := `
		INSERT INTO invoices (
			invoice_number,
			merchant_id,
			amount,
			status,
			due_date,
			created_at,
			updated_at,
			cust_name,
			cust_email,
			description,
			payment_link_token
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := tx.Exec(
		query,
		inv.InvoiceNumber,
		inv.MerchantID,
		inv.Amount,
		inv.Status,
		inv.DueDate,
		inv.CreatedAt,
		inv.UpdatedAt,
		inv.CustName,
		inv.CustEmail,
		inv.Description,
		inv.PaymentLinkToken,
	)

	if err != nil {
		data.Error = &model.ErrorBody{
			Code:    http.StatusText(http.StatusInternalServerError),
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		}
		return data
	}

	return data
}

func (r *InvoiceRepository) ListInvoice(merchantID int, req model.ListInvoiceRequest) model.ListInvoiceResponse {
	var data model.ListInvoiceResponse
	offset := (req.Page - 1) * req.PerPage

	query := `
		SELECT invoice_number, amount, status, due_date, created_at, updated_at, cust_name, cust_email, description, payment_link_token
		FROM invoices
		WHERE merchant_id = ?
	`

	args := []interface{}{merchantID}
	if req.Status != "" {
		query += " AND status = ?"
		args = append(args, req.Status)
	}

	if req.CustName != "" {
		query += " AND cust_name LIKE ?"
		args = append(args, "%"+req.CustName+"%")
	}

	if req.CustEmail != "" {
		query += " AND cust_email LIKE ?"
		args = append(args, "%"+req.CustEmail+"%")
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, req.PerPage, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		data.Error = &model.ErrorBody{
			Code:    http.StatusText(http.StatusBadRequest),
			Message: err.Error(),
			Status:  http.StatusBadRequest,
		}
		return data
	}
	defer rows.Close()

	var res []model.Invoice
	for rows.Next() {
		var item model.Invoice
		err := rows.Scan(&item.InvoiceNumber, &item.Amount, &item.Status, &item.DueDate, &item.CreatedAt, &item.UpdatedAt,
			&item.CustName, &item.CustEmail, &item.Description, &item.PaymentLinkToken)
		if err != nil {
			data.Error = &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: err.Error(),
				Status:  http.StatusBadRequest,
			}
			return data
		}

		res = append(res, item)
	}

	if err = rows.Err(); err != nil {
		data.Error = &model.ErrorBody{
			Code:    http.StatusText(http.StatusBadRequest),
			Message: err.Error(),
			Status:  http.StatusBadRequest,
		}
		return data
	}

	data.Data = res
	return data
}

func (r *InvoiceRepository) GetInvoiceByCode(merchantID int, code string) model.InvoiceByCodeResponse {
	var data model.InvoiceByCodeResponse

	query := `
		SELECT invoice_number, amount, status, due_date, created_at, updated_at, cust_name, cust_email, description, payment_link_token
		FROM invoices
		WHERE merchant_id = ? and invoice_number = ?
	`

	err := r.db.QueryRow(query, merchantID, code).Scan(&data.Data.InvoiceNumber, &data.Data.Amount, &data.Data.Status, &data.Data.DueDate, &data.Data.CreatedAt, &data.Data.UpdatedAt,
		&data.Data.CustName, &data.Data.CustEmail, &data.Data.Description, &data.Data.PaymentLinkToken)
	if err == sql.ErrNoRows {
		data.Error = &model.ErrorBody{
			Code:    http.StatusText(http.StatusBadRequest),
			Message: err.Error(),
			Status:  http.StatusBadRequest,
		}
		return data
	}

	if err != nil {
		data.Error = &model.ErrorBody{
			Code:    http.StatusText(http.StatusInternalServerError),
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		}
		return data
	}

	return data
}

func (r *InvoiceRepository) GetInvoiceById(id int) model.InvoiceByCodeResponse {
	var data model.InvoiceByCodeResponse

	query := `
		SELECT invoice_number, amount, status, due_date, created_at, updated_at, cust_name, cust_email, description, payment_link_token
		FROM invoices
		WHERE id = ?
	`

	err := r.db.QueryRow(query, id).Scan(&data.Data.InvoiceNumber, &data.Data.Amount, &data.Data.Status, &data.Data.DueDate, &data.Data.CreatedAt, &data.Data.UpdatedAt,
		&data.Data.CustName, &data.Data.CustEmail, &data.Data.Description, &data.Data.PaymentLinkToken)
	if err == sql.ErrNoRows {
		data.Error = &model.ErrorBody{
			Code:    http.StatusText(http.StatusBadRequest),
			Message: err.Error(),
			Status:  http.StatusBadRequest,
		}
		return data
	}

	if err != nil {
		data.Error = &model.ErrorBody{
			Code:    http.StatusText(http.StatusInternalServerError),
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		}
		return data
	}

	return data
}

func (r *InvoiceRepository) GetInvoiceByToken(token string) model.InvoiceByCodeResponse {
	var data model.InvoiceByCodeResponse

	query := `
		SELECT id, invoice_number, amount, status, due_date, created_at, updated_at, cust_name, cust_email, description, payment_link_token
		FROM invoices
		WHERE payment_link_token = ?
	`

	err := r.db.QueryRow(query, token).Scan(&data.Data.ID, &data.Data.InvoiceNumber, &data.Data.Amount, &data.Data.Status, &data.Data.DueDate, &data.Data.CreatedAt, &data.Data.UpdatedAt,
		&data.Data.CustName, &data.Data.CustEmail, &data.Data.Description, &data.Data.PaymentLinkToken)
	if err == sql.ErrNoRows {
		data.Error = &model.ErrorBody{
			Code:    http.StatusText(http.StatusBadRequest),
			Message: err.Error(),
			Status:  http.StatusBadRequest,
		}
		return data
	}

	if err != nil {
		data.Error = &model.ErrorBody{
			Code:    http.StatusText(http.StatusInternalServerError),
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		}
		return data
	}

	return data
}

func (r *InvoiceRepository) UpdateInvoice(req model.UpdateInvoiceReq) error {
	query := `
		UPDATE invoices
		SET status = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query, req.Status, req.UpdatedAt, req.ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *InvoiceRepository) GetInvoiceStatistic(req model.InvoiceStatisticReq) model.InvoiceStatisticResponse {
	var res model.InvoiceStatisticResponse

	query := `
		SELECT
			COUNT(i.id) AS count_all,
			SUM(CASE WHEN i.status = 'PAID' THEN 1 ELSE 0 END) AS count_paid,
			SUM(CASE WHEN i.status = 'PENDING' THEN 1 ELSE 0 END) AS count_pending,
			SUM(CASE WHEN i.status = 'EXPIRED' OR (i.status = 'PAID' AND i.due_date < NOW()) THEN 1 ELSE 0 END) AS count_expired,
			COALESCE(SUM(CASE WHEN i.status = 'PAID' THEN i.amount ELSE 0 END), 0) AS sum_transaction,
			COALESCE(SUM(r.total_refund), 0) AS sum_refund
		FROM invoices i
		LEFT JOIN (
			SELECT invoice_id, SUM(amount) AS total_refund
			FROM refunds
			WHERE status = 'SUCCESS'
			GROUP BY invoice_id
		) r ON r.invoice_id = i.id
		WHERE 1=1
	`

	args := []interface{}{}
	if req.MerchantID != nil && *req.MerchantID > 0 {
		query += " AND merchant_id = ?"
		args = append(args, req.MerchantID)
	}

	if req.StartCreatedDate != "" {
		query += " AND i.created_at >= ?"
		args = append(args, req.StartCreatedDate+" 00:00:00")
	}

	if req.EndCreatedDate != "" {
		query += " AND i.created_at <= ?"
		args = append(args, req.EndCreatedDate+" 23:59:59")
	}

	err := r.db.QueryRow(query, args...).Scan(&res.Data.CountAll, &res.Data.CountPaid, &res.Data.CountPending, &res.Data.CountExpired,
		&res.Data.SumTransaction, &res.Data.SumRefund)

	if err != nil {
		res.Error = &model.ErrorBody{
			Code:    http.StatusText(http.StatusInternalServerError),
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		}
		return res
	}

	return res
}
