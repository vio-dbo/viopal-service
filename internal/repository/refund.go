package repository

import (
	"context"
	"database/sql"
	"net/http"
	"viopal-service/internal/model"
)

type RefundRepository struct {
	db *sql.DB
}

func NewRefundRepository(db *sql.DB) *RefundRepository {
	return &RefundRepository{db: db}
}

func (r *RefundRepository) BeginTxRefund(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *RefundRepository) CreateRefund(tx *sql.Tx, ref model.Refund) (int64, error) {
	query := `
		INSERT INTO refunds (
			refund_number,
			reason,
			payment_intent_id,
			invoice_id,
			merchant_id,
			amount,
			status,
			request_by,
			decided_by,
			decided_at,
			processed_by,
			processed_at,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := tx.Exec(
		query,
		ref.RefundNumber,
		ref.Reason,
		ref.PaymentIntentID,
		ref.InvoiceID,
		ref.MerchantID,
		ref.Amount,
		ref.Status,
		ref.RequestBy,
		ref.DecidedBy,
		ref.DecidedAt,
		ref.ProcessedBy,
		ref.ProcessedAt,
		ref.CreatedAt,
	)

	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (r *RefundRepository) CreateRefundLog(tx *sql.Tx, log model.RefundLog) error {
	query := `
		INSERT INTO refund_logs (
			refund_id,
			event,
			actor_type,
			actor_id,
			old_status,
			new_status,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := tx.Exec(
		query,
		log.RefundID,
		log.Event,
		log.ActorType,
		log.ActorID,
		log.OldStatus,
		log.NewStatus,
		log.CreatedAt,
	)

	return err
}

func (r *RefundRepository) ListRefund(req model.ListRefundRequest) model.ListRefundResponse {
	var data model.ListRefundResponse
	offset := (req.Page - 1) * req.PerPage

	query := `
		SELECT r.refund_number, r.reason, r.amount, r.status, pi.payment_intent_number, i.invoice_number, r.created_at, m.bussiness_name
		FROM refunds r
		LEFT JOIN payment_intent pi ON pi.id = r.payment_intent_id
		LEFT JOIN invoices i ON i.id = r.invoice_id
		LEFT JOIN merchant m on m.id = r.merchant_id
		WHERE 1+1
	`

	args := []interface{}{}
	if req.Status != "" {
		query += " AND r.status = ?"
		args = append(args, req.Status)
	}

	if req.MerchantID > 0 {
		query += " AND r.merchant_id = ?"
		args = append(args, req.MerchantID)
	}

	query += " ORDER BY r.created_at DESC LIMIT ? OFFSET ?"
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

	var res []model.Refund
	for rows.Next() {
		var item model.Refund
		err := rows.Scan(&item.RefundNumber, &item.Reason, &item.Amount, &item.Status, &item.PaymentIntentNumber, &item.InvoiceNumber,
			&item.CreatedAt, &item.Merchant)
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

func (r *RefundRepository) GetRefundByNumber(refund_number string) model.RefundResponse {
	var data model.RefundResponse

	query := `SELECT id, refund_number, amount, status, reason, merchant_id FROM refunds WHERE refund_number = ?`

	err := r.db.QueryRow(query, refund_number).Scan(&data.Data.ID, &data.Data.RefundNumber, &data.Data.Amount, &data.Data.Status, &data.Data.Reason, &data.Data.MerchantID)
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

func (r *PaymentRepository) UpdateDecideRefundByNumber(tx *sql.Tx, req model.RefundDecideRequest) error {
	query := `
		UPDATE refunds
		SET status = ?, decided_by = ?, decided_at = ?
		WHERE refund_number = ?
	`

	_, err := tx.Exec(query, req.Status, req.DecidedBy, req.DecidedAt, req.RefundNumber)

	if err != nil {
		return err
	}

	return nil
}

func (r *PaymentRepository) UpdateProcessRefundByNumber(tx *sql.Tx, req model.RefundProcessRequest) error {
	query := `
		UPDATE refunds
		SET status = ?, processed_by = ?, processed_at = ?
		WHERE refund_number = ?
	`

	_, err := tx.Exec(query, req.Status, req.ProcessedBy, req.ProcessedAt, req.RefundNumber)

	if err != nil {
		return err
	}

	return nil
}
