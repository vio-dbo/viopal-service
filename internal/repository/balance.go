package repository

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"viopal-service/internal/model"
)

type BalanceRepository struct {
	db *sql.DB
}

func NewBalanceRepository(db *sql.DB) *BalanceRepository {
	return &BalanceRepository{db: db}
}

func (r *BalanceRepository) BeginTxBalance(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *BalanceRepository) GetMerchantBalance(id int) model.BalanceResponse {
	var data model.BalanceResponse

	query := `SELECT m.balance FROM merchant m WHERE m.user_id  = ? LIMIT 1`

	err := r.db.QueryRow(query, id).Scan(&data.Balance)
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

func (r *BalanceRepository) GetTopUpByRefNumber(refNumber string) model.TopUpResponse {
	var data model.TopUpResponse

	query := `
		SELECT tu.id, tu.ref_number, CAST(tu.amount AS UNSIGNED), tu.status, tu.request_at, m.bussiness_name, m.id
		FROM top_up tu 
		LEFT JOIN merchant m on m.id = tu.merchant_id
		WHERE tu.ref_number = ? LIMIT 1
	`

	err := r.db.QueryRow(query, refNumber).Scan(&data.Data.ID, &data.Data.RefNumber, &data.Data.Amount, &data.Data.Status, &data.Data.RequestAt, &data.Data.Merchant, &data.Data.MerchantID)
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

func (r *BalanceRepository) UpdateTopUp(ctx context.Context, tx *sql.Tx, req model.UpdateTopUpRequest) error {
	query := `
		UPDATE top_up
		SET status = ?, processed_by = ?, processed_at = ?
		WHERE ref_number = ?
	`

	_, err := r.db.Exec(query, req.Status, req.ProcessedBy, req.ProcessedAt, req.RefNumber)
	if err != nil {
		return err
	}

	return nil
}

func (r *BalanceRepository) UpdateBalance(ctx context.Context, tx *sql.Tx, merchantID int, amount float64) error {
	query := `
		UPDATE merchant
		SET balance = balance + ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query, amount, merchantID)
	if err != nil {
		return err
	}

	return nil
}

func (r *BalanceRepository) UpdateBalanceRefund(ctx context.Context, tx *sql.Tx, merchantID int, amount float64) error {
	query := `
		UPDATE merchant
		SET balance = balance - ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query, amount, merchantID)
	if err != nil {
		return err
	}

	return nil
}

func (r *BalanceRepository) CreateTopUp(ctx context.Context, tx *sql.Tx, req model.TopUp) (int64, error) {
	query := `INSERT INTO top_up (ref_number, merchant_id, amount, status, request_at) VALUES (?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, req.RefNumber, req.MerchantID, req.Amount, req.Status, req.RequestAt)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *BalanceRepository) CreateTopUpLog(ctx context.Context, tx *sql.Tx, req model.TopUpLog) error {
	query := `
		INSERT INTO top_up_logs (
			top_up_id,
			event,
			actor_type,
			actor_id,
			old_status,
			new_status,
			balance_before,
			balance_after,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		req.TopUpID,
		req.Event,
		req.ActorType,
		req.ActorID,
		req.OldStatus,
		req.NewStatus,
		req.BalanceBefore,
		req.BalanceAfter,
		req.CreatedAt,
	)

	return err
}

func (r *BalanceRepository) ListMerchantBalance(req model.ListTopUpRequest) model.ListTopUpResponse {
	var data model.ListTopUpResponse
	offset := (req.Page - 1) * req.PerPage

	query := `
		SELECT tu.ref_number, CAST(tu.amount AS UNSIGNED), tu.status, tu.request_at, m.bussiness_name
		FROM top_up tu 
		LEFT JOIN merchant m on m.id = tu.merchant_id
		WHERE 1=1
	`

	args := []interface{}{}
	if req.Status != "" {
		query += " AND tu.status = ?"
		args = append(args, req.Status)
	}

	if req.RefNumber != "" {
		query += " AND tu.ref_number LIKE ?"
		args = append(args, "%"+req.RefNumber+"%")
	}

	if req.MerchantID > 0 {
		query += " AND tu.merchant_id = ?"
		args = append(args, req.MerchantID)
	}

	query += " ORDER BY tu.request_at DESC LIMIT ? OFFSET ?"
	args = append(args, req.PerPage, offset)

	log.Print(query)
	log.Print(args)

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

	var res []model.TopUpRes
	for rows.Next() {
		var item model.TopUpRes
		err := rows.Scan(&item.RefNumber, &item.Amount, &item.Status, &item.RequestAt, &item.Merchant)
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
