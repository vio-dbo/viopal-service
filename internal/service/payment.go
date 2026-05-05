package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"viopal-service/internal/constant"
	constants "viopal-service/internal/constant"
	"viopal-service/internal/model"
	"viopal-service/internal/repository"
)

type PaymentService struct {
	authRepo    *repository.AuthRepository
	paymentRepo *repository.PaymentRepository
	invoiceRepo *repository.InvoiceRepository
	balanceRepo *repository.BalanceRepository
}

func NewPaymentService(authRepo *repository.AuthRepository, paymentRepo *repository.PaymentRepository, invoiceRepo *repository.InvoiceRepository, balanceRepo *repository.BalanceRepository) *PaymentService {
	return &PaymentService{
		authRepo:    authRepo,
		paymentRepo: paymentRepo,
		invoiceRepo: invoiceRepo,
		balanceRepo: balanceRepo,
	}
}

func (s *PaymentService) Payment(token string) model.PaymentResponse {
	var res model.PaymentResponse
	data := s.invoiceRepo.GetInvoiceByToken(token)
	if data.Error != nil {
		res.Error = data.Error
		return res
	}

	res.Data = model.PaymentIntent{
		InvoiceNumber: data.Data.InvoiceNumber,
		Status:        data.Data.Status,
		CreatedAt:     data.Data.CreatedAt,
	}
	return res
}

func (s *PaymentService) PaymentIntent(token string, paymentMethodID int) model.PaymentResponse {
	var res model.PaymentResponse
	timeNow := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Get invoice by token
	data := s.invoiceRepo.GetInvoiceByToken(token)
	if data.Error != nil {
		res.Error = data.Error
		return res
	}

	if data.Data.Status == constant.InvoiceStatusPaid {
		return model.PaymentResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: "Invoice already success payed",
				Status:  http.StatusBadRequest,
			},
		}
	}

	if data.Data.Status == constant.InvoiceStatusExpired || timeNow.After(data.Data.DueDate) {
		return model.PaymentResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: "Invoice has expired",
				Status:  http.StatusBadRequest,
			},
		}
	}

	tx, err := s.paymentRepo.BeginTxPayment(ctx)
	if err != nil {
		res.Error = &model.ErrorBody{
			Code:    "DB_ERROR",
			Message: err.Error(),
			Status:  500,
		}
		return res
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	intentNumber := fmt.Sprintf("PI-%d", time.Now().UnixNano())

	// 2. Create Payment Intent
	intent := model.PaymentIntent{
		PaymentIntentNumber: intentNumber,
		InvoiceID:           data.Data.ID,
		PaymentMethodID:     paymentMethodID,
		Status:              constants.PaymentIntentStatusPending,
		ExpiredAt:           timeNow.AddDate(0, 0, 1),
		CreatedAt:           timeNow,
	}
	intentID, err := s.paymentRepo.CreatePaymentIntent(tx, intent)
	if err != nil {
		tx.Rollback()
		res.Error = &model.ErrorBody{
			Code:    "CREATE_INTENT_FAILED",
			Message: err.Error(),
			Status:  500,
		}
		return res
	}

	// 3. Create Payment Log
	meta := map[string]interface{}{
		"invoice_id":        data.Data.InvoiceNumber,
		"payment_method_id": intent.PaymentMethodID,
		"amount":            data.Data.Amount,
	}

	metaBytes, err := json.Marshal(meta)
	if err != nil {
		tx.Rollback()
		res.Error = &model.ErrorBody{
			Code:    "META_ERROR",
			Message: err.Error(),
			Status:  500,
		}
		return res
	}

	log := model.PaymentLog{
		PaymentIntentID: intentID,
		Event:           "CREATE_PAYMENT_INTENT",
		ActorType:       "SYSTEM",
		ActorID:         data.Data.MerchantID,
		OldStatus:       "",
		NewStatus:       constants.PaymentIntentStatusPending,
		MetaData:        metaBytes,
		CreatedAt:       timeNow,
	}

	err = s.paymentRepo.CreatePaymentLog(tx, log)
	if err != nil {
		tx.Rollback()
		res.Error = &model.ErrorBody{
			Code:    "LOG_FAILED",
			Message: err.Error(),
			Status:  500,
		}
		return res
	}

	err = tx.Commit()
	if err != nil {
		res.Error = &model.ErrorBody{
			Code:    "COMMIT_FAILED",
			Message: err.Error(),
			Status:  500,
		}
		return res
	}

	res.Data = model.PaymentIntent{
		PaymentIntentNumber: intentNumber,
		Status:              constants.PaymentIntentStatusPending,
		CreatedAt:           timeNow,
		ExpiredAt:           timeNow.AddDate(0, 0, 1),
		InvoiceNumber:       data.Data.InvoiceNumber,
	}

	return res
}

