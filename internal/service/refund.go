package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
	"viopal-service/internal/constant"
	"viopal-service/internal/model"
	"viopal-service/internal/repository"
)

type RefundService struct {
	authRepo    *repository.AuthRepository
	refundRepo  *repository.RefundRepository
	paymentRepo *repository.PaymentRepository
	balanceRepo *repository.BalanceRepository
}

func NewRefundService(authRepo *repository.AuthRepository, refundRepo *repository.RefundRepository, paymentRepo *repository.PaymentRepository, balanceRepo *repository.BalanceRepository) *RefundService {
	return &RefundService{
		authRepo:    authRepo,
		refundRepo:  refundRepo,
		paymentRepo: paymentRepo,
		balanceRepo: balanceRepo,
	}
}

func (s *RefundService) RequestRefund(paymentIntentID string, reason string, userID int) model.RefundResponse {
	var res model.RefundResponse
	timeNow := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	merchant, err := s.authRepo.GetMerchant(userID)
	if err != nil {
		return model.RefundResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: err.Error(),
				Status:  http.StatusBadRequest,
			},
		}
	}

	paymentIntent := s.paymentRepo.PaymentIntentByCode(paymentIntentID)
	if paymentIntent.Error != nil {
		res.Error = paymentIntent.Error
		return res
	}

	if paymentIntent.Data.Status != constant.PaymentIntentStatusSuccess {
		res.Error = &model.ErrorBody{
			Code:    http.StatusText(http.StatusBadRequest),
			Message: "Refund not allowed. Payment not successful",
			Status:  http.StatusBadRequest,
		}
	}

	tx, err := s.refundRepo.BeginTxRefund(ctx)
	if err != nil {
		res.Error = &model.ErrorBody{
			Code:    "DB_ERROR",
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		}
		return res
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	refundNumber := fmt.Sprintf("RF-%d", time.Now().UnixNano())
	refund := model.Refund{
		RefundNumber:    refundNumber,
		Reason:          reason,
		PaymentIntentID: paymentIntent.Data.ID,
		InvoiceID:       paymentIntent.Data.InvoiceID,
		MerchantID:      merchant.ID,
		Amount:          paymentIntent.Data.Amount,
		Status:          constant.RefundStatusRequested,
		RequestBy:       merchant.ID,
		CreatedAt:       &timeNow,
	}
	refundID, err := s.refundRepo.CreateRefund(tx, refund)
	if err != nil {
		tx.Rollback()
		res.Error = &model.ErrorBody{
			Code:    "META_ERROR",
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		}
		return res
	}

	// 5. Create log
	metaLog := model.RefundLog{
		RefundID:  refundID,
		Event:     "REQUEST_REFUND",
		ActorType: "USER",
		ActorID:   userID,
		OldStatus: "",
		NewStatus: constant.RefundStatusRequested,
		CreatedAt: timeNow,
	}

	err = s.refundRepo.CreateRefundLog(tx, metaLog)
	if err != nil {
		tx.Rollback()
		res.Error = &model.ErrorBody{
			Code:    "LOG_FAILED",
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		}
		return res
	}

	// 6. commit
	err = tx.Commit()
	if err != nil {
		res.Error = &model.ErrorBody{
			Code:    "COMMIT_FAILED",
			Message: err.Error(),
			Status:  500,
		}
		return res
	}

	res.Data = model.Refund{
		RefundNumber: refundNumber,
		Reason:       reason,
		Status:       constant.RefundStatusRequested,
	}

	return res
}

func (s *RefundService) ListRefund(userId int, role string, req model.ListRefundRequest) model.ListRefundResponse {
	if role == "merchant" {
		merchant, err := s.authRepo.GetMerchant(userId)
		if err != nil {
			return model.ListRefundResponse{
				Error: &model.ErrorBody{
					Code:    http.StatusText(http.StatusBadRequest),
					Message: err.Error(),
					Status:  http.StatusBadRequest,
				},
			}
		}
		req.MerchantID = merchant.ID
	}

	data := s.refundRepo.ListRefund(req)
	return data
}

