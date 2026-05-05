package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
	constants "viopal-service/internal/constant"
	"viopal-service/internal/model"
	"viopal-service/internal/repository"
)

type InvoiceService struct {
	authRepo    *repository.AuthRepository
	invoiceRepo *repository.InvoiceRepository
}

func NewInvoiceService(authRepo *repository.AuthRepository, invoiceRepo *repository.InvoiceRepository) *InvoiceService {
	return &InvoiceService{
		authRepo:    authRepo,
		invoiceRepo: invoiceRepo,
	}
}

func GeneratePaymentToken() string {
	b := make([]byte, 12)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func (s *InvoiceService) RequestInvoices(userId, dueInDays int, amount int64, custName, custEmail string) model.InvoiceResponse {
	var res model.InvoiceResponse
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	merchant, err := s.authRepo.GetMerchant(userId)
	if err != nil {
		return model.InvoiceResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: err.Error(),
				Status:  http.StatusBadRequest,
			},
		}
	}

	tx, err := s.invoiceRepo.BeginTxInvoice(ctx)
	if err != nil {
		return model.InvoiceResponse{
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

	var req model.Invoice
	timeNow := time.Now()
	req.InvoiceNumber = fmt.Sprintf("INV-%d", time.Now().UnixNano())
	req.Status = constants.InvoiceStatusPending
	req.MerchantID = int64(merchant.ID)
	req.Amount = float64(amount)
	req.DueDate = timeNow.AddDate(0, 0, dueInDays)
	req.PaymentLinkToken = GeneratePaymentToken()
	req.CustName = custName
	req.CustEmail = custEmail
	req.CreatedAt = timeNow
	req.UpdatedAt = timeNow

	// insert
	res = s.invoiceRepo.CreateInvoice(ctx, tx, req)
	if res.Error != nil {
		return res
	}

	err = tx.Commit()
	if err != nil {
		return model.InvoiceResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	res.Data = model.Invoice{
		InvoiceNumber:    req.InvoiceNumber,
		Amount:           req.Amount,
		Status:           req.Status,
		DueDate:          req.DueDate,
		CreatedAt:        req.CreatedAt,
		PaymentLinkToken: req.PaymentLinkToken,
	}
	return res
}

func (s *InvoiceService) ListInvoice(userId int, req model.ListInvoiceRequest) model.ListInvoiceResponse {
	merchant, err := s.authRepo.GetMerchant(userId)
	if err != nil {
		return model.ListInvoiceResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: err.Error(),
				Status:  http.StatusBadRequest,
			},
		}
	}

	data := s.invoiceRepo.ListInvoice(merchant.ID, req)
	return data
}

func (s *InvoiceService) InvoiceByCode(userId int, code string) model.InvoiceByCodeResponse {
	merchant, err := s.authRepo.GetMerchant(userId)
	if err != nil {
		return model.InvoiceByCodeResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: err.Error(),
				Status:  http.StatusBadRequest,
			},
		}
	}

	data := s.invoiceRepo.GetInvoiceByCode(merchant.ID, code)
	return data
}

func (s *InvoiceService) InvoiceStatistic(req model.InvoiceStatisticReq) model.InvoiceStatisticResponse {
	res := s.invoiceRepo.GetInvoiceStatistic(req)
	return res
}