func (s *PaymentService) PaymentIntentByCode(code string) model.PaymentResponse {
	data := s.paymentRepo.PaymentIntentByCode(code)

	data.Data = model.PaymentIntent{
		PaymentIntentNumber: data.Data.PaymentIntentNumber,
		Status:              data.Data.Status,
		CreatedAt:           data.Data.CreatedAt,
		ExpiredAt:           data.Data.ExpiredAt,
		InvoiceNumber:       data.Data.InvoiceNumber,
		Amount:              data.Data.Amount,
		PaymentMethod:       data.Data.PaymentMethod,
	}
	return data
}

func (s *PaymentService) ListPaymentIntent(req model.ListPaymentIntentRequest) model.ListPaymentIntentResponse {
	data := s.paymentRepo.ListPaymentIntent(req)
	return data
}

func (s *PaymentService) UpdatePaymentIntent(userID int, req model.UpdatePaymentIntentRequest) model.PaymentResponse {
	var res model.PaymentResponse
	timeNow := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	paymentIntentData := s.paymentRepo.PaymentIntentByCode(req.PaymentIntentNumber)
	if paymentIntentData.Error != nil {
		res.Error = paymentIntentData.Error
		return res
	}

	if paymentIntentData.Data.Status == constant.PaymentIntentStatusSuccess {
		return model.PaymentResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: "Payment Intent already success payed",
				Status:  http.StatusBadRequest,
			},
		}
	}

	invoiceData := s.invoiceRepo.GetInvoiceById(int(paymentIntentData.Data.InvoiceID))
	if invoiceData.Error != nil {
		res.Error = invoiceData.Error
		return res
	}

	if invoiceData.Data.Status == constant.InvoiceStatusPaid {
		return model.PaymentResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: "Invoice already success payed",
				Status:  http.StatusBadRequest,
			},
		}
	}

	if invoiceData.Data.Status == constant.InvoiceStatusExpired || timeNow.After(invoiceData.Data.DueDate) {
		return model.PaymentResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: "Invoice has expired",
				Status:  http.StatusBadRequest,
			},
		}
	}

	merchant, err := s.authRepo.GetMerchantById(paymentIntentData.Data.MerchantID)
	if err != nil {
		return model.PaymentResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: err.Error(),
				Status:  http.StatusBadRequest,
			},
		}
	}

	tx, err := s.paymentRepo.BeginTxPayment(ctx)
	if err != nil {
		return model.PaymentResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	req.ApprovedAt = &timeNow
	req.ApprovedBy = &userID

	err = s.paymentRepo.UpdatePaymentIntent(tx, req)
	if err != nil {
		return model.PaymentResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	// log
	meta := map[string]interface{}{
		"invoice_id":        paymentIntentData.Data.InvoiceNumber,
		"payment_method_id": paymentIntentData.Data.PaymentMethodID,
		"amount":            paymentIntentData.Data.Amount,
	}

	metaBytes, err := json.Marshal(meta)
	if err != nil {
		tx.Rollback()
		return model.PaymentResponse{
			Error: &model.ErrorBody{
				Code:    "META_ERROR",
				Message: err.Error(),
				Status:  500,
			},
		}
	}

	log := model.PaymentLog{
		PaymentIntentID: paymentIntentData.Data.ID,
		Event:           "VALIDATE_PAYMENT_INTENT",
		ActorType:       "ADMIN",
		ActorID:         int64(userID),
		OldStatus:       paymentIntentData.Data.Status,
		NewStatus:       req.Status,
		MetaData:        metaBytes,
		CreatedAt:       timeNow,
	}

	err = s.paymentRepo.CreatePaymentLog(tx, log)
	if err != nil {
		tx.Rollback()
		return model.PaymentResponse{
			Error: &model.ErrorBody{
				Code:    "LOG_FAILED",
				Message: err.Error(),
				Status:  500,
			},
		}
	}

	if req.Status == constants.PaymentIntentStatusSuccess {
		invoice := model.UpdateInvoiceReq{
			ID:        paymentIntentData.Data.InvoiceID,
			Status:    constants.InvoiceStatusPaid,
			UpdatedAt: timeNow,
		}
		err = s.invoiceRepo.UpdateInvoice(invoice)
		if err != nil {
			return model.PaymentResponse{
				Error: &model.ErrorBody{
					Code:    http.StatusText(http.StatusInternalServerError),
					Message: err.Error(),
					Status:  http.StatusInternalServerError,
				},
			}
		}

		err = s.balanceRepo.UpdateBalance(ctx, tx, merchant.ID, paymentIntentData.Data.Amount)
		if err != nil {
			return model.PaymentResponse{
				Error: &model.ErrorBody{
					Code:    http.StatusText(http.StatusInternalServerError),
					Message: err.Error(),
					Status:  http.StatusInternalServerError,
				},
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return model.PaymentResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	res.Data = model.PaymentIntent{
		PaymentIntentNumber: paymentIntentData.Data.PaymentIntentNumber,
		Status:              req.Status,
		ApprovedAt:          &timeNow,
		InvoiceNumber:       paymentIntentData.Data.InvoiceNumber,
	}

	return res
}
