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

type BalanceService struct {
	authRepo    *repository.AuthRepository
	balanceRepo *repository.BalanceRepository
}

func NewBalanceService(authRepo *repository.AuthRepository, balanceRepo *repository.BalanceRepository) *BalanceService {
	return &BalanceService{
		authRepo:    authRepo,
		balanceRepo: balanceRepo,
	}
}

func (s *BalanceService) MerchantBalance(id int) model.BalanceResponse {
	data := s.balanceRepo.GetMerchantBalance(id)
	return data
}

func (s *BalanceService) RequestTopUp(userId int, amount float64) model.TopUpResponse {
	var res model.TopUpResponse
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	merchant, err := s.authRepo.GetMerchant(userId)
	if err != nil {
		return model.TopUpResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: err.Error(),
				Status:  http.StatusBadRequest,
			},
		}
	}

	tx, err := s.balanceRepo.BeginTxBalance(ctx)
	if err != nil {
		return model.TopUpResponse{
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

	ref := fmt.Sprintf("TOPUP-%d", time.Now().UnixNano())
	timeNow := time.Now()

	topUp := model.TopUp{
		RefNumber:  ref,
		MerchantID: merchant.ID,
		Amount:     amount,
		Status:     "PENDING",
		RequestAt:  timeNow,
	}

	topUpID, err := s.balanceRepo.CreateTopUp(ctx, tx, topUp)
	if err != nil {
		return model.TopUpResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	// log
	log := model.TopUpLog{
		TopUpID:       topUpID,
		Event:         "REQUEST",
		ActorType:     "MERCHANT",
		ActorID:       merchant.ID,
		OldStatus:     "",
		NewStatus:     "PENDING",
		BalanceBefore: 0,
		BalanceAfter:  0,
		CreatedAt:     timeNow,
	}

	err = s.balanceRepo.CreateTopUpLog(ctx, tx, log)
	if err != nil {
		return model.TopUpResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	err = tx.Commit()
	if err != nil {
		return model.TopUpResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	res.Data = model.TopUpRes{
		RefNumber: ref,
		Amount:    amount,
		Status:    "PENDING",
		RequestAt: timeNow,
	}

	return res
}

func (s *BalanceService) ListRequestTopUp(userId int, role string, req model.ListTopUpRequest) model.ListTopUpResponse {
	if role == "merchant" {
		merchant, err := s.authRepo.GetMerchant(userId)
		if err != nil {
			return model.ListTopUpResponse{
				Error: &model.ErrorBody{
					Code:    http.StatusText(http.StatusBadRequest),
					Message: err.Error(),
					Status:  http.StatusBadRequest,
				},
			}
		}
		req.MerchantID = merchant.ID
	}

	data := s.balanceRepo.ListMerchantBalance(req)
	return data
}

func (s *BalanceService) UpdateTopUp(userID int, req model.UpdateTopUpRequest) model.TopUpResponse {
	log.Print("userId: ", userID)
	var res model.TopUpResponse
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	topUpData := s.balanceRepo.GetTopUpByRefNumber(req.RefNumber)
	if topUpData.Error != nil {
		return topUpData
	}

	if topUpData.Data.Status != constant.TopUpStatusPending {
		return model.TopUpResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: "Top Up already processed",
				Status:  http.StatusBadRequest,
			},
		}
	}

	merchant, err := s.authRepo.GetMerchantById(topUpData.Data.MerchantID)
	if err != nil {
		return model.TopUpResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusBadRequest),
				Message: err.Error(),
				Status:  http.StatusBadRequest,
			},
		}
	}

	tx, err := s.balanceRepo.BeginTxBalance(ctx)
	if err != nil {
		return model.TopUpResponse{
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

	timeNow := time.Now()
	req.ProcessedAt = &timeNow
	req.ProcessedBy = &userID

	err = s.balanceRepo.UpdateTopUp(ctx, tx, req)
	if err != nil {
		return model.TopUpResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	// log
	balanceAfter := merchant.Balance
	if req.Status == constant.TopUpStatusSuccess {
		balanceAfter += float64(topUpData.Data.Amount)
	}

	log := model.TopUpLog{
		TopUpID:       topUpData.Data.ID,
		Event:         "REQUEST",
		ActorType:     "ADMIN",
		ActorID:       userID,
		OldStatus:     "PENDING",
		NewStatus:     req.Status,
		BalanceBefore: merchant.Balance,
		BalanceAfter:  balanceAfter,
		CreatedAt:     timeNow,
	}

	err = s.balanceRepo.CreateTopUpLog(ctx, tx, log)
	if err != nil {
		return model.TopUpResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	if req.Status == constant.TopUpStatusSuccess {
		err = s.balanceRepo.UpdateBalance(ctx, tx, merchant.ID, topUpData.Data.Amount)
		if err != nil {
			return model.TopUpResponse{
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
		return model.TopUpResponse{
			Error: &model.ErrorBody{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			},
		}
	}

	res.Data = model.TopUpRes{
		RefNumber: topUpData.Data.RefNumber,
		Amount:    topUpData.Data.Amount,
		Status:    req.Status,
		RequestAt: timeNow,
	}

	return res
}
