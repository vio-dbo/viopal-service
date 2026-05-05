package repository

import (
	"context"
	"database/sql"
	"net/http"
	"viopal-service/internal/model"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) BeginTxPayment(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *PaymentRepository) CreatePaymentIntent(tx *sql.Tx, req model.PaymentIntent) (int64, error) {
	query := `
		INSERT INTO payment_intent (
			payment_intent_number,
			invoice_id,
			payment_method_id,
			payed_by,
			approved_by,
			approved_at,
			failure_reason,
			expired_at,
			status,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := tx.Exec(
		query,
		req.PaymentIntentNumber,
		req.InvoiceID,
		req.PaymentMethodID,
		req.PayedBy,
		req.ApprovedBy,
		req.ApprovedAt,
		req.FailureReason,
		req.ExpiredAt,
		req.Status,
		req.CreatedAt,
	)

	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (r *PaymentRepository) UpdatePaymentIntent(tx *sql.Tx, req model.UpdatePaymentIntentRequest) error {
	query := `
		UPDATE payment_intent
		SET status = ?, approved_by = ?, approved_at = ?
		WHERE payment_intent_number = ?
	`

	_, err := tx.Exec(query, req.Status, req.ApprovedBy, req.ApprovedAt, req.PaymentIntentNumber)

	if err != nil {
		return err
	}

	return nil
}

func (r *PaymentRepository) CreatePaymentLog(tx *sql.Tx, log model.PaymentLog) error {
	query := `
		INSERT INTO payment_logs (
			payment_intent_id,
			event,
			actor_type,
			actor_id,
			old_status,
			new_status,
			meta_data,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := tx.Exec(
		query,
		log.PaymentIntentID,
		log.Event,
		log.ActorType,
		log.ActorID,
		log.OldStatus,
		log.NewStatus,
		log.MetaData,
		log.CreatedAt,
	)

	return err
}

func (r *PaymentRepository) PaymentIntentByCode(code string) model.PaymentResponse {
	var data model.PaymentResponse

	query := `
		SELECT
			pi.id,
			pi.payment_intent_number,
			i.id,
			i.invoice_number as invoice_number,
			pm.name as payment_method,
			pi.failure_reason,
			pi.expired_at,
			pi.status,
			pi.created_at,
			i.amount,
			i.merchant_id
		FROM payment_intent pi
		LEFT JOIN payment_method pm on pm.id = pi.payment_method_id
		LEFT JOIN invoices i on i.id = pi.invoice_id
		WHERE pi.payment_intent_number = ?
	`

	err := r.db.QueryRow(query, code).Scan(&data.Data.ID, &data.Data.PaymentIntentNumber, &data.Data.InvoiceID, &data.Data.InvoiceNumber, &data.Data.PaymentMethod,
		&data.Data.FailureReason, &data.Data.ExpiredAt, &data.Data.Status,
		&data.Data.CreatedAt, &data.Data.Amount, &data.Data.MerchantID)
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

func (r *PaymentRepository) ListPaymentIntent(req model.ListPaymentIntentRequest) model.ListPaymentIntentResponse {
	var data model.ListPaymentIntentResponse
	offset := (req.Page - 1) * req.PerPage

	query := `
		SELECT pi.payment_intent_number, i.invoice_number, pi.status, pm.name as payment_method, pi.created_at
		FROM payment_intent pi
		LEFT JOIN invoices i ON i.id = pi.invoice_id
		LEFT JOIN payment_method pm ON pi.payment_method_id = pm.id
		WHERE 1=1 
	`

	args := []interface{}{}
	if req.Status != "" {
		query += " AND pi.status = ?"
		args = append(args, req.Status)
	}

	if req.PaymentIntentNumber != "" {
		query += " AND pi.payment_intent_number LIKE ?"
		args = append(args, "%"+req.PaymentIntentNumber+"%")
	}

	if req.PaymentMethodID > 0 {
		query += " AND pi.payment_method_id = ?"
		args = append(args, req.PaymentMethodID)
	}

	query += " ORDER BY pi.created_at DESC LIMIT ? OFFSET ?"
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

	var res []model.PaymentIntent
	for rows.Next() {
		var item model.PaymentIntent
		err := rows.Scan(&item.PaymentIntentNumber, &item.InvoiceNumber, &item.Status, &item.PaymentMethod, &item.CreatedAt)
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