func (s *RefundService) DecideRefund(userID int, req model.RefundDecideRequest) model.RefundResponse {
	var res model.RefundResponse
	timeNow := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var decideStatuses = map[string]struct{}{
		constant.RefundStatusFailed:  {},
		constant.RefundStatusSuccess: {},
	}

	if _, exists := decideStatuses[req.Status]; exists {
		return model.RefundResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: "Refund not decided yet",
				Status:  http.StatusBadRequest,
			},
		}
	}

	refund := s.refundRepo.GetRefundByNumber(req.RefundNumber)
	if refund.Error != nil {
		res.Error = refund.Error
		return res
	}

	if refund.Data.Status != constant.RefundStatusRequested {
		return model.RefundResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: fmt.Sprintf("Invalid refund status '%s'. Only 'REQUESTED' refunds can be decided.", refund.Data.Status),
				Status:  http.StatusBadRequest,
			},
		}
	}

	tx, err := s.paymentRepo.BeginTxPayment(ctx)
	if err != nil {
		return model.RefundResponse{
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

	req.DecidedAt = timeNow
	req.DecidedBy = userID

	err = s.paymentRepo.UpdateDecideRefundByNumber(tx, req)
	if err != nil {
		return model.RefundResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	metaLog := model.RefundLog{
		RefundID:  refund.Data.ID,
		Event:     "DECIDE_REFUND",
		ActorType: "ADMIN",
		ActorID:   userID,
		OldStatus: refund.Data.Status,
		NewStatus: req.Status,
		CreatedAt: timeNow,
	}

	err = s.refundRepo.CreateRefundLog(tx, metaLog)
	if err != nil {
		tx.Rollback()
		res.Error = &model.ErrorBody{
			Code:    "LOG_FAILED",
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		}
		return res
	}

	err = tx.Commit()
	if err != nil {
		return model.RefundResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	res.Data = model.Refund{
		RefundNumber: refund.Data.RefundNumber,
		Reason:       refund.Data.Reason,
		Status:       req.Status,
		Amount:       refund.Data.Amount,
		DecidedAt:    &timeNow,
	}

	return res
}

func (s *RefundService) ProcessRefund(userID int, req model.RefundProcessRequest) model.RefundResponse {
	var res model.RefundResponse
	timeNow := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Print("iniii 0")
	refund := s.refundRepo.GetRefundByNumber(req.RefundNumber)
	if refund.Error != nil {
		res.Error = refund.Error
		return res
	}

	log.Print("iniii 1")

	if refund.Data.Status == constant.RefundStatusRequested {
		return model.RefundResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: fmt.Sprintf("Current refund status '%s', refund cannot processed", refund.Data.Status),
				Status:  http.StatusBadRequest,
			},
		}
	}

	var decideStatuses = map[string]struct{}{
		constant.RefundStatusSuccess: {},
		constant.RefundStatusFailed:  {},
	}

	if _, exists := decideStatuses[refund.Data.Status]; exists {
		return model.RefundResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: fmt.Sprintf("Current refund status '%s', refund already processed", refund.Data.Status),
				Status:  http.StatusBadRequest,
			},
		}
	}

	log.Print("iniii 2: ", refund.Data.MerchantID)
	merchant, err := s.authRepo.GetMerchantById(refund.Data.MerchantID)
	if err != nil {
		return model.RefundResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: err.Error(),
				Status:  http.StatusBadRequest,
			},
		}
	}
	log.Print("iniii 3")

	tx, err := s.paymentRepo.BeginTxPayment(ctx)
	if err != nil {
		return model.RefundResponse{
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

	req.ProcessedAt = timeNow
	req.ProcessedBy = userID
	req.Status = constant.RefundStatusSuccess
	if refund.Data.Status == constant.RefundStatusRejected {
		req.Status = constant.RefundStatusFailed
	}

	err = s.paymentRepo.UpdateProcessRefundByNumber(tx, req)
	if err != nil {
		return model.RefundResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	metaLog := model.RefundLog{
		RefundID:  refund.Data.ID,
		Event:     "PROCESS_REFUND",
		ActorType: "ADMIN",
		ActorID:   userID,
		OldStatus: refund.Data.Status,
		NewStatus: req.Status,
		CreatedAt: timeNow,
	}

	err = s.refundRepo.CreateRefundLog(tx, metaLog)
	if err != nil {
		tx.Rollback()
		res.Error = &model.ErrorBody{
			Code:    "LOG_FAILED",
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		}
		return res
	}

	if req.Status == constant.RefundStatusSuccess {
		err = s.balanceRepo.UpdateBalanceRefund(ctx, tx, merchant.ID, refund.Data.Amount)
		if err != nil {
			return model.RefundResponse{
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
		return model.RefundResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	res.Data = model.Refund{
		RefundNumber: refund.Data.RefundNumber,
		Reason:       refund.Data.Reason,
		Status:       req.Status,
		Amount:       refund.Data.Amount,
		ProcessedAt:  &timeNow,
	}

	return res
}
